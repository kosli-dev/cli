package evaluate

import (
	"encoding/json"
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

const iteratingPolicy = `package policy

import rego.v1

default allow := false

allow if {
	every batch in input.batches {
		batch_ok(batch)
	}
}

# METADATA
# title: Batch is OK
batch_ok(b) if { b.temp_c >= 175 }
`

func threeBatchInput() map[string]interface{} {
	return map[string]interface{}{
		"batches": []interface{}{
			map[string]interface{}{"temp_c": 180},
			map[string]interface{}{"temp_c": 200},
			map[string]interface{}{"temp_c": 150},
		},
	}
}

func TestDecide_EveryProducesOneItemPerElement(t *testing.T) {
	decision, err := Decide(iteratingPolicy, threeBatchInput(), nil)
	require.NoError(t, err)
	require.Len(t, decision.Items, 3)
}

func TestDecide_EveryItemResultMatchesElement(t *testing.T) {
	decision, err := Decide(iteratingPolicy, threeBatchInput(), nil)
	require.NoError(t, err)
	require.Equal(t, "allow", decision.Items[0].Result)
	require.Equal(t, "allow", decision.Items[1].Result)
	require.Equal(t, "deny", decision.Items[2].Result)
}

func TestDecide_EveryTopLevelDenyWhenAnyItemFails(t *testing.T) {
	decision, err := Decide(iteratingPolicy, threeBatchInput(), nil)
	require.NoError(t, err)
	require.Equal(t, "deny", decision.Result)
}

func TestDecide_EveryItemContainsAnnotatedCheck(t *testing.T) {
	decision, err := Decide(iteratingPolicy, threeBatchInput(), nil)
	require.NoError(t, err)
	require.Len(t, decision.Items[2].Checks, 1)
	check := decision.Items[2].Checks[0]
	require.Equal(t, "batch_ok", check.Name)
	require.Equal(t, "Batch is OK", check.Title)
	require.Equal(t, "fail", check.Result)
}

func TestDecide_EveryDetectedAfterGuardPredicates(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if {
	is_array(input.batches)
	count(input.batches) > 0
	every batch in input.batches {
		batch_ok(batch)
	}
}

# METADATA
# title: Batch is OK
batch_ok(b) if { b.temp_c >= 175 }
`
	decision, err := Decide(policy, threeBatchInput(), nil)
	require.NoError(t, err)
	require.Len(t, decision.Items, 3)
	require.Equal(t, "deny", decision.Result)
}

func TestDecide_EveryUnannotatedCheckProducesItemsWithEmptyChecks(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if {
	every batch in input.batches {
		batch_ok(batch)
	}
}

batch_ok(b) if { b.temp_c >= 175 }
`
	decision, err := Decide(policy, threeBatchInput(), nil)
	require.NoError(t, err)
	require.Len(t, decision.Items, 3)
	require.Empty(t, decision.Items[0].Checks)
	require.Equal(t, "allow", decision.Items[0].Result)
	require.Equal(t, "deny", decision.Items[2].Result)
}

func TestDecide_EmptyChecksMarshalAsArrayNotNull(t *testing.T) {
	policy := `package policy

allow = true
`
	decision, err := Decide(policy, map[string]interface{}{}, nil)
	require.NoError(t, err)
	raw, err := json.Marshal(decision)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"checks":[]`)
	require.NotContains(t, string(raw), `"checks":null`)
}
