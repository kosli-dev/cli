package main

import (
	"fmt"
	"testing"

	ghUtils "github.com/kosli-dev/cli/internal/github"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/suite"
)

type AssertPRGithubCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	commitWithPR          string
	commitWithNoPR        string
}

func (suite *AssertPRGithubCommandTestSuite) SetupTest() {
	suite.commitWithPR = "480e5a00379a52b8e184d6815080242a878ca295"
	suite.commitWithNoPR = "7d1db1c8b7e71ee0ce369f1b722cc8844d3a7af6"

	ghUtils.NewGithubRetrieverFunc = func(token, baseURL, org, repository string) types.PRRetriever {
		return &ghUtils.FakeGitHubClient{
			PRsByCommit: map[string][]*types.PREvidence{
				suite.commitWithPR: {{URL: "https://github.com/kosli-dev/cli/pull/1", State: "MERGED"}},
			},
		}
	}

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --github-token fake --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *AssertPRGithubCommandTestSuite) TearDownTest() {
	ghUtils.ResetGithubRetrieverFunc()
}

func (suite *AssertPRGithubCommandTestSuite) TestAssertPRGithubCmd() {
	tests := []cmdTestCase{
		{
			name: "assert Github PR evidence passes when commit has a PR in github",
			cmd: `assert pullrequest github --github-org kosli-dev --repository cli
			--commit ` + suite.commitWithPR + suite.defaultKosliArguments,
			golden: fmt.Sprintf("found [1] pull request(s) in Github for commit: %s\n", suite.commitWithPR),
		},
		{
			wantError: true,
			name:      "assert Github PR evidence fails when commit has no PRs in github",
			cmd: `assert pullrequest github --github-org kosli-dev --repository cli
			--commit ` + suite.commitWithNoPR + suite.defaultKosliArguments,
			golden: fmt.Sprintf("Error: assert failed: found no pull request(s) in Github for commit: %s\n", suite.commitWithNoPR),
		},
		{
			wantError: true,
			name:      "assert Github PR evidence fails when commit does not exist",
			cmd: `assert pullrequest github --github-org kosli-dev --repository cli
			--commit 0000000000000000000000000000000000000000` + suite.defaultKosliArguments,
			golden: "Error: assert failed: found no pull request(s) in Github for commit: 0000000000000000000000000000000000000000\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestAssertPRGithubCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertPRGithubCommandTestSuite))
}
