package evaluate

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/open-policy-agent/opa/v1/topdown"
)

// DecisionSchemaVersion is the version of the Decision JSON shape produced
// by Decide. Bump when the shape changes in a non-additive way.
const DecisionSchemaVersion = "0.1.0"

const (
	resultAllow = "allow"
	resultDeny  = "deny"
)

// Decision is the structured outcome of an explainable policy evaluation.
// Slice 1 carries only the skeleton: top-level result, package-level policy
// metadata, and a single item containing per-check pass/fail.
type Decision struct {
	SchemaVersion string     `json:"schema_version"`
	Result        string     `json:"result"`
	Policy        PolicyMeta `json:"policy"`
	Items         []Item     `json:"items"`
}

type PolicyMeta struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type Item struct {
	ID     string  `json:"id,omitempty"`
	Result string  `json:"result"`
	Checks []Check `json:"checks"`
}

type Check struct {
	Name                string         `json:"name"`
	Title               string         `json:"title,omitempty"`
	Result              string         `json:"result"`
	AlternativesApplied []*Alternative `json:"alternatives_applied,omitempty"`
}

// Alternative describes one definition of a multi-definition rule and
// whether it fired during evaluation. When that definition itself
// invoked another multi-definition rule, the chain nests under
// AlternativesApplied.
type Alternative struct {
	Rule                string         `json:"rule"`
	Title               string         `json:"title,omitempty"`
	Result              string         `json:"result"`
	AlternativesApplied []*Alternative `json:"alternatives_applied,omitempty"`
}

// Decide evaluates the given Rego policy against the input and produces a
// structured Decision describing the outcome. The policy must use
// `package policy` and declare an `allow` rule, same as Evaluate.
func Decide(policySource string, input interface{}, params map[string]interface{}) (*Decision, error) {
	res, err := Evaluate(policySource, input, params)
	if err != nil {
		return nil, err
	}
	decision := &Decision{
		SchemaVersion: DecisionSchemaVersion,
		Result:        resultDeny,
	}
	if res.Allow {
		decision.Result = resultAllow
	}

	parsed, err := parseWithAnnotations(policySource)
	if err != nil {
		return nil, err
	}

	decision.Policy = packageMeta(parsed)

	if iter, ok := detectIteration(parsed); ok {
		items, err := evaluateIteration(parsed, policySource, input, params, iter)
		if err != nil {
			return nil, err
		}
		decision.Items = items
		return decision, nil
	}

	module, err := compileWithAnnotations(policySource)
	if err != nil {
		return nil, err
	}
	checks, err := collectChecks(module, policySource, input, params)
	if err != nil {
		return nil, err
	}
	if checks == nil {
		checks = []Check{}
	}
	decision.Items = []Item{{Result: decision.Result, Checks: checks}}

	return decision, nil
}

// parseWithAnnotations parses the policy with annotation processing
// enabled, without invoking the compiler. The parsed AST preserves
// `every x in <ref> { ... }` structure that the compiler would otherwise
// lower into a synthesised local variable.
func parseWithAnnotations(policySource string) (*ast.Module, error) {
	m, err := ast.ParseModuleWithOpts("policy.rego", policySource, ast.ParserOptions{
		ProcessAnnotation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}
	return m, nil
}

// iteration describes an `every x in <domain> { <check>(x) }` pattern
// detected in `allow`'s body.
type iteration struct {
	domain    ast.Ref
	checkName string
}

// detectIteration looks for `every x in input.<path> { <rule>(x) }` in
// any non-default `allow` body and returns its structure. The detection
// requires exactly one expression inside the `every` body, calling a
// same-package rule with the iteration variable.
func detectIteration(parsed *ast.Module) (iteration, bool) {
	for _, rule := range parsed.Rules {
		if rule.Default || rule.Head.Name.String() != "allow" {
			continue
		}
		for _, expr := range rule.Body {
			every, ok := expr.Terms.(*ast.Every)
			if !ok {
				continue
			}
			domainRef, ok := every.Domain.Value.(ast.Ref)
			if !ok || !isInputRef(domainRef) {
				continue
			}
			checkName, ok := singleCallWithVarArg(every.Body, every.Value)
			if !ok {
				continue
			}
			return iteration{domain: domainRef, checkName: checkName}, true
		}
	}
	return iteration{}, false
}

func isInputRef(ref ast.Ref) bool {
	if len(ref) == 0 {
		return false
	}
	v, ok := ref[0].Value.(ast.Var)
	return ok && v.String() == "input"
}

// singleCallWithVarArg returns the function-rule name when body has
// exactly one expression of the form `<name>(<iterVar>)`.
func singleCallWithVarArg(body ast.Body, iterVarTerm *ast.Term) (string, bool) {
	if len(body) != 1 {
		return "", false
	}
	terms, ok := body[0].Terms.([]*ast.Term)
	if !ok || len(terms) != 2 {
		return "", false
	}
	callRef, ok := terms[0].Value.(ast.Ref)
	if !ok || len(callRef) == 0 {
		return "", false
	}
	nameVar, ok := callRef[0].Value.(ast.Var)
	if !ok {
		return "", false
	}
	argVar, ok := terms[1].Value.(ast.Var)
	if !ok {
		return "", false
	}
	iterVar, ok := iterVarTerm.Value.(ast.Var)
	if !ok || argVar != iterVar {
		return "", false
	}
	return nameVar.String(), true
}

// evaluateIteration resolves the iteration domain against the input,
// then evaluates the per-item check rule for each element and returns
// one Item per element.
func evaluateIteration(parsed *ast.Module, policySource string, input interface{}, params map[string]interface{}, iter iteration) ([]Item, error) {
	elements, err := resolveInputArray(input, iter.domain)
	if err != nil {
		return nil, err
	}

	title := ruleTitle(parsed, iter.checkName)
	annotated := ruleIsAnnotated(parsed, iter.checkName)

	items := make([]Item, 0, len(elements))
	for _, elem := range elements {
		pass, err := evaluateRuleCall(policySource, iter.checkName, elem, params)
		if err != nil {
			return nil, err
		}
		item := Item{
			Result: resultDeny,
			Checks: []Check{},
		}
		if pass {
			item.Result = resultAllow
		}
		if annotated {
			alts, err := collectAlternatives(parsed, policySource, iter.checkName, elem, params, true)
			if err != nil {
				return nil, err
			}
			item.Checks = []Check{{
				Name:                iter.checkName,
				Title:               title,
				Result:              passOrFail(pass),
				AlternativesApplied: alts,
			}}
		}
		items = append(items, item)
	}
	return items, nil
}

// resolveInputArray walks the input map following an `input.x.y` ref and
// returns the array at that path.
func resolveInputArray(input interface{}, ref ast.Ref) ([]interface{}, error) {
	cur := input
	for _, part := range ref[1:] {
		key, ok := part.Value.(ast.String)
		if !ok {
			return nil, fmt.Errorf("iteration domain %s contains a non-string key segment", ref)
		}
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("iteration domain %s: expected object at intermediate path", ref)
		}
		cur = m[string(key)]
	}
	arr, ok := cur.([]interface{})
	if !ok {
		return nil, fmt.Errorf("iteration domain %s did not resolve to an array", ref)
	}
	return arr, nil
}

// evaluateRuleCall calls a function rule with `element` as its single
// argument, returning whether the call evaluated to true.
func evaluateRuleCall(policySource, name string, element interface{}, params map[string]interface{}) (bool, error) {
	opts := []func(*rego.Rego){
		rego.Query(fmt.Sprintf("data.policy.%s(input)", name)),
		rego.Module("policy.rego", policySource),
		rego.Input(element),
	}
	if params != nil {
		store := inmem.NewFromObject(map[string]interface{}{"params": params})
		opts = append(opts, rego.Store(store))
	}
	rs, err := rego.New(opts...).Eval(context.Background())
	if err != nil {
		return false, fmt.Errorf("evaluating per-item check %q: %w", name, err)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return false, nil
	}
	b, ok := rs[0].Expressions[0].Value.(bool)
	return ok && b, nil
}

func ruleIsAnnotated(parsed *ast.Module, name string) bool {
	for _, r := range parsed.Rules {
		if r.Head.Name.String() == name && len(r.Annotations) > 0 {
			return true
		}
	}
	return false
}

// ruleTitle returns the overall title for a rule by name, preferring a
// `scope: document` annotation when one is attached to any definition.
func ruleTitle(parsed *ast.Module, name string) string {
	var firstAny string
	for _, r := range parsed.Rules {
		if r.Head.Name.String() != name {
			continue
		}
		for _, a := range r.Annotations {
			if a.Scope == "document" {
				return a.Title
			}
		}
		if firstAny == "" && len(r.Annotations) > 0 {
			firstAny = r.Annotations[0].Title
		}
	}
	return firstAny
}

// ruleDefinitions returns the non-default rule definitions sharing the
// given head name, in source order.
func ruleDefinitions(parsed *ast.Module, name string) []*ast.Rule {
	var defs []*ast.Rule
	for _, r := range parsed.Rules {
		if r.Default {
			continue
		}
		if r.Head.Name.String() == name {
			defs = append(defs, r)
		}
	}
	return defs
}

// collectAlternatives evaluates the named rule with a tracer attached
// and reports per-definition pass/fail for multi-definition rules.
// When a successful definition's body calls another multi-definition
// rule, those nested alternatives are attached recursively. Returns
// nil when the rule has zero or one definitions.
func collectAlternatives(parsed *ast.Module, policySource, name string, input interface{}, params map[string]interface{}, asFunctionCall bool) ([]*Alternative, error) {
	defs := ruleDefinitions(parsed, name)
	if len(defs) <= 1 {
		return nil, nil
	}

	tracer := topdown.NewBufferTracer()
	query := "data.policy." + name
	if asFunctionCall {
		query += "(input)"
	}
	opts := []func(*rego.Rego){
		rego.Query(query),
		rego.Module("policy.rego", policySource),
		rego.Input(input),
		rego.QueryTracer(tracer),
	}
	if params != nil {
		store := inmem.NewFromObject(map[string]interface{}{"params": params})
		opts = append(opts, rego.Store(store))
	}
	if _, err := rego.New(opts...).Eval(context.Background()); err != nil {
		return nil, fmt.Errorf("evaluating alternatives for %q: %w", name, err)
	}

	return ruleAlternatives(parsed, *tracer, name, 0), nil
}

// ruleAlternatives builds the per-definition Alternative entries for a
// multi-definition rule, scoped to the given trace parent QueryID. The
// returned entries recurse into other multi-definition rules invoked
// from each definition's body.
func ruleAlternatives(parsed *ast.Module, events []*topdown.Event, name string, parentQID uint64) []*Alternative {
	defs := ruleDefinitions(parsed, name)
	if len(defs) <= 1 {
		return nil
	}

	alts := make([]*Alternative, 0, len(defs))
	for _, def := range defs {
		pass, enterQID := definitionOutcome(events, def, parentQID)
		alt := &Alternative{
			Rule:   name,
			Title:  defTitle(def),
			Result: passOrFail(pass),
		}
		if enterQID != 0 {
			alt.AlternativesApplied = nestedAlternatives(parsed, events, enterQID)
		}
		alts = append(alts, alt)
	}
	return alts
}

// nestedAlternatives finds multi-definition annotated rules invoked
// inside a parent body (matched by ParentID) and builds Alternative
// chains for each of them.
func nestedAlternatives(parsed *ast.Module, events []*topdown.Event, parentQID uint64) []*Alternative {
	seen := map[string]bool{}
	var names []string
	for _, e := range events {
		if e.Op != topdown.EnterOp || e.ParentID != parentQID {
			continue
		}
		r, ok := e.Node.(*ast.Rule)
		if !ok {
			continue
		}
		n := r.Head.Name.String()
		if seen[n] {
			continue
		}
		defs := ruleDefinitions(parsed, n)
		if len(defs) <= 1 || !anyAnnotated(defs) {
			continue
		}
		seen[n] = true
		names = append(names, n)
	}

	var out []*Alternative
	for _, n := range names {
		out = append(out, ruleAlternatives(parsed, events, n, parentQID)...)
	}
	return out
}

// definitionOutcome searches the trace for the rule definition's
// Enter/Exit pair scoped to parentQID and returns whether it fired and
// the Enter event's QueryID (zero when the definition wasn't entered).
func definitionOutcome(events []*topdown.Event, def *ast.Rule, parentQID uint64) (bool, uint64) {
	var enterQID uint64
	pass := false
	for _, e := range events {
		r, ok := e.Node.(*ast.Rule)
		if !ok || !sameRuleDefinition(r, def) || e.ParentID != parentQID {
			continue
		}
		switch e.Op {
		case topdown.EnterOp:
			enterQID = e.QueryID
		case topdown.ExitOp:
			if e.QueryID == enterQID {
				pass = true
			}
		}
	}
	return pass, enterQID
}

func anyAnnotated(rules []*ast.Rule) bool {
	for _, r := range rules {
		if len(r.Annotations) > 0 {
			return true
		}
	}
	return false
}

// sameRuleDefinition matches a traced rule pointer to a parsed rule
// definition. Tracer events come from a re-parse of the source, so
// pointer equality won't hold; matching on head name + source location
// is reliable across parses.
func sameRuleDefinition(a, b *ast.Rule) bool {
	if a.Head.Name.String() != b.Head.Name.String() {
		return false
	}
	if a.Location == nil || b.Location == nil {
		return false
	}
	return a.Location.Row == b.Location.Row
}

// defTitle returns the per-definition title for an Alternative entry.
// A `scope: rule` annotation describes a specific definition; a
// `scope: document` annotation describes the rule as a whole and is
// surfaced separately as the Check title, not as an alternative's.
func defTitle(def *ast.Rule) string {
	for _, a := range def.Annotations {
		if a.Scope == "rule" {
			return a.Title
		}
	}
	return ""
}

// checkTitle returns the rule's overall title — a `scope: document`
// annotation if one is present, otherwise the first available
// annotation (which for single-definition rules is the natural choice).
func checkTitle(rule *ast.Rule) string {
	for _, a := range rule.Annotations {
		if a.Scope == "document" {
			return a.Title
		}
	}
	if len(rule.Annotations) > 0 {
		return rule.Annotations[0].Title
	}
	return ""
}

// compileWithAnnotations parses and compiles the policy with annotation
// processing enabled. Compilation is required so that body references like
// `temp_ok` become full `data.policy.temp_ok` refs that we can walk.
func compileWithAnnotations(policySource string) (*ast.Module, error) {
	parsed, err := ast.ParseModuleWithOpts("policy.rego", policySource, ast.ParserOptions{
		ProcessAnnotation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}
	c := ast.NewCompiler()
	c.Compile(map[string]*ast.Module{"policy.rego": parsed})
	if c.Failed() {
		return nil, fmt.Errorf("failed to compile policy: %w", c.Errors)
	}
	return c.Modules["policy.rego"], nil
}

// packageMeta returns the title and description from the package-scope
// annotation block, if present.
func packageMeta(module *ast.Module) PolicyMeta {
	for _, a := range module.Annotations {
		if a.Scope == "package" {
			return PolicyMeta{Title: a.Title, Description: a.Description}
		}
	}
	return PolicyMeta{}
}

// collectChecks enumerates the annotated rules referenced directly from
// `allow`'s body and evaluates each one to determine pass/fail. Unannotated
// helpers, value rules (min_temp := 175), and the like are intentionally
// excluded from the check list.
func collectChecks(module *ast.Module, policySource string, input interface{}, params map[string]interface{}) ([]Check, error) {
	referenced := referencedRuleNames(module, "allow")

	seen := map[string]bool{}
	var checks []Check
	for _, rule := range module.Rules {
		name := rule.Head.Name.String()
		if !referenced[name] || seen[name] || len(rule.Annotations) == 0 {
			continue
		}
		seen[name] = true

		pass, err := evaluateRule(policySource, name, input, params)
		if err != nil {
			return nil, err
		}
		alts, err := collectAlternatives(module, policySource, name, input, params, false)
		if err != nil {
			return nil, err
		}
		checks = append(checks, Check{
			Name:                name,
			Title:               checkTitle(rule),
			Result:              passOrFail(pass),
			AlternativesApplied: alts,
		})
	}
	return checks, nil
}

// referencedRuleNames walks the body of every definition of `headName` in
// the module and returns the set of local rule names referenced. Slice 1
// only walks one level deep; transitive references will be needed for
// alternatives_applied in slice 3.
func referencedRuleNames(module *ast.Module, headName string) map[string]bool {
	pkgPath := module.Package.Path.String()
	out := map[string]bool{}
	for _, rule := range module.Rules {
		if rule.Head.Name.String() != headName {
			continue
		}
		ast.WalkRefs(rule.Body, func(ref ast.Ref) bool {
			name, ok := localRuleName(ref, pkgPath)
			if ok {
				out[name] = true
			}
			return false
		})
	}
	return out
}

// localRuleName returns the bare rule name when ref is `data.<pkg>.<rule>`,
// matching this module's package path.
func localRuleName(ref ast.Ref, pkgPath string) (string, bool) {
	prefix := pkgPath + "."
	s := ref.String()
	if len(s) <= len(prefix) || s[:len(prefix)] != prefix {
		return "", false
	}
	tail := s[len(prefix):]
	// Only accept a bare identifier — nested refs (foo.bar.baz) aren't rule
	// references at this level.
	for _, r := range tail {
		if r == '.' || r == '[' {
			return "", false
		}
	}
	return tail, true
}

func evaluateRule(policySource, name string, input interface{}, params map[string]interface{}) (bool, error) {
	opts := []func(*rego.Rego){
		rego.Query("data.policy." + name),
		rego.Module("policy.rego", policySource),
		rego.Input(input),
	}
	if params != nil {
		store := inmem.NewFromObject(map[string]interface{}{"params": params})
		opts = append(opts, rego.Store(store))
	}
	rs, err := rego.New(opts...).Eval(context.Background())
	if err != nil {
		return false, fmt.Errorf("evaluating check %q: %w", name, err)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return false, nil
	}
	b, ok := rs[0].Expressions[0].Value.(bool)
	return ok && b, nil
}

func passOrFail(b bool) string {
	if b {
		return "pass"
	}
	return "fail"
}
