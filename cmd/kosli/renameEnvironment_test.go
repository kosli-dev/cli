package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type RenameEnvironmentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *RenameEnvironmentCommandTestSuite) SetupTest() {
	suite.envName = "rename-env"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateEnv(global.Org, suite.envName, "server", suite.T())
}

func (suite *RenameEnvironmentCommandTestSuite) TestListSnapshotsCmd() {
	tests := []cmdTestCase{
		{
			name:   "can rename environment",
			cmd:    fmt.Sprintf(`rename environment %s new-name-456 %s`, suite.envName, suite.defaultKosliArguments),
			golden: "environment rename-env was renamed to new-name-456\n",
		},
		{
			wantError: true,
			name:      "renaming environment fails if the new name is illegal",
			cmd:       fmt.Sprintf(`rename environment %s new_illegal_name %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: 'new_illegal_name' is an invalid name for environments. Valid names should start with a letter and can contain alphanumeric characters and '-'.\n",
		},
		{
			wantError: true,
			name:      "renaming non-existing env fails",
			cmd:       fmt.Sprintf(`rename environment non-existing new-name-345 %s`, suite.defaultKosliArguments),
			golden:    "Error: Environment named 'non-existing' does not exist for organization 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "rename environment fails when 3 args are provided",
			cmd:       fmt.Sprintf(`rename environment %s arg2 arg3 %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 2 arg(s), received 3\n",
		},
		{
			wantError: true,
			name:      "rename environment fails when no args are provided",
			cmd:       fmt.Sprintf(`rename environment %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 2 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRenameEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(RenameEnvironmentCommandTestSuite))
}
