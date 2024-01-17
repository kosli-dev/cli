package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArchiveEnvironmentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	environmentName       string
}

func (suite *ArchiveEnvironmentCommandTestSuite) SetupTest() {
	suite.environmentName = "archive-environment"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateEnv(global.Org, suite.environmentName, "server", suite.T())
}

func (suite *ArchiveEnvironmentCommandTestSuite) TestArchiveEnvironmentCmd() {
	tests := []cmdTestCase{
		{
			name:   "can archive environment",
			cmd:    fmt.Sprintf(`archive environment %s %s`, suite.environmentName, suite.defaultKosliArguments),
			golden: "environment archive-environment was archived\n",
		},
		{
			wantError: true,
			name:      "archiving non-existing environment fails",
			cmd:       fmt.Sprintf(`archive environment non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Environment named 'non-existing' does not exist for organization 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "archive environment fails when 2 args are provided",
			cmd:       fmt.Sprintf(`archive environment %s arg2 %s`, suite.environmentName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "archive environment fails when no args are provided",
			cmd:       fmt.Sprintf(`archive environment %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArchiveEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveEnvironmentCommandTestSuite))
}
