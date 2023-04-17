package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AssertPRAzureCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *AssertPRAzureCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_AZURE_TOKEN"})

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *AssertPRAzureCommandTestSuite) TestAssertPRAzureCmd() {
	tests := []cmdTestCase{
		{
			name: "assert Azure PR evidence passes when commit has a PR in Azure",
			cmd: `assert pullrequest azure --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli 
			--commit e6b38318747f1c225e6d2cdba1e88aa00fbcae29` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Azure DevOps for commit: e6b38318747f1c225e6d2cdba1e88aa00fbcae29\n",
		},
		{
			wantError: true,
			name:      "assert Azure PR evidence fails when commit has no PRs in Azure",
			cmd: `assert pullrequest azure --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli 
			--commit 58d8aad96e0dcd11ada3dc6650d23909eed336ed` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 58d8aad96e0dcd11ada3dc6650d23909eed336ed\n",
		},
		{
			wantError: true,
			name:      "assert Azure PR evidence fails when commit does not exist",
			cmd: `assert pullrequest azure --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli 
			--commit c4fa4c2ce6bef984abc93be9258a85f9137ff1c9` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: c4fa4c2ce6bef984abc93be9258a85f9137ff1c9\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertPRAzureCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertPRAzureCommandTestSuite))
}
