package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CreateEnvironmentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *CreateEnvironmentCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *CreateEnvironmentCommandTestSuite) TestCreateEnvironmentCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "can create K8S env without error",
			cmd:       "create env newEnv1 --type K8S" + suite.defaultKosliArguments,
			golden:    "environment newEnv1 was created\n",
		},
		{
			wantError: false,
			name:      "description can be provided",
			cmd:       "create env newEnv2 --type K8S --description xxx" + suite.defaultKosliArguments,
			golden:    "environment newEnv2 was created\n",
		},
		{
			wantError: true,
			name:      "fails if the type case does not match what the server accepts",
			cmd:       "create env newEnv1 --type k8s" + suite.defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[type:'k8s' is not one of ['K8S', 'ECS', 'S3', 'lambda', 'server', 'docker', 'azure-apps']]\n",
		},
		{
			wantError: true,
			name:      "fails if the type is not recognized by the server",
			cmd:       "create env newEnv1 --type unknown" + suite.defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[type:'unknown' is not one of ['K8S', 'ECS', 'S3', 'lambda', 'server', 'docker', 'azure-apps']]\n",
		},
		{
			wantError: true,
			name:      "fails when name argument is missing",
			cmd:       "create env --type K8S" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "fails when type is missing",
			cmd:       "create env newEnv1" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"type\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when arguments are provided",
			cmd:       "create env newEnv1 --type K8S xxx" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when name is considered invalid by the server",
			cmd:       "create env foo_bar --type K8S" + suite.defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[name:'foo_bar' does not match '^[a-zA-Z][a-zA-Z0-9\\\\-]*$']\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateEnvironmentCommandTestSuite))
}
