package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AddEnvironmentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	logicalEnvName        string
	physicalEnvName       string
}

func (suite *AddEnvironmentCommandTestSuite) SetupTest() {
	suite.logicalEnvName = "mixForAdd"
	suite.physicalEnvName = "pysicalToBeAdded"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateEnv(global.Org, suite.logicalEnvName, "logical", suite.T())
	CreateEnv(global.Org, suite.physicalEnvName, "server", suite.T())
}

func (suite *AddEnvironmentCommandTestSuite) TestAddEnvironmentCmd() {
	tests := []cmdTestCase{
		{
			name: "can add Physical env to Logical environments",
			cmd: fmt.Sprintf(`add environment --physical %s --logical %s %s`,
				suite.physicalEnvName, suite.logicalEnvName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("environment '%s' was added to '%s'\n", suite.physicalEnvName, suite.logicalEnvName),
		},
		{
			wantError: true,
			name:      "must have --physical flag",
			cmd:       fmt.Sprintf(`add environment --logical %s %s`, suite.logicalEnvName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"physical\" not set\n",
		},
		{
			wantError: true,
			name:      "must have --logical flag",
			cmd:       fmt.Sprintf(`add environment --physical %s %s`, suite.physicalEnvName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"logical\" not set\n",
		},
		{
			wantError: true,
			name:      "accept no arguments",
			cmd: fmt.Sprintf(`add environment --physical %s --logical %s SomeThingExtra %s`,
				suite.physicalEnvName, suite.logicalEnvName, suite.defaultKosliArguments),
			golden: "Error: accepts 0 arg(s), received 1\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAddEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AddEnvironmentCommandTestSuite))
}
