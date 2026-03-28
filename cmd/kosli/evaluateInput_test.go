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
	}
	runTestCmd(suite.T(), tests)
}

func TestEvaluateInputCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateInputCommandTestSuite))
}
