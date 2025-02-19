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
type AssertPRBitbucketCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *AssertPRBitbucketCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{"KOSLI_BITBUCKET_PASSWORD"})

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *AssertPRBitbucketCommandTestSuite) TestAssertPRBitbucketCmd() {
	tests := []cmdTestCase{
		{
			name: "assert Bitbucket PR evidence passes when commit has a PR in bitbucket",
			cmd: `assert pullrequest bitbucket --bitbucket-workspace kosli-dev --repository cli-test 
			--commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Bitbucket for commit: fd54040fc90e7e83f7b152619bfa18917b72c34f\n",
		},
		{
			wantError: true,
			name:      "assert Bitbucket PR evidence fails when commit has no PRs in bitbucket",
			cmd: `assert pullrequest bitbucket --bitbucket-workspace kosli-dev --repository cli-test 
			--commit 3dce097040987c4693d2e4be817474d9d0063c93` + suite.defaultKosliArguments,
			golden: "Error: assert failed: found no pull request(s) in Bitbucket for commit: 3dce097040987c4693d2e4be817474d9d0063c93\n",
		},
		{
			wantError: true,
			name:      "assert Bitbucket PR evidence fails when commit does not exist",
			cmd: `assert pullrequest bitbucket --bitbucket-workspace kosli-dev --repository cli-test 
			--commit 19aab7f063147614451c88969602a10afba123ab` + suite.defaultKosliArguments,
			golden: "Error: map[error:map[message:Resource not found] type:error]\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertPRBitbucketCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertPRBitbucketCommandTestSuite))
}
