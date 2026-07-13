package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ArchiveControlCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *ArchiveControlCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateControl(global.Org, "archive-me", "Archive me", suite.T())
}

func (suite *ArchiveControlCommandTestSuite) TestArchiveControlCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when no identifier argument is provided",
			cmd:       "archive control" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			name:   "archiving an existing control works",
			cmd:    "archive control archive-me" + suite.defaultKosliArguments,
			golden: "control archive-me was archived\n",
		},
		{
			wantError:   true,
			name:        "archiving a non-existing control gives a clear error",
			cmd:         "archive control no-such-control" + suite.defaultKosliArguments,
			goldenRegex: "^Error: Control 'no-such-control' does not exist in org",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestArchiveControlCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveControlCommandTestSuite))
}
