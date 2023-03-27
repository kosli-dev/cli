package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetEnvironmentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
	envType               string
}

func (suite *GetEnvironmentCommandTestSuite) SetupTest() {
	suite.envName = "get-env"
	suite.envType = "K8S"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateEnv(global.Org, suite.envName, suite.envType, suite.T())
}

func (suite *GetEnvironmentCommandTestSuite) TestGetEnvironmentCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "kosli get env newEnv command does not return error",
			cmd:       fmt.Sprintf("get env %s %s", suite.envName, suite.defaultKosliArguments),
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli get env newEnv --output json command does not return error",
			cmd:       fmt.Sprintf("get env %s %s --output json", suite.envName, suite.defaultKosliArguments),
			golden:    "",
		},
		{
			wantError: true,
			name:      "trying to get non-existing env fails",
			cmd:       "get environment non-existing" + suite.defaultKosliArguments,
			golden:    "Error: Environment named 'non-existing' does not exist for organization 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "fails when no argument (env name) provided",
			cmd:       "get environment " + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetEnvironmentCommandTestSuite))
}
