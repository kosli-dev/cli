package evaluate

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
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
	Name   string `json:"name"`
	Title  string `json:"title,omitempty"`
	Result string `json:"result"`
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

	module, err := compileWithAnnotations(policySource)
	if err != nil {
		return nil, err
	}

	decision.Policy = packageMeta(module)

	checks, err := collectChecks(module, policySource, input, params)
	if err != nil {
		return nil, err
	}
	decision.Items = []Item{{Result: decision.Result, Checks: checks}}

	return decision, nil
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
		checks = append(checks, Check{
			Name:   name,
			Title:  rule.Annotations[0].Title,
			Result: passOrFail(pass),
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
