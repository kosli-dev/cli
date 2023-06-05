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
type CommitEvidencePRGithubCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowNames             string
	commitWithPR          string
}

func (suite *CommitEvidencePRGithubCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})

	suite.flowNames = "github-pr"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	suite.commitWithPR = testHelpers.GithubCommitWithPR()

	CreateFlow(suite.flowNames, suite.T())
}

func (suite *CommitEvidencePRGithubCommandTestSuite) TestCommitEvidencePRGithubCmd() {
	tests := []cmdTestCase{
		{
			name: "report Github PR evidence works",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit ` + suite.commitWithPR + suite.defaultKosliArguments,
			golden: fmt.Sprintf("github pull request evidence is reported to commit: %s\n", suite.commitWithPR),
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --org is missing",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6 --api-token foo --host bar`,
			golden: "Error: --org is not set\n" +
				"Usage: kosli report evidence commit pullrequest github [flags]\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --name is missing",
			cmd: `report evidence commit pullrequest github --flows ` + suite.flowNames + ` --github-org kosli-dev
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --github-org is missing",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"github-org\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --repository is missing",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --github-org kosli-dev --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --commit is missing",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --github-org kosli-dev --repository cli` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when commit does not exist",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 1111111111111111111111111111111111111111` + suite.defaultKosliArguments,
			golden: "Error: GET https://api.github.com/repos/kosli-dev/cli/commits/1111111111111111111111111111111111111111/pulls: 422 No commit found for SHA: 1111111111111111111111111111111111111111 []\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --assert is used and commit has no PRs",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + ` --assert
					--build-url example.com --github-org kosli-dev --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n",
		},
		{
			name: "report Github PR evidence does not fail when commit has no PRs",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "no pull requests found for given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n" +
				"github pull request evidence is reported to commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --user-data is not found",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
					  --user-data non-existing.json
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
		{
			name: "report Github PR evidence works with --repository=owner/repo",
			cmd: `report evidence commit pullrequest github --name gh-pr --flows ` + suite.flowNames + `
			          --build-url example.com --github-org kosli-dev --repository kosli-dev/cli --commit ` + suite.commitWithPR + suite.defaultKosliArguments,
			golden: fmt.Sprintf("github pull request evidence is reported to commit: %s\n", suite.commitWithPR),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidencePRGithubCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidencePRGithubCommandTestSuite))
}
