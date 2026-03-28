package main

import (
	"testing"

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
	}
	runTestCmd(suite.T(), tests)
}

func TestEvaluateInputCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateInputCommandTestSuite))
}
