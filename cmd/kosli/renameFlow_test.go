package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type RenameFlowCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
}

func (suite *RenameFlowCommandTestSuite) SetupTest() {
	suite.flowName = "rename-flow"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlow(suite.flowName, suite.T())
}

func (suite *RenameFlowCommandTestSuite) TestRenameFlowCmd() {
	tests := []cmdTestCase{
		{
			name:   "can rename flow",
			cmd:    fmt.Sprintf(`rename flow %s new-name-456 %s`, suite.flowName, suite.defaultKosliArguments),
			golden: "flow rename-flow was renamed to new-name-456\n",
		},
		{
			wantError: true,
			name:      "renaming flow fails if the new name is illegal",
			cmd:       fmt.Sprintf(`rename flow %s 'new illegal name' %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: 'new illegal name' is an invalid name for flows. Valid names should start with an alphanumeric and only contain alphanumeric characters, '.', '-', '_' and '~'.\n",
		},
		{
			wantError: true,
			name:      "renaming non-existing flow fails",
			cmd:       fmt.Sprintf(`rename flow non-existing new-name-345 %s`, suite.defaultKosliArguments),
			golden:    "Error: Flow named 'non-existing' does not exist for organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "rename flow fails when 3 args are provided",
			cmd:       fmt.Sprintf(`rename flow %s arg2 arg3 %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 2 arg(s), received 3\n",
		},
		{
			wantError: true,
			name:      "rename flow fails when no args are provided",
			cmd:       fmt.Sprintf(`rename flow %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 2 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRenameFlowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(RenameFlowCommandTestSuite))
}
