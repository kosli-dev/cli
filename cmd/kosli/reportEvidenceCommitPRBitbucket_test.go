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
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{"KOSLI_BITBUCKET_ACCESS_TOKEN"})

	suite.flowNames = "bitbucket-pr"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowNames, suite.Suite.T())
}

func (suite *CommitEvidencePRBitbucketCommandTestSuite) TestCommitEvidencePRBitbucketCmd() {
	tests := []cmdTestCase{
		{
			name: "report Bitbucket PR evidence works",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "found 1 pull request\\(s\\) for commit: fd54040fc90e7e83f7b152619bfa18917b72c34f\nbitbucket pull request evidence is reported to commit: fd54040fc90e7e83f7b152619bfa18917b72c34f\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --org is missing",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f --api-token foo --host bar`,
			goldenRegex: "Error: --org is not set\n" +
				"Usage: kosli report evidence commit pullrequest bitbucket \\[flags\\]\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --name is missing",
			cmd: `report evidence commit pullrequest bitbucket --flows ` + suite.flowNames + `
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: required flag\\(s\\) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --repository is missing",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: required flag\\(s\\) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --commit is missing",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test` + suite.defaultKosliArguments,
			goldenRegex: "Error: required flag\\(s\\) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when commit does not exist",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit 73d7fee2f31ade8e1a9c456c324255212c3123ab` + suite.defaultKosliArguments,
			goldenRegex: "Error: map\\[error:map\\[message:Resource not found\\] type:error\\]\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --assert is used and commit has no PRs",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--assert
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit 3dce097040987c4693d2e4be817474d9d0063c93` + suite.defaultKosliArguments,
			goldenRegex: "found 0 pull request\\(s\\) for commit: .*\nbitbucket pull request evidence is reported to commit: .*\nError: assert failed: no pull request found for the given commit: .*\n",
		},
		{
			name: "report Bitbucket PR evidence does not fail when commit has no PRs",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit 3dce097040987c4693d2e4be817474d9d0063c93` + suite.defaultKosliArguments,
			goldenRegex: "found 0 pull request\\(s\\) for commit: 3dce097040987c4693d2e4be817474d9d0063c93\n" +
				"bitbucket pull request evidence is reported to commit: 3dce097040987c4693d2e4be817474d9d0063c93\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --user-data is not found",
			cmd: `report evidence commit pullrequest bitbucket --name bb-pr --flows ` + suite.flowNames + `
					--user-data non-existing.json
					--build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: open non-existing.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidencePRBitbucketCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidencePRBitbucketCommandTestSuite))
}
