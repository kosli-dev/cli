package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PipelineCommandTestSuite struct {
	suite.Suite
}

func (suite *PipelineCommandTestSuite) TestPipelineCommandCmd() {

	defaultKosliArguments := " -H http://localhost:8001 --owner cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ"

	tests := []cmdTestCase{
		{
			name:   "declare pipeline",
			cmd:    "pipeline declare --pipeline newPipe --description \"my new pipeline\" -H http://localhost:8001 --owner cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden: "",
		},
		{
			name:   "redeclaring a pipeline updates its metadata",
			cmd:    "pipeline declare --pipeline newPipe --description \"changed description\" -H http://localhost:8001 --owner cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden: "",
		},
		{
			wantError: true,
			name:      "missing --owner flag causes an error",
			cmd:       "pipeline declare --pipeline newPipe --description \"my new pipeline\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --owner is not set\nUsage: kosli pipeline declare [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "pipeline declare --pipeline newPipe --description \"my new pipeline\" --owner cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli pipeline declare [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --pipeline causes an error",
			cmd:       "pipeline declare --description \"my new pipeline\" -H http://localhost:8001 --owner cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --pipeline is required when you are not using --pipefile\nUsage: kosli pipeline declare [flags]\n",
		},
		// Pipeline ls tests
		{
			wantError: false,
			name:      "kosli pipeline ls command does not return error",
			cmd:       "pipeline ls" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli pipeline ls --output json command does not return error",
			cmd:       "pipeline ls --output json" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli pipeline ls --output table command does not return error",
			cmd:       "pipeline ls --output table" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: true,
			name:      "kosli pipeline ls --output text command does return error",
			cmd:       "pipeline ls --output text" + defaultKosliArguments,
			golden:    "",
		},

		// Pipeline pipeline get tests
		{
			wantError: false,
			name:      "kosli pipeline get newPipe command does not return error",
			cmd:       "pipeline get newPipe" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli pipeline get newPipe --output json command does not return error",
			cmd:       "pipeline get newPipe --output json" + defaultKosliArguments,
			golden:    "",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPipelineCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PipelineCommandTestSuite))
}
