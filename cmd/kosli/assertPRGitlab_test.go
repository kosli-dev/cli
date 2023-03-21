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
type AssertPRGitlabCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *AssertPRGitlabCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITLAB_TOKEN"})

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *AssertPRGitlabCommandTestSuite) TestAssertPRGitlabCmd() {
	tests := []cmdTestCase{
		{
			name: "assert Gitlab PR evidence passes when commit has a PR in gitlab",
			cmd: `assert mergerequest gitlab --gitlab-org ewelinawilkosz --repository merkely-gitlab-demo
			--commit e6510880aecdc05d79104d937e1adb572bd91911` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Gitlab for commit: e6510880aecdc05d79104d937e1adb572bd91911\n",
		},
		{
			name: "assert Gitlab PR evidence with aliases 1",
			cmd: `assert mr gl --gitlab-org ewelinawilkosz --repository merkely-gitlab-demo
			--commit e6510880aecdc05d79104d937e1adb572bd91911` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Gitlab for commit: e6510880aecdc05d79104d937e1adb572bd91911\n",
		},
		{
			name: "assert Gitlab PR evidence with aliases 2",
			cmd: `assert pullrequest gitlab --gitlab-org ewelinawilkosz --repository merkely-gitlab-demo
			--commit e6510880aecdc05d79104d937e1adb572bd91911` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Gitlab for commit: e6510880aecdc05d79104d937e1adb572bd91911\n",
		},
		{
			wantError: true,
			name:      "assert Gitlab PR evidence fails when commit has no PRs in gitlab",
			cmd: `assert mergerequest gitlab --gitlab-org ewelinawilkosz --repository merkely-gitlab-demo
			--commit 2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6` + suite.defaultKosliArguments,
			golden: "Error: no merge requests found for the given commit: 2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6\n",
		},
		{
			wantError: true,
			name:      "assert Gitlab PR evidence fails when commit does not exist",
			cmd: `assert mergerequest gitlab --gitlab-org kosli-dev --repository cli
			--commit 1111111111111111111111111111111111111111` + suite.defaultKosliArguments,
			golden: "Error: GET https://gitlab.com/api/v4/projects/kosli-dev/cli/repository/commits/1111111111111111111111111111111111111111/merge_requests: 404 {message: 404 Project Not Found}\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertPRGitlabCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertPRGitlabCommandTestSuite))
}
