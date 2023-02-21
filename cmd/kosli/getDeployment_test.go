package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetDeploymentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
}

func (suite *GetDeploymentCommandTestSuite) SetupTest() {
	suite.flowName = "github-pr"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
}

func (suite *GetDeploymentCommandTestSuite) TestGetDeploymentCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "get deployment fails when --name flag is missing",
			cmd:       `get deployment ` + suite.flowName + `#1 --api-token foo --host bar`,
			golden: "Error: --owner is not set\n" +
				"Usage: kosli get deployment SNAPPISH [flags]\n",
		},
		{
			wantError: true,
			name:      "get deployment fails when --api-token flag is missing",
			cmd:       `get deployment ` + suite.flowName + `#1 --owner foo --host bar`,
			golden: "Error: --api-token is not set\n" +
				"Usage: kosli get deployment SNAPPISH [flags]\n",
		},
		{
			wantError: true,
			name:      "get deployment fails when flow does not exist",
			cmd:       `get deployment foo#1` + suite.defaultKosliArguments,
			golden:    "Error: Pipeline called 'foo' does not exist for Organization 'docs-cmd-test-user'. \n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetDeploymentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetDeploymentCommandTestSuite))
}
