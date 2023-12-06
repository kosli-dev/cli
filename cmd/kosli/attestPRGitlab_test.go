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
type AttestGitlabPRCommandTestSuite struct {
	flowName            string
	trailName           string
	tmpDir              string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestGitlabPRCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITLAB_TOKEN"})

	suite.flowName = "attest-gitlab-pr"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	var err error
	suite.tmpDir, err = os.MkdirTemp("", "testDir")
	require.NoError(suite.T(), err)
	_, err = testHelpers.CloneGitRepo("https://gitlab.com/ewelinawilkosz/merkely-gitlab-demo.git", suite.tmpDir)
	require.NoError(suite.T(), err)

	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root %s --commit e6510880aecdc05d79104d937e1adb572bd91911 --host %s --org %s --api-token %s", suite.flowName, suite.trailName, suite.tmpDir, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.T())
}

func (suite *AttestGitlabPRCommandTestSuite) TearDownSuite() {
	os.RemoveAll(suite.tmpDir)
}

func (suite *AttestGitlabPRCommandTestSuite) TestAttestGitlabPRCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest pullrequest gitlab foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flags",
			cmd:       fmt.Sprintf("attest pullrequest gitlab foo %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"gitlab-org\", \"name\", \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest pullrequest gitlab testdata/file1 --fingerprint xxxx --artifact-type file --name bar   %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest pullrequest gitlab --name foo --fingerprint xxxx  %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest pullrequest gitlab [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd: fmt.Sprintf(`attest pullrequest gitlab --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
						--gitlab-org ewelinawilkosz --repository merkely-gitlab-demo   %s`, suite.defaultKosliArguments),
			golden: "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-gitlab-pr' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name: "can attest gitlab pr against an artifact using artifact name and --artifact-type",
			cmd: fmt.Sprintf(`attest pullrequest gitlab testdata/file1 --artifact-type file --name foo 
					--gitlab-org ewelinawilkosz --repository merkely-gitlab-demo  %s`, suite.defaultKosliArguments),
			golden: "gitlab merge request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest gitlab pr against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest gitlab testdata/file1 --artifact-type file --name bar 
					--gitlab-org ewelinawilkosz --repository merkely-gitlab-demo  %s`, suite.defaultKosliArguments),
			golden: "gitlab merge request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "can attest gitlab pr against an artifact using --fingerprint",
			cmd: fmt.Sprintf(`attest pullrequest gitlab --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
					--gitlab-org ewelinawilkosz --repository merkely-gitlab-demo  %s`, suite.defaultKosliArguments),
			golden: "gitlab merge request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest gitlab pr against a trail",
			cmd: fmt.Sprintf(`attest pullrequest gitlab --name bar 
				--gitlab-org ewelinawilkosz --repository merkely-gitlab-demo  %s`, suite.defaultKosliArguments),
			golden: "gitlab merge request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "can attest gitlab pr against a trail when name is not found in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest gitlab --name additional 
					--gitlab-org ewelinawilkosz --repository merkely-gitlab-demo  %s`, suite.defaultKosliArguments),
			golden: "gitlab merge request attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name: "can attest gitlab pr against an artifact it is created using dot syntax in --name",
			cmd: fmt.Sprintf(`attest pullrequest gitlab --name cli.foo 
				--gitlab-org ewelinawilkosz --repository merkely-gitlab-demo  %s`, suite.defaultKosliArguments),
			golden: "gitlab merge request attestation 'foo' is reported to trail: test-123\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestGitlabPRCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestGitlabPRCommandTestSuite))
}
