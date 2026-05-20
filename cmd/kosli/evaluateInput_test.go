package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		{
			name: "--decision on a single-definition annotated check populates inputs_used and evaluated",
			cmd:  "evaluate input --input-file testdata/evaluate/bakery-pass.json --policy testdata/policies/bakery.rego --decision",
			goldenJson: []jsonCheck{
				{"items.[0].checks.[0].evaluated", "180 >= 175 and 180 <= 200"},
				{"items.[0].checks.[0].inputs_used", "not-nil"},
				{"items.[0].checks.[1].evaluated", "32 >= 25 and 32 <= 40"},
			},
		},
		{
			name: "--decision on a non-iterating policy emits versioned JSON with policy metadata, one item, and a check per annotated rule",
			cmd:  "evaluate input --input-file testdata/evaluate/bakery-pass.json --policy testdata/policies/bakery.rego --decision",
			goldenJson: []jsonCheck{
				{"schema_version", "0.1.0"},
				{"result", "allow"},
				{"policy.title", "Bakery batch compliance"},
				{"items.[0].result", "allow"},
				{"items.[0].checks", "length:2"},
			},
		},
		{
			name: "--decision with --no-assert on a denying policy returns deny in JSON and exits 0",
			cmd:  "evaluate input --input-file testdata/evaluate/bakery-fail.json --policy testdata/policies/bakery.rego --decision --no-assert",
			goldenJson: []jsonCheck{
				{"schema_version", "0.1.0"},
				{"result", "deny"},
				{"items.[0].result", "deny"},
			},
		},
		{
			wantError:   true,
			name:        "--decision on a denying policy exits non-zero by default (assert-on-deny)",
			cmd:         "evaluate input --input-file testdata/evaluate/bakery-fail.json --policy testdata/policies/bakery.rego --decision",
			goldenRegex: `policy denied`,
		},
		{
			name: "--decision on an iterating policy produces one item per element with per-item pass/fail",
			cmd:  "evaluate input --input-file testdata/evaluate/bakery-batches-mixed.json --policy testdata/policies/bakery-batches.rego --decision --no-assert",
			goldenJson: []jsonCheck{
				{"result", "deny"},
				{"items", "length:3"},
				{"items.[0].result", "allow"},
				{"items.[1].result", "allow"},
				{"items.[2].result", "deny"},
				{"items.[2].checks.[0].name", "batch_ok"},
				{"items.[2].checks.[0].result", "fail"},
			},
		},
		{
			name: "--decision on a multi-definition check records alternatives_applied with nested attribution",
			cmd:  "evaluate input --input-file testdata/evaluate/scr-trails-mixed.json --policy testdata/policies/scr-shaped.rego --decision --no-assert",
			goldenJson: []jsonCheck{
				{"items.[0].result", "allow"},
				{"items.[0].checks.[0].title", "Commit has independent review (or is exempt)"},
				{"items.[0].checks.[0].alternatives_applied.[0].title", "exempt — service account author"},
				{"items.[0].checks.[0].alternatives_applied.[0].result", "pass"},
				{"items.[0].checks.[0].alternatives_applied.[1].result", "fail"},

				{"items.[1].result", "allow"},
				{"items.[1].checks.[0].alternatives_applied.[1].result", "pass"},
				{"items.[1].checks.[0].alternatives_applied.[1].alternatives_applied.[1].title", "merge commit — branch authors approved"},
				{"items.[1].checks.[0].alternatives_applied.[1].alternatives_applied.[1].result", "pass"},

				{"items.[2].result", "deny"},
				{"items.[2].checks.[0].alternatives_applied.[1].alternatives_applied.[0].result", "fail"},
				{"items.[2].checks.[0].alternatives_applied.[1].alternatives_applied.[1].result", "fail"},
			},
		},
	}
	runTestCmd(suite.T(), tests)
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
