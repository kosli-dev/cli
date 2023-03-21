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
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_BITBUCKET_PASSWORD"})

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
			cmd: `assert pullrequest bitbucket --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test 
			--commit 2492011ef04a9da09d35be706cf6a4c5bc6f1e69` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Bitbucket for commit: 2492011ef04a9da09d35be706cf6a4c5bc6f1e69\n",
		},
		{
			wantError: true,
			name:      "assert Bitbucket PR evidence fails when commit has no PRs in bitbucket",
			cmd: `assert pullrequest bitbucket --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test 
			--commit cb6ec5fcbb25b1ebe4859d35ab7995ab973f894c` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: cb6ec5fcbb25b1ebe4859d35ab7995ab973f894c\n",
		},
		{
			wantError: true,
			name:      "assert Bitbucket PR evidence fails when commit does not exist",
			cmd: `assert pullrequest bitbucket --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test 
			--commit 19aab7f063147614451c88969602a10afba123ab` + suite.defaultKosliArguments,
			golden: "Error: map[error:map[message:Resource not found] type:error]\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertPRBitbucketCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertPRBitbucketCommandTestSuite))
}
