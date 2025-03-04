package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AttestBitbucketPRCommandTestSuite struct {
	flowName            string
	trailName           string
	tmpDir              string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestBitbucketPRCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{"KOSLI_BITBUCKET_ACCESS_TOKEN"})

	suite.flowName = "attest-bitbucket-pr"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	var err error
	suite.tmpDir, err = os.MkdirTemp("", "testDir")
	require.NoError(suite.Suite.T(), err)
	_, err = testHelpers.CloneGitRepo("https://bitbucket.org/kosli-dev/cli-test.git", suite.tmpDir)
	require.NoError(suite.Suite.T(), err)

	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root %s --commit fd54040fc90e7e83f7b152619bfa18917b72c34f --host %s --org %s --api-token %s", suite.flowName, suite.trailName, suite.tmpDir, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.Suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.Suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.Suite.T())
}

func (suite *AttestBitbucketPRCommandTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tmpDir)
}

func (suite *AttestBitbucketPRCommandTestSuite) TestAttestBitbucketPRCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest pullrequest bitbucket foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flags",
			cmd:       fmt.Sprintf("attest pullrequest bitbucket foo -t file %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"bitbucket-workspace\", \"name\", \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both password and access token are provided",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket testdata/file1 --artifact-type file --name foo
				--bitbucket-workspace kosli-dev --repository cli-test --bitbucket-password xxx %s`, suite.defaultKosliArguments),
			golden: "Error: only one of --bitbucket-password, --bitbucket-access-token is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest pullrequest bitbucket testdata/file1 --fingerprint xxxx --artifact-type file --name bar   %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest pullrequest bitbucket --name foo --fingerprint xxxx  %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest pullrequest bitbucket [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
				--bitbucket-workspace kosli-dev --repository cli-test   %s`, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nError: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-bitbucket-pr\" belonging to organization \"docs-cmd-test-user\"\n",
		},
		{
			name: "can attest bitbucket pr against an artifact using artifact name and --artifact-type",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket testdata/file1 --artifact-type file --name foo
				--bitbucket-workspace kosli-dev --repository cli-test  %s`, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest bitbucket pr against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket testdata/file1 --artifact-type file --name bar
				--bitbucket-workspace kosli-dev --repository cli-test  %s`, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "can attest bitbucket pr against an artifact using --fingerprint",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo
				--bitbucket-workspace kosli-dev --repository cli-test  %s`, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest bitbucket pr against a trail",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket --name bar
				--bitbucket-workspace kosli-dev --repository cli-test  %s`, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "can attest bitbucket pr against a trail when name is not found in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket --name additional
				--bitbucket-workspace kosli-dev --repository cli-test  %s`, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name: "can attest bitbucket pr against an artifact it is created using dot syntax in --name",
			cmd: fmt.Sprintf(`attest pullrequest bitbucket --name cli.foo
				--bitbucket-workspace kosli-dev --repository cli-test  %s`, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request attestation 'foo' is reported to trail: test-123\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestBitbucketPRCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestBitbucketPRCommandTestSuite))
}
