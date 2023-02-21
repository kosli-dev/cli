package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CreateFlowCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *CreateFlowCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
}

func (suite *CreateFlowCommandTestSuite) TestCreateFlowCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when arguments are provided",
			cmd:       "create flow --flow newFlow xxx" + suite.defaultKosliArguments,
			golden:    "Error: unknown command \"xxx\" for \"kosli create flow\"\n",
		},
		{
			wantError: true,
			name:      "fails when name is considered invalid by the server",
			cmd:       "create flow --flow foo_bar" + suite.defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[name:'foo_bar' does not match '^[a-zA-Z0-9\\\\-]+$']\n",
		},
		{
			name:   "can create a flow",
			cmd:    "create flow --flow newFlow --description \"my new flow\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlow' was created\n",
		},
		{
			name:   "re-creating a flow updates its metadata",
			cmd:    "create flow --flow newFlow --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlow' was created\n",
		},
		{
			wantError: true,
			name:      "missing --owner flag causes an error",
			cmd:       "create flow --flow newFlow --description \"my new flow\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --owner is not set\nUsage: kosli create flow [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "create flow --flow newFlow --description \"my new flow\" --owner cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli create flow [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --flow and --pipefile causes an error",
			cmd:       "create flow --description \"my new flow\" -H http://localhost:8001 --owner cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: at least one of --flow, --pipefile is required\n",
		},
		{
			wantError: true,
			name:      "providing both --flow and --pipefile fails",
			cmd:       "create flow --flow newFlow --pipefile /path/to/file.json" + suite.defaultKosliArguments,
			golden:    "Error: only one of --flow, --pipefile is allowed\n",
		},
		{
			wantError: true,
			name:      "providing both --description and --pipefile fails",
			cmd:       "create flow --description \"new flow\" --pipefile /path/to/file.json" + suite.defaultKosliArguments,
			golden:    "Error: only one of --description, --pipefile is allowed\n",
		},
		{
			wantError: true,
			name:      "providing --pipefile with non-existing file fails",
			cmd:       "create flow --pipefile /path/to/file.json" + suite.defaultKosliArguments,
			golden:    "Error: failed to open pipefile: open /path/to/file.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateFlowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateFlowCommandTestSuite))
}
