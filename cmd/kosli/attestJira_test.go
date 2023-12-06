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
type AttestJiraCommandTestSuite struct {
	suite.Suite
	flowName              string
	trailName             string
	artifactFingerprint   string
	tmpDir                string
	workTree              *git.Worktree
	fs                    billy.Filesystem
	defaultKosliArguments string
}

func (suite *AttestJiraCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_JIRA_API_TOKEN"})
	suite.flowName = "attest-jira"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)

	var err error
	suite.tmpDir, err = os.MkdirTemp("", "testDir")
	require.NoError(suite.T(), err)
	_, suite.workTree, suite.fs, err = testHelpers.InitializeGitRepo(suite.tmpDir)
	require.NoError(suite.T(), err)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.T())
}

func (suite *AttestJiraCommandTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tmpDir)
}

func (suite *AttestJiraCommandTestSuite) TestAttestJiraCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest jira foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when missing required flags",
			cmd:       fmt.Sprintf("attest jira foo --jira-username tore@kosli.com %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"commit\", \"jira-base-url\", \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when missing both --jira-username and --jira-pat flag",
			cmd: fmt.Sprintf(`attest jira foo --name bar
							--jira-base-url https://kosli-test.atlassian.net  --jira-api-token xxx
							%s`, suite.defaultKosliArguments),
			golden: "Error: at least one of --jira-pat, --jira-username is required\n",
		},
		{
			wantError: true,
			name:      "fails when missing --commit flag",
			cmd: fmt.Sprintf(`attest jira foo --name bar
							--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
							--jira-api-token secret
							%s`, suite.defaultKosliArguments),
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest jira testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --url example.com %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest jira --name foo --fingerprint xxxx --commit HEAD --url example.com --jira-username tore@kosli.com %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest jira [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd: fmt.Sprintf(`attest jira --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 
								--name foo
								--repo-root %s
								--jira-base-url https://kosli-test.atlassian.net
								--jira-username tore@kosli.com %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-jira' belonging to organization 'docs-cmd-test-user'\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "test commit",
			},
		},
		{
			wantError: true,
			name:      "assert for non-existing Jira issue gives an error",
			cmd: fmt.Sprintf(`attest jira --name jira-validation 
					--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
					--repo-root %s
					--assert %s`, suite.tmpDir, suite.defaultKosliArguments),
			goldenRegex: "Error: missing Jira issues from references found in commit message or branch name.*",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "SAMI-1 test commit",
			},
		},
		{
			name: "can attest jira against an artifact using artifact name and --artifact-type",
			cmd: fmt.Sprintf(`attest jira testdata/file1 --artifact-type file --name foo 
								--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
								--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'foo' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "can attest jira against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd: fmt.Sprintf(`attest jira testdata/file1 --artifact-type file --name bar 
					--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "can attest jira against an artifact using --fingerprint",
			cmd: fmt.Sprintf(`attest jira --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
					--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'foo' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "can attest jira against a trail",
			cmd: fmt.Sprintf(`attest jira --name bar 
					--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "can attest jira against a trail when name is not found in the trail template",
			cmd: fmt.Sprintf(`attest jira --name additional 
				--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
				--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'additional' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "can attest jira against an artifact it is created using dot syntax in --name",
			cmd: fmt.Sprintf(`attest jira --name cli.foo 
					--jira-base-url https://kosli-test.atlassian.net  --jira-username tore@kosli.com
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'foo' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
	}

	for _, test := range tests {
		execJiraTestCase(test, suite)
	}
}

func execJiraTestCase(test cmdTestCase, suite *AttestJiraCommandTestSuite) {
	if test.additionalConfig != nil {
		branchName := test.additionalConfig.(jiraTestsAdditionalConfig).branchName
		if branchName != "" {
			err := testHelpers.CheckoutNewBranch(suite.workTree, branchName)
			require.NoError(suite.T(), err)
			defer testHelpers.CheckoutMaster(suite.workTree, suite.T())
		}
		msg := test.additionalConfig.(jiraTestsAdditionalConfig).commitMessage
		commitSha, err := testHelpers.CommitToRepo(suite.workTree, suite.fs, msg)
		require.NoError(suite.T(), err)

		test.cmd = test.cmd + " --commit " + commitSha
	}

	runTestCmd(suite.T(), []cmdTestCase{test})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestJiraCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestJiraCommandTestSuite))
}
