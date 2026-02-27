package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EvaluateTrailCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	trailName             string
}

func (suite *EvaluateTrailCommandTestSuite) SetupTest() {
	suite.flowName = "evaluate-trail"
	suite.trailName = "test-trail-1"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
}

func (suite *EvaluateTrailCommandTestSuite) TestEvaluateTrailCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing trail name argument fails",
			cmd:       fmt.Sprintf(`evaluate trail --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`evaluate trail %s xxx --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "missing --flow flag fails",
			cmd:       fmt.Sprintf(`evaluate trail %s %s`, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --org orgX`, suite.trailName, suite.flowName),
		},
		{
			wantError: true,
			name:      "evaluating a non-existing trail fails",
			cmd:       fmt.Sprintf(`evaluate trail non-existent --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Trail with name 'non-existent' does not exist for Organization '%s' and Flow '%s'\n", global.Org, suite.flowName),
		},
		{
			name:       "evaluating an existing trail prints wrapped JSON with trail key",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"trail.name", suite.trailName}},
		},
		{
			name: "with --policy allow-all exits 0",
			cmd:  fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy deny-all exits 1",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/deny-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy non-existent file fails",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/does-not-exist.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy invalid rego fails",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/invalid.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			name:       "with --policy allow-all --format json prints JSON with allow true",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --format json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"allow", true}},
		},
		{
			wantError:   true,
			name:        "with --policy deny-all --format json prints JSON with allow false and violations",
			cmd:         fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/deny-all.rego --format json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: `(?s)"allow":\s*false.*"violations":\s*\[.*"always denied"`,
		},
		{
			name:   "with --policy allow-all --format text prints allowed text",
			cmd:    fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --format text %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden: "Policy evaluation: ALLOWED\n",
		},
		{
			wantError: true,
			name:      "with --policy deny-all --format text prints denied text with violations",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/deny-all.rego --format text %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Policy evaluation: DENIED\nViolations:\n  - always denied\nError: policy denied: [always denied]\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEvaluateTrailCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateTrailCommandTestSuite))
}
