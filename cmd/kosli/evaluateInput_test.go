package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/version"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EvaluateInputCommandTestSuite struct {
	suite.Suite
}

func (suite *EvaluateInputCommandTestSuite) TestEvaluateInputCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing --policy flag fails",
			cmd:       "evaluate input",
			golden:    "Error: required flag(s) \"policy\" not set\n",
		},
		{
			name:        "allow-all policy with input file returns ALLOWED",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/allow-all.rego",
			goldenRegex: `RESULT:\s+ALLOWED`,
		},
		{
			wantError:   true,
			name:        "deny-all policy with input file returns DENIED",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/deny-all.rego",
			goldenRegex: `RESULT:\s+DENIED`,
		},
		{
			wantError:   true,
			name:        "non-existent input file returns error",
			cmd:         "evaluate input --input-file testdata/evaluate/no-such-file.json --policy testdata/policies/allow-all.rego",
			goldenRegex: `failed to read input file:`,
		},
		{
			wantError:   true,
			name:        "invalid JSON input file returns error",
			cmd:         "evaluate input --input-file testdata/policies/allow-all.rego --policy testdata/policies/allow-all.rego",
			goldenRegex: `failed to parse input:`,
		},
		{
			wantError:   true,
			name:        "missing --input-file reads from stdin (empty stdin fails)",
			cmd:         "evaluate input --policy testdata/policies/allow-all.rego",
			goldenRegex: `failed to parse input:`,
		},
		{
			name: "JSON output with allow-all policy",
			cmd:  "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/allow-all.rego --output json",
			goldenJson: []jsonCheck{
				{"allow", true},
			},
		},
		{
			wantError:   true,
			name:        "policy with wrong package returns error",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/no-package-policy.rego",
			goldenRegex: `policy package must be 'package policy', got 'foo'`,
		},
		{
			wantError:   true,
			name:        "policy missing allow rule returns error",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/no-allow-rule.rego",
			goldenRegex: `policy must declare an 'allow' rule`,
		},
		{
			wantError:   true,
			name:        "deny without violations rule returns DENIED with no violation messages",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/deny-no-violations.rego",
			goldenRegex: `RESULT:\s+DENIED`,
		},
		{
			name: "show-input includes input in JSON output",
			cmd:  "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/allow-all.rego --output json --show-input",
			goldenJson: []jsonCheck{
				{"allow", true},
				{"input.trail.name", "test-trail"},
			},
		},
		{
			name:        "inline --params overrides policy default threshold",
			cmd:         `evaluate input --input-file testdata/evaluate/score-input.json --policy testdata/policies/check-params-threshold.rego --params '{"threshold":3}'`,
			goldenRegex: `RESULT:\s+ALLOWED`,
		},
		{
			name:        "--params from file overrides policy default threshold",
			cmd:         "evaluate input --input-file testdata/evaluate/score-input.json --policy testdata/policies/check-params-threshold.rego --params @testdata/evaluate/params-low-threshold.json",
			goldenRegex: `RESULT:\s+ALLOWED`,
		},
		{
			wantError:   true,
			name:        "--params with invalid JSON returns error",
			cmd:         "evaluate input --input-file testdata/evaluate/score-input.json --policy testdata/policies/allow-all.rego --params not-json",
			goldenRegex: `failed to parse --params`,
		},
		{
			name: "show-input with params includes params in JSON output",
			cmd:  `evaluate input --input-file testdata/evaluate/score-input.json --policy testdata/policies/check-params-threshold.rego --params '{"threshold":3}' --output json --show-input`,
			goldenJson: []jsonCheck{
				{"allow", true},
				{"input.score", float64(5)},
				{"params.threshold", float64(3)},
			},
		},
		{
			name:        "deny-all with --no-assert exits 0 and prints DENIED",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/deny-all.rego --no-assert",
			goldenRegex: `RESULT:\s+DENIED`,
		},
		{
			wantError:   true,
			name:        "deny-all with --assert exits non-zero (matches default)",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/deny-all.rego --assert",
			goldenRegex: `RESULT:\s+DENIED`,
		},
		{
			wantError:   true,
			name:        "deny-all with no flag still exits non-zero (default unchanged)",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/deny-all.rego",
			goldenRegex: `RESULT:\s+DENIED`,
		},
		{
			wantError:   true,
			name:        "--assert and --no-assert together are mutually exclusive",
			cmd:         "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/allow-all.rego --assert --no-assert",
			goldenRegex: `none of the others can be.*\[assert no-assert\] were all set`,
		},
		{
			name: "deny-all with --no-assert and --output json prints allow false and exits 0",
			cmd:  "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/deny-all.rego --no-assert --output json",
			goldenJson: []jsonCheck{
				{"allow", false},
			},
		},
	}
	runTestCmd(suite.T(), tests)
}

// TestEvaluateInputCmdOPAContract covers the OPA/Rego boundary as the user
// meets it. The Rego semantics themselves are pinned in
// internal/evaluate/opa_contract_test.go — the cases here exist only to prove
// the plumbing: that a compiler diagnostic reaches the terminal intact, and
// that a real policy's verdict and reasons are rendered. Output formats,
// --params sources and the --assert pair are already covered by
// TestEvaluateInputCmd and are deliberately not re-tested per policy.
func (suite *EvaluateInputCommandTestSuite) TestEvaluateInputCmdOPAContract() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "policy with an unsafe variable reports the rego error with line number",
			cmd:       "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/unsafe-var.rego",
			// Note the reported file is always "policy.rego" — the name we pass
			// to rego.Module — not the user's --policy path. The line number is
			// what makes this actionable, so pin it.
			goldenRegex: `policy\.rego:9: rego_unsafe_var_error`,
		},
		{
			wantError: true,
			name:      "a policy that fails to compile errors out even under --no-assert",
			cmd:       "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/unsafe-var.rego --no-assert",
			// Anchored at the start of the combined output: --no-assert
			// suppresses a deny, not a broken policy, and nothing — least of
			// all a RESULT line — precedes the error.
			goldenRegex: `\AError: policy evaluation failed`,
		},
		{
			name:        "realistic compliance policy allows when default requirements are met",
			cmd:         "evaluate input --input-file testdata/evaluate/compliant-trail-input.json --policy testdata/policies/realistic-compliance.rego",
			goldenRegex: `RESULT:\s+ALLOWED`,
		},
		{
			wantError: true,
			name:      "realistic compliance policy denies and names every violation",
			cmd:       "evaluate input --input-file testdata/evaluate/compliant-trail-input.json --policy testdata/policies/realistic-compliance.rego --params @testdata/evaluate/params-required-attestations.json",
			// Rego sets are unordered, so accept either order — but require
			// both reasons to be rendered.
			goldenRegex: `RESULT:\s+DENIED[\s\S]*(snyk-scan[\s\S]*sbom|sbom[\s\S]*snyk-scan)`,
		},
		{
			wantError:   true,
			name:        "policy whose allow rule is undefined errors rather than denying",
			cmd:         "evaluate input --input-file testdata/evaluate/low-score-input.json --policy testdata/policies/undefined-allow.rego",
			goldenRegex: `policy did not return a result for 'data\.policy\.allow'`,
		},
	}
	runTestCmd(suite.T(), tests)
}

// TestPolicyFixturesStillCompile is the cheap canary for every future OPA
// bump: every policy fixture we ship (bar the deliberately broken ones) must
// still compile under the embedded OPA. It catches Rego-language regressions
// across the whole corpus without a test case per fixture.
func TestPolicyFixturesStillCompile(t *testing.T) {
	// Fixtures that are invalid on purpose and are asserted on elsewhere.
	// undefined-allow.rego is not among them: it compiles fine and is only
	// undefined at evaluation time.
	deliberatelyBroken := map[string]bool{
		"invalid.rego":    true,
		"unsafe-var.rego": true,
	}

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

func TestLoadInput(t *testing.T) {
	reader := strings.NewReader(`{"trail": {"name": "from-reader"}}`)
	input, err := loadInput(reader)
	require.NoError(t, err)
	trail, ok := input["trail"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "from-reader", trail["name"])
}

func TestLoadInputInvalidJSON(t *testing.T) {
	reader := strings.NewReader(`not json`)
	_, err := loadInput(reader)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse input")
}

func TestLoadPolicyFromLocalFile(t *testing.T) {
	body, err := loadPolicy("testdata/policies/allow-all.rego")
	require.NoError(t, err)
	require.Contains(t, string(body), "package policy")
}

func TestLoadPolicyMissingLocalFile(t *testing.T) {
	_, err := loadPolicy("testdata/policies/no-such-file.rego")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read policy file")
}

func TestLoadPolicyFromHTTPS(t *testing.T) {
	const rego = "package policy\n\nallow = true\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/policy.rego", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, rego)
	}))
	defer server.Close()

	body, err := loadPolicy(server.URL + "/policy.rego")
	require.NoError(t, err)
	require.Equal(t, rego, string(body))
}

func TestLoadPolicyRemoteNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := loadPolicy(server.URL + "/missing.rego")
	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP 404")
}

func TestLoadPolicyDoesNotReadNon2xxBody(t *testing.T) {
	var bodyServed bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "huge error page")
		bodyServed = true
	}))
	defer server.Close()

	_, err := loadPolicy(server.URL + "/policy.rego")
	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP 500")
	// We may or may not have raced the handler, but the error must not
	// contain the body text — the status check happens before the read.
	require.NotContains(t, err.Error(), "huge error page")
	_ = bodyServed
}

func TestLoadPolicyRejectsOversizedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Stream just over the 5 MiB cap.
		chunk := make([]byte, 1<<20)
		for i := 0; i < 6; i++ {
			_, _ = w.Write(chunk)
		}
	}))
	defer server.Close()

	_, err := loadPolicy(server.URL + "/policy.rego")
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds")
}

func TestLoadPolicyBlocksCrossHostRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "package policy\nallow = true\n")
	}))
	defer target.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/policy.rego", http.StatusFound)
	}))
	defer redirector.Close()

	_, err := loadPolicy(redirector.URL + "/policy.rego")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cross-host redirect")
}

func TestLoadPolicyHonorsHTTPProxy(t *testing.T) {
	const rego = "package policy\n\nallow = true\n"

	var sawProxyStyleRequest bool
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A request routed through an HTTP proxy carries an absolute URL on
		// the request line, so r.URL.Host will be populated.
		if r.URL.Host != "" {
			sawProxyStyleRequest = true
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, rego)
	}))
	defer proxy.Close()

	prev := global
	global = &GlobalOpts{HttpProxy: proxy.URL}
	t.Cleanup(func() { global = prev })

	body, err := loadPolicy("http://policies.example.invalid/policy.rego")
	require.NoError(t, err)
	require.Equal(t, rego, string(body))
	require.True(t, sawProxyStyleRequest, "expected proxy to receive an absolute-URL request")
}

func TestEvaluateInputCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateInputCommandTestSuite))
}
