package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListEnvironmentsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *ListEnvironmentsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
}

func (suite *ListEnvironmentsCommandTestSuite) TestListEnvironmentsCmd() {
	tests := []cmdTestCase{
		{
			name:   "listing environments works",
			cmd:    fmt.Sprintf(`list environments %s`, suite.defaultKosliArguments),
			golden: "No environments found.\n",
		},
		{
			name:   "listing environments with --output json works",
			cmd:    fmt.Sprintf(`list environments --output json %s`, suite.defaultKosliArguments),
			golden: "[]\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListEnvironmentsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListEnvironmentsCommandTestSuite))
}
