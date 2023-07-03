package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CommitEvidenceJiraCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	tmpDir                string
	workTree              *git.Worktree
	fs                    billy.Filesystem
}

func (suite *CommitEvidenceJiraCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_JIRA_API_TOKEN"})
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	var err error
	suite.tmpDir, err = os.MkdirTemp("", "testDir")
	require.NoError(suite.T(), err)
	_, suite.workTree, suite.fs, err = InitializeGitRepo(suite.tmpDir)
	require.NoError(suite.T(), err)

	CreateFlow("flow-for-jira-testing", suite.T())
}

func (suite *CommitEvidenceJiraCommandTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tmpDir)
}

type jiraTestsAdditionalConfig struct {
	commitMessage string
}

func (suite *CommitEvidenceJiraCommandTestSuite) TestCommitEvidenceJiraCommandCmd() {
	tests := []cmdTestCase{
		{
			name: "report Jira commit evidence with tag in start of line works",
			cmd: fmt.Sprintf(`report evidence commit jira --name jira-validation 
					--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
					--repo-root %s
					--build-url example.com %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira evidence is reported to commit: ",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "report Jira commit evidence with tag in middle of line works",
			cmd: fmt.Sprintf(`report evidence commit jira --name jira-validation 
				--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
				--repo-root %s
				--build-url example.com %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira evidence is reported to commit: ",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "Lets test EX-1 test commit",
			},
		},
		{
			name: "report Jira commit evidence with tag in end of line works",
			cmd: fmt.Sprintf(`report evidence commit jira --name jira-validation 
					--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
					--repo-root %s
					--build-url example.com %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira evidence is reported to commit: ",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "Lets test EX-1",
			},
		},
	}
	for _, test := range tests {
		msg := test.additionalConfig.(jiraTestsAdditionalConfig).commitMessage
		commitSha, err := CommitToRepo(suite.workTree, suite.fs, msg)
		require.NoError(suite.T(), err)

		test.cmd = test.cmd + " --commit " + commitSha
		test.golden = test.golden + commitSha + "\n"
		runTestCmd(suite.T(), []cmdTestCase{test})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidenceJiraCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidenceJiraCommandTestSuite))
}
