package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetApprovalCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	fingerprint           string
}

func (suite *GetApprovalCommandTestSuite) SetupTest() {
	suite.flowName = "get-approval"
	suite.fingerprint = "7a498bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	CreateArtifact(suite.flowName, suite.fingerprint, "approved-artifact", suite.T())
	CreateApproval(suite.flowName, suite.fingerprint, suite.T())
	CreateApproval(suite.flowName, suite.fingerprint, suite.T())
}

func (suite *GetApprovalCommandTestSuite) TestGetApprovalCmd() {
	tests := []cmdTestCase{
		{
			name:       "get latest approval works",
			cmd:        fmt.Sprintf("get approval %s %s", suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-approval-latest.txt",
		},
		{
			name:       "get an approval works with # expression",
			cmd:        fmt.Sprintf("get approval %s#1 %s", suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-approval.txt",
		},
		{
			name:       "get an approval works with ~ expression",
			cmd:        fmt.Sprintf("get approval %s~1 %s", suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-approval.txt",
		},
		{
			wantError: true,
			name:      "get an approval with more than one argument fails",
			cmd:       fmt.Sprintf("get approval %s xxx %s", suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "get approval on a non-existing flow fails",
			cmd:       "get approval get-approval-123#20" + suite.defaultKosliArguments,
			golden:    "Error: Flow named 'get-approval-123' does not exist for organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "get non-existing approval fails",
			cmd:       fmt.Sprintf("get approval %s#23 %s", suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: Approval number '23' does not exist in flow 'get-approval' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "missing --org fails",
			cmd:       fmt.Sprintf("get approval %s --api-token secret", suite.flowName),
			golden:    "Error: --org is not set\nUsage: kosli get approval EXPRESSION [flags]\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetApprovalCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetApprovalCommandTestSuite))
}
