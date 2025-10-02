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
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *CreateFlowCommandTestSuite) TestCreateFlowCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       "create flow newFlow xxx" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError:   true,
			name:        "fails when name is considered invalid by the server",
			cmd:         "create flow 'foo bar'" + suite.defaultKosliArguments,
			goldenRegex: "^Error: Input payload validation failed: .*foo bar",
		},
		{
			name:   "can create a flow (by default legacy template is used)",
			cmd:    "create flow newFlow --description \"my new flow\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlow' was created\n",
		},
		{
			name:   "re-creating a flow updates its metadata",
			cmd:    "create flow newFlow --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlow' was updated\n",
		},
		{
			wantError: true,
			name:      "missing --org flag causes an error",
			cmd:       "create flow newFlow --description \"my new flow\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --org is not set\nUsage: kosli create flow FLOW-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "create flow newFlow --description \"my new flow\" --org cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli create flow FLOW-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing name argument fails",
			cmd:       "create flow --description \"my new flow\" -H http://localhost:8001 --org cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "cannot use --template and --template-file together",
			cmd:       "create flow newFlow --description \"my new flow\" --template foo --template-file testdata/valid_template.yml" + suite.defaultKosliArguments,
			golden:    "Error: only one of --template, --template-file is allowed\n",
		},
		// flows v2
		{
			name:   "can create a flow with a valid template",
			cmd:    "create flow newFlowWithTemplate --template-file testdata/valid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithTemplate' was created\n",
		},
		{
			name:   "re-creating a flow (with template) updates its metadata",
			cmd:    "create flow newFlowWithTemplate --template-file testdata/valid_template.yml --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithTemplate' was updated\n",
		},
		{
			wantError:   true,
			name:        "creating a flow with an invalid template fails",
			cmd:         "create flow newFlowWithTemplate --template-file testdata/invalid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			goldenRegex: "Error: Input payload validation failed.*",
		},
		{
			wantError: true,
			name:      "fails when both --template-file and --use-empty-template are provided",
			cmd:       "create flow newFlowWithTemplate --use-empty-template --template-file testdata/valid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			golden:    "Error: only one of --template-file, --use-empty-template is allowed\n",
		},
		{
			name:   "creating a flow with --use-empty-template works",
			cmd:    "create flow newFlowWithEmptyTemplate --use-empty-template --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithEmptyTemplate' was created\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateFlowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateFlowCommandTestSuite))
}
