package evaluate

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
)

// Result holds the outcome of a policy evaluation.
type Result struct {
	Allow      bool
	Violations []string
	Outputs    map[string]any
}

// Evaluate evaluates a Rego policy against the given input.
// The policy must use `package policy` and declare an `allow` rule.
// An optional params map can be provided to populate data.params in the policy.
// The values of outputRules, rules of the policy package, are returned in Result.Outputs.
func Evaluate(policySource string, input any, params map[string]any, outputRules ...string) (*Result, error) {
	if err := validatePolicy(policySource, outputRules); err != nil {
		return nil, err
	}

	ctx := context.Background()

	rs, err := evalQuery(ctx, "data.policy.allow", policySource, input, params)
	if err != nil {
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return nil, fmt.Errorf("policy did not return a result for 'data.policy.allow'")
	}

	allow, ok := rs[0].Expressions[0].Value.(bool)
	if !ok {
		return nil, fmt.Errorf("policy 'allow' rule must evaluate to boolean, got %T", rs[0].Expressions[0].Value)
	}

	result := &Result{Allow: allow}

	if !result.Allow {
		violations, err := collectViolations(ctx, policySource, input, params)
		if err != nil {
			return nil, err
		}
		result.Violations = violations
	}

	for _, rule := range outputRules {
		value, err := evaluateRule(ctx, policySource, input, params, rule)
		if err != nil {
			return nil, err
		}
		if result.Outputs == nil {
			result.Outputs = map[string]any{}
		}
		result.Outputs[rule] = value
	}

	return result, nil
}

func evaluateRule(ctx context.Context, policySource string, input any, params map[string]any, rule string) (any, error) {
	rs, err := evalQuery(ctx, "data.policy."+rule, policySource, input, params)
	if err != nil {
		return nil, fmt.Errorf("%s evaluation failed: %w", rule, err)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return nil, nil
	}
	return rs[0].Expressions[0].Value, nil
}

func validatePolicy(policySource string, outputRules []string) error {
	module, err := ast.ParseModuleWithOpts("policy.rego", policySource, ast.ParserOptions{})
	if err != nil {
		return fmt.Errorf("failed to parse policy: %w", err)
	}

	if module.Package.Path.String() != "data.policy" {
		return fmt.Errorf("policy package must be 'package policy', got '%s'",
			module.Package.Path[1:].String())
	}

	declared := map[string]bool{}
	for _, rule := range module.Rules {
		declared[rule.Head.Name.String()] = true
	}
	if !declared["allow"] {
		return fmt.Errorf("policy must declare an 'allow' rule")
	}
	for _, rule := range outputRules {
		if !declared[rule] {
			return fmt.Errorf("policy does not declare a '%s' rule", rule)
		}
	}

	return nil
}

func collectViolations(ctx context.Context, policySource string, input any, params map[string]any) ([]string, error) {
	rs, err := evalQuery(ctx, "data.policy.violations", policySource, input, params)
	if err != nil {
		return nil, fmt.Errorf("violations evaluation failed: %w", err)
	}

	var violations []string
	if len(rs) > 0 && len(rs[0].Expressions) > 0 {
		if vs, ok := rs[0].Expressions[0].Value.([]any); ok {
			for _, v := range vs {
				if s, ok := v.(string); ok {
					violations = append(violations, s)
				}
			}
		}
	}

	return violations, nil
}

func evalQuery(ctx context.Context, query string, policySource string, input any, params map[string]any) (rego.ResultSet, error) {
	opts := []func(*rego.Rego){
		rego.Query(query),
		rego.Module("policy.rego", policySource),
		rego.Input(input),
	}
	if params != nil {
		store := inmem.NewFromObject(map[string]any{"params": params})
		opts = append(opts, rego.Store(store))
	}
	return rego.New(opts...).Eval(ctx)
}
