package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CreateFlowWithTemplateCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *CreateFlowWithTemplateCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *CreateFlowWithTemplateCommandTestSuite) TestCreateFlowCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       "create flow2 newFlowWithTemplate xxx" + suite.defaultKosliArguments,
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when name is considered invalid by the server",
			cmd:       "create flow2 foo_bar --template-file testdata/valid_template.yml" + suite.defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[name:'foo_bar' does not match '^[a-zA-Z0-9\\\\-]+$']\n",
		},
		{
			wantError: true,
			name:      "fails when --template-file is missing",
			cmd:       "create flow2 newFlowWithTemplate --description \"my new flow\" " + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"template-file\" not set\n",
		},
		{
			wantError:   true,
			name:        "creating a flow with an invalid template fails",
			cmd:         "create flow2 newFlowWithTemplate --template-file testdata/invalid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			goldenRegex: "Error: template file is invalid 1 validation error for Template\n.*",
		},
		{
			name:   "can create a flow with a valid template",
			cmd:    "create flow2 newFlowWithTemplate --template-file testdata/valid_template.yml --description \"my new flow\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithTemplate' was created\n",
		},
		{
			name:   "re-creating a flow updates its metadata",
			cmd:    "create flow2 newFlowWithTemplate --template-file testdata/valid_template.yml --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "flow 'newFlowWithTemplate' was updated\n",
		},
		{
			wantError: true,
			name:      "missing --org flag causes an error",
			cmd:       "create flow2 newFlowWithTemplate --description \"my new flow\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --org is not set\nUsage: kosli create flow2 FLOW-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "create flow2 newFlowWithTemplate --description \"my new flow\" --org cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli create flow2 FLOW-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing name argument fails",
			cmd:       "create flow2 --description \"my new flow\" -H http://localhost:8001 --org cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: flow name must be provided as an argument\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateFlowWithTemplateCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateFlowWithTemplateCommandTestSuite))
}
