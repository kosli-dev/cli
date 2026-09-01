package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	billy "github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/jira"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type jiraTestsAdditionalConfig struct {
	commitMessage string
	branchName    string
}

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
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_JIRA_API_TOKEN", "KOSLI_JIRA_USERNAME"})
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
	err := os.RemoveAll(suite.tmpDir)
	require.NoError(suite.T(), err, "failed to remove temp dir %s", suite.tmpDir)
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
			// cobra splits the list with encoding/csv, which does not trim, so without
			// normalisation the " EX" fragment fails validation and the command never
			// reaches the matcher. --assert carries the rest of the path end to end.
			name: "20b can specify jira project keys as a comma-separated list with spaces",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-project-key "ABC, EX"
					--repo-root %s
					--assert %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			// --assert so this covers the whole path rather than validation alone: the
			// success line above is printed either way, but a padded key that reached the
			// matcher unusable would find no references and fail the assert.
			name: "20c a jira project key padded with spaces is accepted",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-project-key " EX "
					--repo-root %s
					--assert %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
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
			golden: "Error: invalid Jira project keys: [\"1AB\" \"AB-44\"]\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			wantError: true,
			name:      "22 if no matching issue exists, assert fails with a non-zero exit code",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-project-key ABC
					--repo-root %s 
					--assert %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\nError: no Jira references are found in commit message or branch name\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			name: "23 assert works and exits with zero code if issue exists",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--repo-root %s 
					--assert %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			wantError: true,
			name:      "24 if there is a server error, this is output even when assert fails due to no matching issue",
			cmd: fmt.Sprintf(`attest jira --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
								--name foo
								--repo-root %s
								--jira-project-key ABC
								--jira-base-url https://kosli-test.atlassian.net
								--assert
								%s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-jira\" belonging to organization \"docs-cmd-test-user\"\nError: no Jira references are found in commit message or branch name\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "EX-1 test commit",
			},
		},
		{
			wantError: true,
			name:      "25 if there is a server error, this is output even when assert fails due to non-existing issue",
			cmd: fmt.Sprintf(`attest jira --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
								--name foo
								--repo-root %s
								--jira-base-url https://kosli-test.atlassian.net
								--assert
								%s`, suite.tmpDir, suite.defaultKosliArguments),
			goldenRegex: "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-jira\" belonging to organization \"docs-cmd-test-user\"\nError: missing Jira issues from references found in commit message or branch name.*",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "SAMI-1 test commit",
			},
		},
		{
			wantError: true,
			name:      "26 fails when --name has invalid dot format",
			cmd:       fmt.Sprintf("attest jira --name .foo --commit HEAD --jira-base-url https://kosli-test.atlassian.net %s", suite.defaultKosliArguments),
			golden:    "Error: failed to parse attestation name: invalid attestation name format: .foo\n",
		},
		{
			name: "27 can attest jira using --jira-trailer to extract issue key from commit trailer",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-trailer Jira
					--assert
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "fix: some change\n\nJira: EX-1\nOna-Environment-Id: ONA-999",
			},
		},
		{
			name: "28 --jira-trailer with no matching trailer produces no issue IDs (non-compliant but reported)",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-trailer Jira
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "fix: some change with no jira trailer",
			},
		},
		{
			wantError: true,
			name:      "29 --jira-trailer with --assert fails when trailer is absent",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-trailer Jira
					--assert
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\nError: no Jira references are found in trailer 'Jira'\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "fix: some change with no jira trailer",
			},
		},
		{
			wantError: true,
			name:      "30 --jira-trailer and --jira-secondary-source are mutually exclusive",
			cmd:       fmt.Sprintf("attest jira --name bar --jira-base-url https://kosli-test.atlassian.net --jira-trailer Jira --jira-secondary-source foo --commit HEAD --repo-root %s %s", suite.tmpDir, suite.defaultKosliArguments),
			golden:    "Error: only one of --jira-trailer, --jira-secondary-source is allowed\n",
		},
		{
			wantError: true,
			name:      "31 --jira-trailer with a blank-ish value is rejected",
			cmd:       fmt.Sprintf("attest jira --name bar --jira-base-url https://kosli-test.atlassian.net --jira-trailer : --commit HEAD --repo-root %s %s", suite.tmpDir, suite.defaultKosliArguments),
			golden:    "Error: flag '--jira-trailer' was given an empty value\n",
		},
		{
			wantError: true,
			name:      "32 --jira-trailer with an internal colon is rejected",
			cmd:       fmt.Sprintf("attest jira --name bar --jira-base-url https://kosli-test.atlassian.net --jira-trailer A:B --commit HEAD --repo-root %s %s", suite.tmpDir, suite.defaultKosliArguments),
			golden:    "Error: flag '--jira-trailer' is not a valid trailer key: trailer keys cannot contain colons or spaces\n",
		},
		{
			name: "33 --jira-trailer warns when trailer key is present but value is empty",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-trailer Jira
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "[warning] trailer 'Jira' was found but contained no valid Jira issue keys: []\njira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "fix: some change\n\nJira:",
			},
		},
		{
			name: "34 --jira-trailer warns when trailer value is present but not a valid Jira key",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-trailer Jira
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "[warning] trailer 'Jira' was found but contained no valid Jira issue keys: [not-a-key]\njira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "fix: some change\n\nJira: not-a-key",
			},
		},
		{
			name: "35 --ignore-branch-match warns that it has no effect in trailer mode",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-trailer Jira
					--ignore-branch-match
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "[warning] --ignore-branch-match has no effect when --jira-trailer is set\njira attestation 'bar' is reported to trail: test-123\n",
			additionalConfig: jiraTestsAdditionalConfig{
				commitMessage: "fix: some change\n\nJira: EX-1",
			},
		},
		{
			wantError: true,
			name:      "36 --jira-trailer does not scan branch name even when branch contains a Jira key",
			cmd: fmt.Sprintf(`attest jira --name bar
					--jira-base-url https://kosli-test.atlassian.net
					--jira-trailer Jira
					--assert
					--repo-root %s %s`, suite.tmpDir, suite.defaultKosliArguments),
			golden: "jira attestation 'bar' is reported to trail: test-123\nError: no Jira references are found in trailer 'Jira'\n",
			additionalConfig: jiraTestsAdditionalConfig{
				branchName:    "EX-1-some-feature",
				commitMessage: "fix: some change with no jira trailer",
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

func TestJiraAssertHeadline(t *testing.T) {
	for _, tc := range []struct {
		name        string
		missing     int
		unconfirmed int
		want        string
	}{
		{
			// every lookup was answered, so the long-standing wording is the accurate one
			name:    "all missing",
			missing: 2,
			want:    "missing Jira issues",
		},
		{
			name:        "none answered",
			unconfirmed: 2,
			want:        "unconfirmed Jira issues",
		},
		{
			name:        "some missing, some unanswered",
			missing:     1,
			unconfirmed: 1,
			want:        "missing or unconfirmed Jira issues",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, jiraAssertHeadline(tc.missing, tc.unconfirmed))
		})
	}
}

func TestJiraUnconfirmedWarning(t *testing.T) {
	// three keys sharing one cause, as an expired token produces: stated once, not per key
	warning := jiraUnconfirmedWarning(
		[]string{"EX-1", "EX-2", "EX-3"},
		[]string{"Jira did not accept the username user@example.com and API token"})

	require.Equal(t, "could not confirm EX-1, EX-2, EX-3 in Jira: Jira did not accept the username user@example.com and API token."+
		" Unconfirmed issues are counted as not found in the attestation.", warning)
	require.Equal(t, 1, strings.Count(warning, "did not accept"))
}

// TestAttestJiraWarnsOnUnconfirmedIssue covers the wiring from an unanswered lookup to the
// warning the user sees. It runs against a fake Jira in dry-run mode, so it needs neither
// Jira credentials nor a Kosli server.
func TestAttestJiraWarnsOnUnconfirmedIssue(t *testing.T) {
	// a 404 that says the request was handled as anonymous, which is what Jira answers
	// once the API token has expired
	jiraServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-AUSERNAME", "anonymous")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist or you do not have permission to see it."]}`))
	}))
	defer jiraServer.Close()

	tmpDir, err := os.MkdirTemp("", "testDir")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir), "failed to remove temp dir %s", tmpDir)
	}()

	_, workTree, fs, err := testHelpers.InitializeGitRepo(tmpDir)
	require.NoError(t, err)
	commitSha, err := testHelpers.CommitToRepo(workTree, fs, "EX-1 fix the thing")
	require.NoError(t, err)

	_, _, _, errOut, err := executeCommandC(fmt.Sprintf(`attest jira --name foo --flow flow --trail trail
		--org org --api-token DRY_RUN --repo-root %s --commit %s
		--jira-base-url %s --jira-username user@example.com --jira-api-token secret`,
		tmpDir, commitSha, jiraServer.URL))
	require.NoError(t, err)
	require.Contains(t, errOut, "could not confirm EX-1 in Jira")
	require.Contains(t, errOut, "user@example.com")
	require.NotContains(t, errOut, "secret", "the API token must not appear in the warning")
}

// TestAttestJiraAssertNamesUnconfirmedIssues covers the --assert failure as a whole - the
// headline, the per-issue lines and the reasons are assembled in run(), and the assert block
// is guarded by !global.DryRun, so the dry-run test above never reaches it.
func TestAttestJiraAssertNamesUnconfirmedIssues(t *testing.T) {
	// the 404 an expired token produces
	jiraServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-AUSERNAME", "anonymous")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist or you do not have permission to see it."]}`))
	}))
	defer jiraServer.Close()

	// accepts the attestation, so the run gets as far as the assert block
	kosliServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer kosliServer.Close()

	tmpDir, err := os.MkdirTemp("", "testDir")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir), "failed to remove temp dir %s", tmpDir)
	}()

	_, workTree, fs, err := testHelpers.InitializeGitRepo(tmpDir)
	require.NoError(t, err)
	commitSha, err := testHelpers.CommitToRepo(workTree, fs, "EX-1 fix the thing")
	require.NoError(t, err)

	_, _, _, errOut, err := executeCommandC(fmt.Sprintf(`attest jira --name foo --flow flow --trail trail
		--org org --host %s --api-token %s --assert --repo-root %s --commit %s
		--jira-base-url %s --jira-username user@example.com --jira-api-token secret`,
		kosliServer.URL,
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		tmpDir, commitSha, jiraServer.URL))

	require.Error(t, err)
	require.Contains(t, errOut, "Error: unconfirmed Jira issues from references found in commit message or branch name")
	require.Contains(t, errOut, "\n\tEX-1: issue not confirmed")
	require.Contains(t, errOut, "\n\treason: Jira did not accept the username user@example.com and API token")
	// the count a swapped argument would put here instead
	require.NotContains(t, errOut, "missing Jira issues")
}

func TestJiraIssueLogLine(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result *jira.JiraIssueInfo
		want   string
	}{
		{
			name:   "a found issue",
			result: &jira.JiraIssueInfo{LookupStatus: jira.IssueFound, IssueExists: true},
			want:   "issue found",
		},
		{
			name:   "a missing issue",
			result: &jira.JiraIssueInfo{LookupStatus: jira.IssueMissing},
			want:   "issue not found",
		},
		{
			// the reason is reported once for the run, so it is not repeated per line
			name:   "an unanswered lookup is not claimed to be missing",
			result: &jira.JiraIssueInfo{LookupStatus: jira.LookupUnverified, LookupReason: "no response from Jira at https://example.atlassian.net: timeout"},
			want:   "issue not confirmed",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, jiraIssueLogLine(tc.result))
		})
	}
}

func TestJiraSearchText(t *testing.T) {
	commitInfo := &gitview.CommitInfo{
		BasicCommitInfo: gitview.BasicCommitInfo{
			Message: "EX-1 fix the thing",
			Branch:  "bugfix/EX-2",
		},
	}
	for _, tc := range []struct {
		name              string
		secondarySource   string
		ignoreBranchMatch bool
		want              string
	}{
		{
			name: "the commit message and the branch name are searched",
			want: "EX-1 fix the thing\nbugfix/EX-2",
		},
		{
			name:              "the branch name is skipped when ignoreBranchMatch is set",
			ignoreBranchMatch: true,
			want:              "EX-1 fix the thing",
		},
		{
			name:            "a secondary source is appended",
			secondarySource: "EX-3",
			want:            "EX-1 fix the thing\nbugfix/EX-2\nEX-3",
		},
		{
			name:              "a secondary source is appended without the branch name",
			secondarySource:   "EX-3",
			ignoreBranchMatch: true,
			want:              "EX-1 fix the thing\nEX-3",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, jiraSearchText(commitInfo, tc.secondarySource, tc.ignoreBranchMatch))
		})
	}
}

func TestNormaliseJiraProjectKeys(t *testing.T) {
	for _, tc := range []struct {
		name        string
		projectKeys []string
		want        []string
	}{
		{
			// cobra splits on the comma without trimming, so this is what
			// --jira-project-key "ABC, DEF" actually delivers
			name:        "a space after the comma is trimmed",
			projectKeys: []string{"ABC", " DEF"},
			want:        []string{"ABC", "DEF"},
		},
		{
			name:        "space on both sides is trimmed",
			projectKeys: []string{" EX "},
			want:        []string{"EX"},
		},
		{
			name:        "tabs and newlines are trimmed",
			projectKeys: []string{"\tEX", "ABC\n"},
			want:        []string{"EX", "ABC"},
		},
		{
			name:        "keys that need no trimming are left alone",
			projectKeys: []string{"ABC", "low", "A_99"},
			want:        []string{"ABC", "low", "A_99"},
		},
		{
			// a trailing comma yields an empty fragment, which stays empty so that
			// validateJiraProjectKeys still rejects it
			name:        "an empty key stays empty",
			projectKeys: []string{"ABC", ""},
			want:        []string{"ABC", ""},
		},
		{
			name:        "a whitespace-only key becomes empty and is still rejected downstream",
			projectKeys: []string{"ABC", " "},
			want:        []string{"ABC", ""},
		},
		{
			name:        "no keys is left alone",
			projectKeys: []string{},
			want:        []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := &attestJiraOptions{projectKeys: tc.projectKeys}
			o.normaliseJiraProjectKeys()
			require.Equal(t, tc.want, o.projectKeys)
		})
	}
}

// TestValidateJiraProjectKeys pins that validation normalises the keys itself, rather than
// relying on its caller to have done it, and that a key which is only whitespace is named in
// the error instead of rendering as nothing.
func TestValidateJiraProjectKeys(t *testing.T) {
	for _, tc := range []struct {
		name        string
		projectKeys []string
		wantErr     string
		wantKeys    []string
	}{
		{
			name:        "a padded key is accepted and left trimmed",
			projectKeys: []string{" EX "},
			wantKeys:    []string{"EX"},
		},
		{
			name:        "a padded comma-separated list is accepted",
			projectKeys: []string{"ABC", " EX"},
			wantKeys:    []string{"ABC", "EX"},
		},
		{
			name:        "an invalid key is still rejected, and quoted",
			projectKeys: []string{"1AB", "AB-44"},
			wantErr:     `invalid Jira project keys: ["1AB" "AB-44"]`,
		},
		{
			name:        "a whitespace-only key is named in the error rather than rendering as nothing",
			projectKeys: []string{"ABC", " "},
			wantErr:     `invalid Jira project keys: [""]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := &attestJiraOptions{projectKeys: tc.projectKeys}
			err := o.validateJiraProjectKeys()
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantKeys, o.projectKeys)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestJiraCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestJiraCommandTestSuite))
}
