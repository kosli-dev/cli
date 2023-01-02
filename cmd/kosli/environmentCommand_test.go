package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EnvironmentCommandTestSuite struct {
	suite.Suite
}

func (suite *EnvironmentCommandTestSuite) TestEnvironmentCommandCmd() {

	defaultKosliArguments := " -H http://localhost:8001 --owner docs-cmd-test-user -a eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"

	tests := []cmdTestCase{
		{
			name:   "declare S3 env",
			cmd:    "environment declare --name newEnv --environment-type S3 --description \"my new env\" " + defaultKosliArguments,
			golden: "",
		},
		{
			wantError: true,
			name:      "declare env with wrong case of type is rejected",
			cmd:       "environment declare --name newK8SEnv --environment-type k8s --description \"my new env\" " + defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[type:'k8s' is not one of ['K8S', 'ECS', 'S3', 'lambda', 'server', 'docker']]\n",
		},
		{
			wantError: true,
			name:      "declare env with illegal name",
			cmd:       "environment declare --name foo_bar --environment-type S3 --description \"my new env\" " + defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[name:'foo_bar' does not match '^[a-zA-Z][a-zA-Z0-9\\\\-]*$']\n",
		},
		{
			// TODO: Is this really updating the environment?
			name:   "re-declaring an env updates its metadata",
			cmd:    "environment declare --name newEnv --environment-type S3 --description \"changed description\" " + defaultKosliArguments,
			golden: "",
		},
		{
			wantError: true,
			name:      "missing --owner flag causes an error",
			cmd:       "environment declare --name newEnv --environment-type S3 --description \"my new env\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --owner is not set\nUsage: kosli environment declare [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "environment declare --name newEnv --environment-type S3 --description \"my new env\" --owner cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli environment declare [flags]\n",
		},
		{
			wantError: true,
			name:      "unknown --environment-type causes an error",
			cmd:       "environment declare --name newEnv --environment-type UNKNOWN --description \"my new env\" " + defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[type:'UNKNOWN' is not one of ['K8S', 'ECS', 'S3', 'lambda', 'server', 'docker']]\n",
		},
		{
			wantError: true,
			name:      "missing --name causes an error",
			cmd:       "environment declare --environment-type S3 --description \"my new env\" " + defaultKosliArguments,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --environment-type causes an error",
			cmd:       "environment declare --name newEnv --description \"my new env\" " + defaultKosliArguments,
			golden:    "Error: required flag(s) \"environment-type\" not set\n",
		},
		// Environment ls tests
		{
			wantError: false,
			name:      "kosli env ls command lists newEnv does not return error",
			cmd:       "env ls" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli env ls --output json command does not return error",
			cmd:       "env ls --output json" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli env ls --output table command does not return error",
			cmd:       "env ls --output table" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: true,
			name:      "kosli env ls --output text command does return error",
			cmd:       "env ls --output text" + defaultKosliArguments,
			golden:    "",
		},

		// Environment env get tests
		{
			wantError: false,
			name:      "kosli env inspect newEnv command does not return error",
			cmd:       "env inspect newEnv" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli env inspect newEnv --output json command does not return error",
			cmd:       "env inspect newEnv --output json" + defaultKosliArguments,
			golden:    "",
		},

		// Environment rename tests
		{
			name:   "rename: create initial environment",
			cmd:    "environment declare --name firstEnvName --environment-type S3 --description \"first environment\" " + defaultKosliArguments,
			golden: "",
		},
		{
			name:   "rename: rename from firstEnvName to secondEnvName",
			cmd:    "environment rename firstEnvName secondEnvName" + defaultKosliArguments,
			golden: "",
		},
		{
			name:   "rename: can get env based on firstEnvName",
			cmd:    "env inspect firstEnvName" + defaultKosliArguments,
			golden: "",
		},
		{
			name:   "rename: can get env based on secondEnvName",
			cmd:    "env inspect secondEnvName" + defaultKosliArguments,
			golden: "",
		},
		{
			wantError: true,
			name:      "rename: cannot rename a non existing environment",
			cmd:       "environment rename unknownEnvName someOtherEnvName" + defaultKosliArguments,
			golden:    "",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentCommandTestSuite))
}
