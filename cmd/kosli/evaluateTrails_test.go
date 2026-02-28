package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type EvaluateTrailsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	trailName             string
	trailName2            string
}

func (suite *EvaluateTrailsCommandTestSuite) SetupTest() {
	suite.flowName = "evaluate-trails"
	suite.trailName = "test-trail-1"
	suite.trailName2 = "test-trail-2"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	BeginTrail(suite.trailName2, suite.flowName, "", suite.T())
}

func (suite *EvaluateTrailsCommandTestSuite) TestEvaluateTrailsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing trail name argument fails",
			cmd:       fmt.Sprintf(`evaluate trails --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: requires at least 1 arg(s), only received 0\n",
		},
		{
			wantError: true,
			name:      "missing --flow flag fails",
			cmd:       fmt.Sprintf(`evaluate trails %s %s`, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --org orgX`, suite.trailName, suite.flowName),
		},
		{
			wantError: true,
			name:      "evaluating a non-existing trail fails",
			cmd:       fmt.Sprintf(`evaluate trails non-existent --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Trail with name 'non-existent' does not exist for Organization '%s' and Flow '%s'\n", global.Org, suite.flowName),
		},
		{
			name:       "evaluating one trail prints JSON with trails array containing one item",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"trails", "length:1"}, {"trails.[0].name", suite.trailName}},
		},
		{
			name:       "evaluating two trails prints JSON with trails array containing two items",
			cmd:        fmt.Sprintf(`evaluate trails %s %s --flow %s %s`, suite.trailName, suite.trailName2, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"trails", "length:2"}, {"trails.[0].name", suite.trailName}, {"trails.[1].name", suite.trailName2}},
		},
		{
			name: "with --policy allow-all exits 0",
			cmd:  fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy deny-all exits 1",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/deny-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			name:       "with --output json and --policy prints allow/violations JSON",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"allow", true}},
		},
		{
			name:   "with --output table and --policy prints ALLOWED text",
			cmd:    fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output table %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden: "Policy evaluation: ALLOWED\n",
		},
		{
			wantError: true,
			name:      "with --output table and deny --policy prints DENIED text with violations",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/deny-all.rego --output table %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Policy evaluation: DENIED\nViolations:\n  - always denied\nError: policy denied: [always denied]\n",
		},
		{
			wantError: true,
			name:      "with --output invalid returns error",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --output invalid %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: invalid --output value \"invalid\": must be one of [table, json]\n",
		},
		{
			name:       "with --policy and --show-input --output json includes input in JSON output",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"allow", true}, {"input.trails.[0].name", suite.trailName}},
		},
		{
			name:       "with --output json but no --policy prints trails JSON",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"trails.[0].name", suite.trailName}},
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestEvaluateTrailsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateTrailsCommandTestSuite))
}
