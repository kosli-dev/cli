package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UpdateControlCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *UpdateControlCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateControl(global.Org, "update-me", "Update me", suite.T())
}

func (suite *UpdateControlCommandTestSuite) TestUpdateControlCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when no identifier argument is provided",
			cmd:       "update control --name 'New name'" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "fails when no updatable flags are provided",
			cmd:       "update control update-me" + suite.defaultKosliArguments,
			golden:    "Error: at least one of --name, --description, --link is required\n",
		},
		{
			name:   "updating a control's name works",
			cmd:    "update control update-me --name 'Updated name'" + suite.defaultKosliArguments,
			golden: "control update-me was updated\n",
		},
		{
			name:   "updating a control's description works",
			cmd:    "update control update-me --description 'checks something new'" + suite.defaultKosliArguments,
			golden: "control update-me was updated\n",
		},
		{
			name:   "updating a control's links works",
			cmd:    "update control update-me --link runbook=https://example.com/runbook" + suite.defaultKosliArguments,
			golden: "control update-me was updated\n",
		},
		{
			name:   "updating multiple fields in one call works",
			cmd:    "update control update-me --name 'Another name' --description 'and a description'" + suite.defaultKosliArguments,
			golden: "control update-me was updated\n",
		},
		{
			wantError:   true,
			name:        "updating a non-existing control gives a clear error",
			cmd:         "update control no-such-control --name 'New name'" + suite.defaultKosliArguments,
			goldenRegex: "^Error: Control 'no-such-control' does not exist in org",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestUpdateControlCommandTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateControlCommandTestSuite))
}
