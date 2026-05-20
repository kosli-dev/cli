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

const altsPolicy = `package policy

import rego.v1

default allow := false

allow if {
	compliant
}

# METADATA
# title: exempt — bot author
compliant if { input.author == "bot" }

# METADATA
# title: human commit approved
compliant if { count(input.approvers) > 0 }
`

func TestDecide_MultiDefinitionCheckRecordsAlternatives(t *testing.T) {
	input := map[string]interface{}{
		"author":    "alice",
		"approvers": []interface{}{"bob"},
	}
	decision, err := Decide(altsPolicy, input, nil)
	require.NoError(t, err)
	require.Equal(t, "allow", decision.Result)
	require.Len(t, decision.Items[0].Checks, 1)

	check := decision.Items[0].Checks[0]
	require.Equal(t, "pass", check.Result)
	require.Len(t, check.AlternativesApplied, 2)

	require.Equal(t, "compliant", check.AlternativesApplied[0].Rule)
	require.Equal(t, "exempt — bot author", check.AlternativesApplied[0].Title)
	require.Equal(t, "fail", check.AlternativesApplied[0].Result)

	require.Equal(t, "compliant", check.AlternativesApplied[1].Rule)
	require.Equal(t, "human commit approved", check.AlternativesApplied[1].Title)
	require.Equal(t, "pass", check.AlternativesApplied[1].Result)
}

func TestDecide_MultiDefinitionCheckAllAlternativesFail(t *testing.T) {
	input := map[string]interface{}{
		"author":    "alice",
		"approvers": []interface{}{},
	}
	decision, err := Decide(altsPolicy, input, nil)
	require.NoError(t, err)
	require.Equal(t, "deny", decision.Result)

	check := decision.Items[0].Checks[0]
	require.Equal(t, "fail", check.Result)
	require.Len(t, check.AlternativesApplied, 2)
	require.Equal(t, "fail", check.AlternativesApplied[0].Result)
	require.Equal(t, "fail", check.AlternativesApplied[1].Result)
}

func TestDecide_SingleDefinitionCheckHasNoAlternatives(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if { temp_ok }

# METADATA
# title: Temperature in range
temp_ok if { input.temp >= 175 }
`
	decision, err := Decide(policy, map[string]interface{}{"temp": 180}, nil)
	require.NoError(t, err)
	require.Nil(t, decision.Items[0].Checks[0].AlternativesApplied)
}

func TestDecide_DocumentScopeAnnotationProvidesCheckTitle(t *testing.T) {
	// When a multi-def rule carries a `# METADATA scope: document`
	// annotation, that title summarises the rule and is used at the
	// check level; per-definition `scope: rule` annotations remain the
	// titles of each Alternative.
	policy := `package policy

import rego.v1

default allow := false

allow if { compliant }

# METADATA
# scope: document
# title: Commit is compliant (exempt or approved)

# METADATA
# title: exempt — bot
compliant if { input.author == "bot" }

# METADATA
# title: approved
compliant if { count(input.approvers) > 0 }
`
	input := map[string]interface{}{"author": "alice", "approvers": []interface{}{"bob"}}
	decision, err := Decide(policy, input, nil)
	require.NoError(t, err)

	check := decision.Items[0].Checks[0]
	require.Equal(t, "Commit is compliant (exempt or approved)", check.Title)
	require.Equal(t, "exempt — bot", check.AlternativesApplied[0].Title)
	require.Equal(t, "approved", check.AlternativesApplied[1].Title)
}

func TestDecide_NestedAlternativesOnFailingHumanDef(t *testing.T) {
	// All alternatives fail: a human commit on a regular branch with no
	// approvers. The bot def fails (not a bot), the human def is
	// entered but fails because has_independent_approval finds no
	// fitting alternative. Both nested alternatives must still be
	// attributed under the human def for the auditor to see why.
	policy := scrShapedPolicy
	input := map[string]interface{}{
		"trails": []interface{}{
			map[string]interface{}{
				"author":    "alice",
				"kind":      "regular",
				"approvers": []interface{}{},
			},
		},
	}
	decision, err := Decide(policy, input, nil)
	require.NoError(t, err)
	require.Equal(t, "deny", decision.Result)

	check := decision.Items[0].Checks[0]
	require.Equal(t, "fail", check.Result)
	require.Len(t, check.AlternativesApplied, 2)
	require.Equal(t, "fail", check.AlternativesApplied[0].Result) // bot

	human := check.AlternativesApplied[1]
	require.Equal(t, "fail", human.Result)
	require.Len(t, human.AlternativesApplied, 2,
		"nested has_independent_approval alternatives must be attributed even on failure")
	require.Equal(t, "fail", human.AlternativesApplied[0].Result)
	require.Equal(t, "fail", human.AlternativesApplied[1].Result)
}

const scrShapedPolicy = `package policy

import rego.v1

default allow := false

allow if {
	every trail in input.trails {
		trail_compliant(trail)
	}
}

# METADATA
# title: exempt — service account author
trail_compliant(trail) if {
	startswith(trail.author, "bot-")
}

# METADATA
# title: human commit with independent PR approval
trail_compliant(trail) if {
	not startswith(trail.author, "bot-")
	has_independent_approval(trail)
}

# METADATA
# title: regular commit — branch authors + PR author approved
has_independent_approval(trail) if {
	trail.kind == "regular"
	count(trail.approvers) > 0
}

# METADATA
# title: merge commit — branch authors approved
has_independent_approval(trail) if {
	trail.kind == "merge"
}
`

func TestDecide_NestedAlternativesAttribution(t *testing.T) {
	// Mirrors the SCR shape: trail_compliant has two definitions, the
	// second of which calls has_independent_approval (also multi-def).
	// Input describes a human-author merge commit, so the bot
	// alternative fails, the human alternative fires, and within it the
	// regular-commit nested alternative fails while the merge-commit
	// one fires.
	policy := scrShapedPolicy
	input := map[string]interface{}{
		"trails": []interface{}{
			map[string]interface{}{
				"author":    "alice",
				"kind":      "merge",
				"approvers": []interface{}{},
			},
		},
	}
	decision, err := Decide(policy, input, nil)
	require.NoError(t, err)
	require.Equal(t, "allow", decision.Result)

	require.Len(t, decision.Items, 1)
	check := decision.Items[0].Checks[0]
	require.Equal(t, "trail_compliant", check.Name)
	require.Equal(t, "pass", check.Result)

	require.Len(t, check.AlternativesApplied, 2)
	require.Equal(t, "exempt — service account author", check.AlternativesApplied[0].Title)
	require.Equal(t, "fail", check.AlternativesApplied[0].Result)
	require.Empty(t, check.AlternativesApplied[0].AlternativesApplied)

	humanAlt := check.AlternativesApplied[1]
	require.Equal(t, "human commit with independent PR approval", humanAlt.Title)
	require.Equal(t, "pass", humanAlt.Result)
	require.Len(t, humanAlt.AlternativesApplied, 2, "nested has_independent_approval alternatives should be attributed")

	require.Equal(t, "has_independent_approval", humanAlt.AlternativesApplied[0].Rule)
	require.Equal(t, "regular commit — branch authors + PR author approved", humanAlt.AlternativesApplied[0].Title)
	require.Equal(t, "fail", humanAlt.AlternativesApplied[0].Result)

	require.Equal(t, "has_independent_approval", humanAlt.AlternativesApplied[1].Rule)
	require.Equal(t, "merge commit — branch authors approved", humanAlt.AlternativesApplied[1].Title)
	require.Equal(t, "pass", humanAlt.AlternativesApplied[1].Result)
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
