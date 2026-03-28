package main

import (
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
			name:      "missing --input-file and --policy flags fails",
			cmd:       "evaluate input",
			golden:    "Error: required flag(s) \"input-file\", \"policy\" not set\n",
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
			wantError: true,
			name:      "missing --policy flag fails",
			cmd:       "evaluate input --input-file testdata/evaluate/trail-input.json",
			golden:    "Error: required flag(s) \"policy\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --input-file flag fails",
			cmd:       "evaluate input --policy testdata/policies/allow-all.rego",
			golden:    "Error: required flag(s) \"input-file\" not set\n",
		},
		{
			name: "JSON output with allow-all policy",
			cmd:  "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/allow-all.rego --output json",
			goldenJson: []jsonCheck{
				{"allow", true},
			},
		},
		{
			name: "show-input includes input in JSON output",
			cmd:  "evaluate input --input-file testdata/evaluate/trail-input.json --policy testdata/policies/allow-all.rego --output json --show-input",
			goldenJson: []jsonCheck{
				{"allow", true},
				{"input.trail.name", "test-trail"},
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

func TestEvaluateInputCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateInputCommandTestSuite))
}
