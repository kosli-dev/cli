package evaluate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecide_SchemaVersion(t *testing.T) {
	policy := `package policy

allow = true
`
	decision, err := Decide(policy, map[string]interface{}{}, nil)
	require.NoError(t, err)
	require.Equal(t, "0.1.0", decision.SchemaVersion)
}

func TestDecide_AllowTrue(t *testing.T) {
	policy := `package policy

allow = true
`
	decision, err := Decide(policy, map[string]interface{}{}, nil)
	require.NoError(t, err)
	require.Equal(t, "allow", decision.Result)
}

func TestDecide_AllowFalse(t *testing.T) {
	policy := `package policy

allow = false
`
	decision, err := Decide(policy, map[string]interface{}{}, nil)
	require.NoError(t, err)
	require.Equal(t, "deny", decision.Result)
}

func TestDecide_PolicyTitle(t *testing.T) {
	policy := `# METADATA
# title: Bakery batch compliance
# description: A batch is compliant if it has the right temperature.
package policy

allow = true
`
	decision, err := Decide(policy, map[string]interface{}{}, nil)
	require.NoError(t, err)
	require.Equal(t, "Bakery batch compliance", decision.Policy.Title)
	require.Equal(t, "A batch is compliant if it has the right temperature.", decision.Policy.Description)
}

func TestDecide_NoPolicyAnnotation(t *testing.T) {
	policy := `package policy

allow = true
`
	decision, err := Decide(policy, map[string]interface{}{}, nil)
	require.NoError(t, err)
	require.Empty(t, decision.Policy.Title)
	require.Empty(t, decision.Policy.Description)
}

func TestDecide_OneAnnotatedCheckPasses(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if {
	temp_ok
}

# METADATA
# title: Temperature in range
temp_ok if {
	input.temp >= 175
}
`
	decision, err := Decide(policy, map[string]interface{}{"temp": 180}, nil)
	require.NoError(t, err)
	require.Equal(t, "allow", decision.Result)
	require.Len(t, decision.Items, 1)
	require.Len(t, decision.Items[0].Checks, 1)
	check := decision.Items[0].Checks[0]
	require.Equal(t, "temp_ok", check.Name)
	require.Equal(t, "Temperature in range", check.Title)
	require.Equal(t, "pass", check.Result)
}

func TestDecide_OneAnnotatedCheckFails(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if {
	temp_ok
}

# METADATA
# title: Temperature in range
temp_ok if {
	input.temp >= 175
}
`
	decision, err := Decide(policy, map[string]interface{}{"temp": 100}, nil)
	require.NoError(t, err)
	require.Equal(t, "deny", decision.Result)
	require.Len(t, decision.Items, 1)
	require.Len(t, decision.Items[0].Checks, 1)
	check := decision.Items[0].Checks[0]
	require.Equal(t, "temp_ok", check.Name)
	require.Equal(t, "fail", check.Result)
}

func TestDecide_MultipleAnnotatedChecks(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if {
	temp_ok
	time_ok
}

# METADATA
# title: Temperature in range
temp_ok if { input.temp >= 175 }

# METADATA
# title: Time in range
time_ok if { input.minutes >= 25 }
`
	decision, err := Decide(policy, map[string]interface{}{"temp": 180, "minutes": 30}, nil)
	require.NoError(t, err)
	require.Equal(t, "allow", decision.Result)
	require.Len(t, decision.Items, 1)
	require.Len(t, decision.Items[0].Checks, 2)
	names := []string{decision.Items[0].Checks[0].Name, decision.Items[0].Checks[1].Name}
	require.ElementsMatch(t, []string{"temp_ok", "time_ok"}, names)
}

func TestDecide_UnannotatedRulesAreIgnored(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if {
	temp_ok
}

# METADATA
# title: Temperature in range
temp_ok if { input.temp >= min_temp }

min_temp := 175
`
	decision, err := Decide(policy, map[string]interface{}{"temp": 180}, nil)
	require.NoError(t, err)
	require.Len(t, decision.Items, 1)
	require.Len(t, decision.Items[0].Checks, 1)
	require.Equal(t, "temp_ok", decision.Items[0].Checks[0].Name)
}

func TestDecide_SingleItemAlways(t *testing.T) {
	policy := `package policy

allow = true
`
	decision, err := Decide(policy, map[string]interface{}{}, nil)
	require.NoError(t, err)
	require.Len(t, decision.Items, 1)
}
