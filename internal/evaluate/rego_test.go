package evaluate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluate_AllowAllPolicy(t *testing.T) {
	policy := `package policy

allow = true
`
	input := map[string]interface{}{
		"trail": map[string]interface{}{
			"name": "test-trail",
		},
	}

	result, err := Evaluate(policy, input)
	require.NoError(t, err)
	require.True(t, result.Allow)
	require.Empty(t, result.Violations)
}

func TestEvaluate_DenyAllPolicy(t *testing.T) {
	policy := `package policy

allow = false

violations[msg] {
	msg := "always denied"
}
`
	input := map[string]interface{}{
		"trail": map[string]interface{}{
			"name": "test-trail",
		},
	}

	result, err := Evaluate(policy, input)
	require.NoError(t, err)
	require.False(t, result.Allow)
	require.Contains(t, result.Violations, "always denied")
}

func TestEvaluate_MissingPackagePolicy(t *testing.T) {
	policy := `package wrong

allow = true
`
	input := map[string]interface{}{}

	_, err := Evaluate(policy, input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "package policy")
}

func TestEvaluate_MissingAllowRule(t *testing.T) {
	policy := `package policy

violations[msg] {
	msg := "no allow rule"
}
`
	input := map[string]interface{}{}

	_, err := Evaluate(policy, input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "allow")
}

func TestEvaluate_SyntaxError(t *testing.T) {
	policy := `package policy

allow = {{{
`
	input := map[string]interface{}{}

	_, err := Evaluate(policy, input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse")
}
