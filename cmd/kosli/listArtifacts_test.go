package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListArtifactsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
}

func (suite *ListArtifactsCommandTestSuite) SetupTest() {
	suite.flowName = "list-artifacts"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	CreateFlow(suite.flowName, suite.T())
}

func (suite *ListArtifactsCommandTestSuite) TestListArtifactsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing flow name arg causes an error",
			cmd:       fmt.Sprintf(`list artifacts %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "non-existing flow causes an error",
			cmd:       fmt.Sprintf(`list artifacts non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Pipeline called 'non-existing' does not exist for Organization 'docs-cmd-test-user'.\n",
		},
		// TODO: the correct error is overwritten by the hack flag value check in root.go
		{
			wantError: true,
			name:      "negative page number causes an error",
			cmd:       fmt.Sprintf(`list artifacts foo --page -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "negative page limit causes an error",
			cmd:       fmt.Sprintf(`list artifacts foo --page-limit -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page-limit' has value '-1' which is illegal\n",
		},
		{
			name:   "listing artifacts on an empty flow works",
			cmd:    fmt.Sprintf(`list artifacts %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden: "No artifacts were found\n",
		},
		{
			name:   "listing artifacts on an empty flow with --output json works",
			cmd:    fmt.Sprintf(`list artifacts %s --output json %s`, suite.flowName, suite.defaultKosliArguments),
			golden: "[]\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListArtifactsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListArtifactsCommandTestSuite))
}
