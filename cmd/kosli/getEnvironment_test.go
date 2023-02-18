package main

import (
	"fmt"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetEnvironmentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
}

func (suite *GetEnvironmentCommandTestSuite) SetupTest() {
	suite.flowName = "github-pr"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	kosliClient = requests.NewKosliClient(1, false, log.NewStandardLogger())

	CreateFlow(suite.flowName, suite.T())
}

func (suite *GetEnvironmentCommandTestSuite) TestGetEnvironmentCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "kosli get env newEnv command does not return error",
			cmd:       "get env newEnv" + suite.defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli get env newEnv --output json command does not return error",
			cmd:       "get environment newEnv --output json" + suite.defaultKosliArguments,
			golden:    "",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetEnvironmentCommandTestSuite))
}
