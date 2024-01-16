package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArchiveFlowCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
}

func (suite *ArchiveFlowCommandTestSuite) SetupTest() {
	suite.flowName = "archive-flow"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlow(suite.flowName, suite.T())
}

func (suite *ArchiveFlowCommandTestSuite) TestArchiveFlowCmd() {
	tests := []cmdTestCase{
		{
			name:   "can archive flow",
			cmd:    fmt.Sprintf(`archive flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden: "flow archive-flow was archived\n",
		},
		{
			wantError: true,
			name:      "archiving non-existing flow fails",
			cmd:       fmt.Sprintf(`archive flow non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Flow named 'non-existing' does not exist for organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "archive flow fails when 2 args are provided",
			cmd:       fmt.Sprintf(`archive flow %s arg2 %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "archive flow fails when no args are provided",
			cmd:       fmt.Sprintf(`archive flow %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArchiveFlowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveFlowCommandTestSuite))
}
