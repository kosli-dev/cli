package evaluate

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

// Result holds the outcome of a policy evaluation.
type Result struct {
	Allow      bool
	Violations []string
}

// Evaluate evaluates a Rego policy against the given input.
// The policy must use `package policy` and declare an `allow` rule.
func Evaluate(policySource string, input interface{}) (*Result, error) {
	if err := validatePolicy(policySource); err != nil {
		return nil, err
	}

	ctx := context.Background()

	r := rego.New(
		rego.Query("data.policy.allow"),
		rego.Module("policy.rego", policySource),
		rego.Input(input),
	)

	rs, err := r.Eval(ctx)
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
		violations, err := collectViolations(ctx, policySource, input)
		if err != nil {
			return nil, err
		}
		result.Violations = violations
	}

	return result, nil
}

func validatePolicy(policySource string) error {
	module, err := ast.ParseModuleWithOpts("policy.rego", policySource, ast.ParserOptions{})
	if err != nil {
		return fmt.Errorf("failed to parse policy: %w", err)
	}

	if module.Package.Path.String() != "data.policy" {
		return fmt.Errorf("policy package must be 'package policy', got '%s'",
			module.Package.Path[1:].String())
	}

	hasAllow := false
	for _, rule := range module.Rules {
		if rule.Head.Name.String() == "allow" {
			hasAllow = true
			break
		}
	}
	if !hasAllow {
		return fmt.Errorf("policy must declare an 'allow' rule")
	}

	return nil
}

func collectViolations(ctx context.Context, policySource string, input interface{}) ([]string, error) {
	r := rego.New(
		rego.Query("data.policy.violations"),
		rego.Module("policy.rego", policySource),
		rego.Input(input),
	)

	rs, err := r.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("violations evaluation failed: %w", err)
	}

	var violations []string
	if len(rs) > 0 && len(rs[0].Expressions) > 0 {
		if vs, ok := rs[0].Expressions[0].Value.([]interface{}); ok {
			for _, v := range vs {
				if s, ok := v.(string); ok {
					violations = append(violations, s)
				}
			}
		}
	}

	return violations, nil
}
