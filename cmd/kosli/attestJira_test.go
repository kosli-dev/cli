package main

import (
	"fmt"
	"os"
	"testing"

	billy "github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
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
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{"KOSLI_JIRA_API_TOKEN", "KOSLI_JIRA_USERNAME"})
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
	require.NoError(suite.Suite.T(), err)
	_, suite.workTree, suite.fs, err = testHelpers.InitializeGitRepo(suite.tmpDir)
	require.NoError(suite.Suite.T(), err)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.Suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.Suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.Suite.T())
}

func (suite *AttestJiraCommandTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tmpDir)
}

func (suite *AttestJiraCommandTestSuite) TestAttestJiraCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "01 fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest jira foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "02 fails when missing required flags",
			cmd:       fmt.Sprintf("attest jira foo -t file %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"commit\", \"jira-base-url\", \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "03 fails when missing --commit flag",
			cmd: fmt.Sprintf(`attest jira foo -t file --name bar
							--jira-base-url https://kosli-test.atlassian.net
							--jira-api-token secret
							%s`, suite.defaultKosliArguments),
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "04 fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest jira testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "05 fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest jira --name foo --fingerprint xxxx --commit HEAD --origin-url http://www..com %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest jira [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "06 attesting against an artifact that does not exist fails",
			cmd: fmt.Sprintf(`attest jira --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
								--name foo
								--repo-root %s
								--jira-base-url https://kosli-test.atlassian.net
								%s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-jira\" belonging to organization \"docs-cmd-test-user\"\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "test commit",
			},
		},
		{
			wantError: true,
			name:      "07 assert for non-existing Jira issue gives an error",
			cmd: fmt.Sprintf(`attest jira --name jira-validation
					--jira-base-url https://kosli-test.atlassian.net
					--repo-root %s
					--assert %s`, suite.tmpDir, suite.defaultKosliArguments),
			goldenRegex: "Error: missing Jira issues from references found in commit message or branch name.*",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "SAMI-1 test commit",
			},
		},
		{
			name: "08 can attest jira against an artifact using artifact name and --artifact-type",
			cmd: fmt.Sprintf(`attest jira testdata/file1 --artifact-type file --name foo 
								--jira-base-url https://kosli-test.atlassian.net
								--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'foo' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "09 can attest jira when the issue doesn't exist",
			cmd: fmt.Sprintf(`attest jira testdata/file1 --artifact-type file --name foo
								--jira-base-url https://kosli-test.atlassian.net
								--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'foo' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-999 test commit",
			},
		},
		{
			name: "10 can attest jira against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd: fmt.Sprintf(`attest jira testdata/file1 --artifact-type file --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "11 can attest jira against an artifact using --fingerprint",
			cmd: fmt.Sprintf(`attest jira --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo
					--jira-base-url https://kosli-test.atlassian.net
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'foo' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "12 can attest jira against a trail",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "13 can attest jira against a trail with summary and description from jira issue fields",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-issue-fields "summary,description"
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "14 can attest jira against a trail when name is not found in the trail template",
			cmd: fmt.Sprintf(`attest jira --name additional
				--jira-base-url https://kosli-test.atlassian.net
				--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'additional' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "15 can attest jira against an artifact it is created using dot syntax in --name",
			cmd: fmt.Sprintf(`attest jira --name cli.foo
					--jira-base-url https://kosli-test.atlassian.net
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'foo' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "16 can attest jira against a trail with attachment and external-url",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--attachments testdata/file1 --external-url foo=https://foo.com --external-url bar=https://bar.com
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "17 can attest jira against a trail with external-url and external-fingerprint",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--external-url foo=https://foo.com --external-fingerprint foo=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			wantError: true,
			name:      "18 fails when external-url and external-fingerprint labels don't match",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--external-url foo=https://foo.com --external-fingerprint bar=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "Error: bar in --external-fingerprint does not match any labels in --external-url\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "19 can specify the jira project key",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-project-key EX
					--jira-project-key ABC
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "20 can specify lower case and underscore jira project key",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-project-key low
					--jira-project-key A99
					--jira-project-key A_99
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "low-1 test commit",
			},
		},
		{
			wantError: true,
			name:      "21 fails with an invalid Jira project key specified",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-project-key 1AB
					--jira-project-key AB-44
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "Error: Invalid Jira project keys: 1AB, AB-44\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			wantError: true,
			name:      "22 if no matching issue exists, fails with a non-zero exit code",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-project-key ABC
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "Error: no Jira references are found in commit message or branch name\n",
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
			require.NoError(suite.Suite.T(), err)
			defer testHelpers.CheckoutMaster(suite.workTree, suite.Suite.T())
		}
		msg := test.additionalConfig.(jiraTestsAdditionalConfig).commitMessage
		commitSha, err := testHelpers.CommitToRepo(suite.workTree, suite.fs, msg)
		require.NoError(suite.Suite.T(), err)

		test.cmd = test.cmd + " --commit " + commitSha
	}

	runTestCmd(suite.Suite.T(), []cmdTestCase{test})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestJiraCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestJiraCommandTestSuite))
}
