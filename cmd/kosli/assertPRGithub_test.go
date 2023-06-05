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
type AssertPRGithubCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *AssertPRGithubCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *AssertPRGithubCommandTestSuite) TestAssertPRGithubCmd() {
	tests := []cmdTestCase{
		{
			name: "assert Github PR evidence passes when commit has a PR in github",
			cmd: `assert pullrequest github --github-org kosli-dev --repository cli
			--commit ` + testHelpers.GithubCommitWithPR() + suite.defaultKosliArguments,
			golden: fmt.Sprintf("found [1] pull request(s) in Github for commit: %s\n", testHelpers.GithubCommitWithPR()),
		},
		{
			wantError: true,
			name:      "assert Github PR evidence fails when commit has no PRs in github",
			cmd: `assert pullrequest github --github-org kosli-dev --repository cli 
			--commit 19aab7f063147614451c88969602a10afbabb43d` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 19aab7f063147614451c88969602a10afbabb43d\n",
		},
		{
			wantError: true,
			name:      "assert Github PR evidence fails when commit does not exist",
			cmd: `assert pullrequest github --github-org kosli-dev --repository cli 
			--commit 19aab7f063147614451c88969602a10afba123ab` + suite.defaultKosliArguments,
			golden: "Error: GET https://api.github.com/repos/kosli-dev/cli/commits/19aab7f063147614451c88969602a10afba123ab/pulls: 422 No commit found for SHA: 19aab7f063147614451c88969602a10afba123ab []\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertPRGithubCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertPRGithubCommandTestSuite))
}
