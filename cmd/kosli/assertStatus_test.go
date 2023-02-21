package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AssertStatusCommandTestSuite struct {
	suite.Suite
}

func (suite *AssertStatusCommandTestSuite) TestAssertPRGitlabCmd() {
	tests := []cmdTestCase{
		{
			name:   "assert status works",
			cmd:    `assert status`,
			golden: "OK\n",
		},
		{
			wantError: true,
			name:      "assert status on a non-existing host fails",
			cmd:       `assert status -H https://kosli.example.com`,
			golden:    "Error: kosli server https://kosli.example.com is unresponsive\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertStatusCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertStatusCommandTestSuite))
}
