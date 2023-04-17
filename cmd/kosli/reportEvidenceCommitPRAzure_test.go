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
type CommitEvidencePRAzureCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowNames             string
}

func (suite *CommitEvidencePRAzureCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_AZURE_TOKEN"})

	suite.flowNames = "azure-pr"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowNames, suite.T())
}

func (suite *CommitEvidencePRAzureCommandTestSuite) TestCommitEvidencePRAzureCmd() {
	tests := []cmdTestCase{
		{
			name: "report Azure PR evidence works",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "azure pull request evidence is reported to commit: 5f61be8f00a01c84e491922a630c9a418c684c7a\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --org is missing",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a --api-token foo --host bar`,
			golden: "Error: --org is not set\n" +
				"Usage: kosli report evidence commit pullrequest azure [flags]\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --name is missing",
			cmd: `report evidence commit pullrequest azure --flows ` + suite.flowNames + ` --azure-org-url https://dev.azure.com/kosli --project kosli-azure
			          --build-url example.com --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --azure-org-url is missing",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"azure-org-url\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --project is missing",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"project\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --repository is missing",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --commit is missing",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			name: "report Azure PR evidence does not fail when commit does not exist, empty evidence is reported instead",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 1111111111111111111111111111111111111111` + suite.defaultKosliArguments,
			golden: "no pull requests found for given commit: 1111111111111111111111111111111111111111\n" +
				"azure pull request evidence is reported to commit: 1111111111111111111111111111111111111111\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --assert is used and commit has no PRs",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + ` --assert
					--build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 1a877d0c3cf4644b4225bf3eb23ced26818d685a` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 1a877d0c3cf4644b4225bf3eb23ced26818d685a\n",
		},
		{
			name: "report Azure PR evidence does not fail when commit has no PRs",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 1a877d0c3cf4644b4225bf3eb23ced26818d685a` + suite.defaultKosliArguments,
			golden: "no pull requests found for given commit: 1a877d0c3cf4644b4225bf3eb23ced26818d685a\n" +
				"azure pull request evidence is reported to commit: 1a877d0c3cf4644b4225bf3eb23ced26818d685a\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --user-data is not found",
			cmd: `report evidence commit pullrequest azure --name az-pr --flows ` + suite.flowNames + `
					  --user-data non-existing.json
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidencePRAzureCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidencePRAzureCommandTestSuite))
}
