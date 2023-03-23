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
type CommitEvidencePRBitbucketCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowNames             string
}

func (suite *CommitEvidencePRBitbucketCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_BITBUCKET_PASSWORD"})

	suite.flowNames = "bitbucket-pr"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowNames, suite.T())
}

func (suite *CommitEvidencePRBitbucketCommandTestSuite) TestArtifactEvidencePRBitbucketCmd() {
	tests := []cmdTestCase{
		{
			name: "report Bitbucket PR evidence works",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test --commit 2492011ef04a9da09d35be706cf6a4c5bc6f1e69` + suite.defaultKosliArguments,
			golden: "bitbucket pull request evidence is reported to commit: 2492011ef04a9da09d35be706cf6a4c5bc6f1e69\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --org is missing",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test --commit 2492011ef04a9da09d35be706cf6a4c5bc6f1e69 --api-token foo --host bar`,
			golden: "Error: --org is not set\n" +
				"Usage: kosli report evidence commit pullrequest bitbucket [flags]\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --name is missing",
			cmd: `report evidence commit pullrequest bitbucket --flows ` + suite.flowNames + `
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test --commit 2492011ef04a9da09d35be706cf6a4c5bc6f1e69` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --bitbucket-username is missing",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url example.com --bitbucket-workspace ewelinawilkosz --repository cli-test --commit 2492011ef04a9da09d35be706cf6a4c5bc6f1e69` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"bitbucket-username\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --repository is missing",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --commit 2492011ef04a9da09d35be706cf6a4c5bc6f1e69` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --commit is missing",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when commit does not exist",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test --commit 73d7fee2f31ade8e1a9c456c324255212c3123ab` + suite.defaultKosliArguments,
			golden: "Error: map[error:map[message:Resource not found] type:error]\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --assert is used and commit has no PRs",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--assert
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test --commit cb6ec5fcbb25b1ebe4859d35ab7995ab973f894c` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: cb6ec5fcbb25b1ebe4859d35ab7995ab973f894c\n",
		},
		{
			name: "report Bitbucket PR evidence does not fail when commit has no PRs",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
			          --build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test --commit cb6ec5fcbb25b1ebe4859d35ab7995ab973f894c` + suite.defaultKosliArguments,
			golden: "no pull requests found for given commit: cb6ec5fcbb25b1ebe4859d35ab7995ab973f894c\n" +
				"bitbucket pull request evidence is reported to commit: cb6ec5fcbb25b1ebe4859d35ab7995ab973f894c\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --user-data is not found",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--user-data non-existing.json
					--build-url example.com --bitbucket-username ewelinawilkosz --bitbucket-workspace ewelinawilkosz --repository cli-test --commit 2492011ef04a9da09d35be706cf6a4c5bc6f1e69` + suite.defaultKosliArguments,
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidencePRBitbucketCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidencePRBitbucketCommandTestSuite))
}
