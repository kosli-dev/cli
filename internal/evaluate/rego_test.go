package evaluate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluate_AllowAllPolicy(t *testing.T) {
	policy := `package policy

allow = true
`
	input := map[string]any{
		"trail": map[string]any{
			"name": "test-trail",
		},
	}

	result, err := Evaluate(policy, input, nil)
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
	input := map[string]any{
		"trail": map[string]any{
			"name": "test-trail",
		},
	}

	result, err := Evaluate(policy, input, nil)
	require.NoError(t, err)
	require.False(t, result.Allow)
	require.Contains(t, result.Violations, "always denied")
}

func TestEvaluate_MissingPackagePolicy(t *testing.T) {
	policy := `package wrong

allow = true
`
	input := map[string]any{}

	_, err := Evaluate(policy, input, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "package policy")
}

func TestEvaluate_MissingAllowRule(t *testing.T) {
	policy := `package policy

violations contains msg if {
	msg := "no allow rule"
}
`
	input := map[string]any{}

	_, err := Evaluate(policy, input, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "allow")
}

func TestEvaluate_NoViolationsRule(t *testing.T) {
	policy := `package policy

allow = false
`
	input := map[string]any{}

	result, err := Evaluate(policy, input, nil)
	require.NoError(t, err)
	require.False(t, result.Allow)
	require.Empty(t, result.Violations)
}

func TestEvaluate_NonBooleanAllow(t *testing.T) {
	policy := `package policy

allow = "yes"
`
	input := map[string]any{}

	_, err := Evaluate(policy, input, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "boolean")
}

func TestEvaluate_SyntaxError(t *testing.T) {
	policy := `package policy

allow = {{{
`
	input := map[string]any{}

	_, err := Evaluate(policy, input, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse")
}

func TestEvaluate_ParamsProvided(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

default threshold := 10

threshold := data.params.threshold if { data.params.threshold }

allow if { input.score >= threshold }

violations contains msg if {
	input.score < threshold
	msg := sprintf("score %d is below threshold %d", [input.score, threshold])
}
`
	input := map[string]any{
		"score": 5,
	}
	params := map[string]any{
		"threshold": 3,
	}

	result, err := Evaluate(policy, input, params)
	require.NoError(t, err)
	require.True(t, result.Allow, "score 5 should pass threshold 3")
}

func TestEvaluate_ParamsDefault(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

default threshold := 10

threshold := data.params.threshold if { data.params.threshold }

allow if { input.score >= threshold }

violations contains msg if {
	input.score < threshold
	msg := sprintf("score %d is below threshold %d", [input.score, threshold])
}
`
	input := map[string]any{
		"score": 5,
	}

	// No params — policy uses default threshold of 10
	result, err := Evaluate(policy, input, nil)
	require.NoError(t, err)
	require.False(t, result.Allow, "score 5 should fail default threshold 10")
	require.Len(t, result.Violations, 1)
	require.Contains(t, result.Violations[0], "below threshold 10")
}

func TestEvaluate_ParamsIgnoredByPolicy(t *testing.T) {
	policy := `package policy

allow = true
`
	input := map[string]any{}
	params := map[string]any{
		"unused_key": "unused_value",
	}

	result, err := Evaluate(policy, input, params)
	require.NoError(t, err)
	require.True(t, result.Allow, "params not referenced by policy should have no effect")
}

func TestEvaluate_OutputRules(t *testing.T) {
	for _, allow := range []string{"true", "false"} {
		t.Run("allow = "+allow, func(t *testing.T) {
			policy := `package policy

allow = ` + allow + `

report := {"compliant": ` + allow + `}
`
			result, err := Evaluate(policy, map[string]any{}, nil, "report")
			require.NoError(t, err)
			require.Equal(t, map[string]any{"report": map[string]any{"compliant": allow == "true"}}, result.Outputs)
		})
	}
}
