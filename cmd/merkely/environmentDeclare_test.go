package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EnvironmentDeclareTestSuite struct {
	suite.Suite
}

func (suite *EnvironmentDeclareTestSuite) TestEnvironmentDeclareCmd() {
	tests := []cmdTestCase{
		{
			name:   "declare S3 env",
			cmd:    "environment declare --name newEnv --environment-type S3 --description \"my new env\" -H http://localhost:8001 -o cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden: "",
		},
		{
			name:   "redeclaring an env updates its metadata",
			cmd:    "environment declare --name newEnv --environment-type S3 --description \"changed description\" -H http://localhost:8001 -o cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden: "",
		},
		{
			wantError: true,
			name:      "missing --owner flag causes an error",
			cmd:       "environment declare --name newEnv --environment-type S3 --description \"my new env\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --owner is not set\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "environment declare --name newEnv --environment-type S3 --description \"my new env\" -o cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\n",
		},
		{
			wantError: true,
			name:      "unknown --environment-type causes an error",
			cmd:       "environment declare --name newEnv --environment-type UNKNOWN --description \"my new env\" -H http://localhost:8001 -o cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: UNKNOWN is not a valid environment type\n",
		},
		{
			wantError: true,
			name:      "missing --name causes an error",
			cmd:       "environment declare --environment-type S3 --description \"my new env\" -H http://localhost:8001 -o cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --environment-type causes an error",
			cmd:       "environment declare --name newEnv --description \"my new env\" -H http://localhost:8001 -o cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: required flag(s) \"environment-type\" not set\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnvironmentDeclareTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentDeclareTestSuite))
}
