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

violations contains msg if {
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

violations contains msg if {
	msg := "no allow rule"
}
`
	input := map[string]interface{}{}

	_, err := Evaluate(policy, input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "allow")
}

func TestEvaluate_NoViolationsRule(t *testing.T) {
	policy := `package policy

allow = false
`
	input := map[string]interface{}{}

	result, err := Evaluate(policy, input)
	require.NoError(t, err)
	require.False(t, result.Allow)
	require.Empty(t, result.Violations)
}

func TestEvaluate_NonBooleanAllow(t *testing.T) {
	policy := `package policy

allow = "yes"
`
	input := map[string]interface{}{}

	_, err := Evaluate(policy, input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "boolean")
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

func TestEvaluate_ParamsProvided(t *testing.T) {
	policy := `package policy

import rego.v1

default threshold := 10

threshold := data.params.threshold if { data.params.threshold }

allow if { input.score >= threshold }

violations contains msg if {
	input.score < threshold
	msg := sprintf("score %d is below threshold %d", [input.score, threshold])
}
`
	input := map[string]interface{}{
		"score": 5,
	}
	params := map[string]interface{}{
		"threshold": 3,
	}

	result, err := Evaluate(policy, input, params)
	require.NoError(t, err)
	require.True(t, result.Allow, "score 5 should pass threshold 3")
}
