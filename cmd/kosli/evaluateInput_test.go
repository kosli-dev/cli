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
