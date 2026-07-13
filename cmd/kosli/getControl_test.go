package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GetControlCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *GetControlCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateControl(global.Org, "get-control", "Gettable control", suite.T())
}

func (suite *GetControlCommandTestSuite) TestGetControlCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when no identifier argument is provided",
			cmd:       "get control" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			name:   "getting an existing control works",
			cmd:    "get control get-control" + suite.defaultKosliArguments,
			golden: "",
		},
		{
			name:       "getting an existing control as json works",
			cmd:        "get control get-control --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"identifier", "get-control"}, {"name", "Gettable control"}},
		},
		{
			wantError:   true,
			name:        "getting a non-existing control gives a clear error",
			cmd:         "get control no-such-control" + suite.defaultKosliArguments,
			goldenRegex: "^Error: Control 'no-such-control' does not exist in org",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestGetControlCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetControlCommandTestSuite))
}
