package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CreateControlTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *CreateControlTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *CreateControlTestSuite) TestCreateControlCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when no identifier argument is provided",
			cmd:       "create control --name 'My Control'" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "fails when --name is missing",
			cmd:       "create control my-control" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: false,
			name:      "creates a control with identifier and --name",
			cmd:       "create control my-control --name 'My Control'" + suite.defaultKosliArguments,
			golden:    "control my-control was created\n",
		},
		{
			wantError: false,
			name:      "creates a control with a --description",
			cmd:       "create control my-control-2 --name 'My Second Control' --description 'checks something'" + suite.defaultKosliArguments,
			golden:    "control my-control-2 was created\n",
		},
		{
			// Relies on "my-control" created by the earlier test case (cases run
			// sequentially against the same server within the suite).
			wantError:   true,
			name:        "fails with a clear error when the identifier already exists",
			cmd:         "create control my-control --name 'My Control'" + suite.defaultKosliArguments,
			goldenRegex: "^Error: A control with identifier 'my-control' already exists in organization",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestCreateControlTestSuite(t *testing.T) {
	suite.Run(t, new(CreateControlTestSuite))
}
