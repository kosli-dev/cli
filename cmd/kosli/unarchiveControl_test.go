package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UnarchiveControlCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *UnarchiveControlCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateControl(global.Org, "unarchive-me", "Unarchive me", suite.T())
	ArchiveControl(global.Org, "unarchive-me", suite.T())
}

func (suite *UnarchiveControlCommandTestSuite) TestUnarchiveControlCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when no identifier argument is provided",
			cmd:       "unarchive control" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			name:   "unarchiving an archived control works",
			cmd:    "unarchive control unarchive-me" + suite.defaultKosliArguments,
			golden: "control unarchive-me was unarchived\n",
		},
		{
			name:       "the control is active afterwards",
			cmd:        "get control unarchive-me --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"archived", false}},
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestUnarchiveControlCommandTestSuite(t *testing.T) {
	suite.Run(t, new(UnarchiveControlCommandTestSuite))
}
