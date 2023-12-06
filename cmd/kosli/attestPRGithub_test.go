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
type AttestGithubPRCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestGithubPRCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})

	suite.flowName = "attest-github-pr"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --commit a72d2b5cfae42cb95700b3645de0c8ba3129a2ae --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.T())
}

func (suite *AttestGithubPRCommandTestSuite) TestAttestGithubPRCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest pullrequest github foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flags",
			cmd:       fmt.Sprintf("attest pullrequest github foo %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"github-org\", \"name\", \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest pullrequest github testdata/file1 --fingerprint xxxx --artifact-type file --name bar  %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest pullrequest github --name foo --fingerprint xxxx %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd: fmt.Sprintf(`attest pullrequest github --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
				--github-org kosli-dev --repository cli   %s`, suite.defaultKosliArguments),
			golden: "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-github-pr' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name: "can attest github pr against an artifact using artifact name and --artifact-type",
			cmd: fmt.Sprintf(`attest pullrequest github testdata/file1 --artifact-type file --name foo 
				--github-org kosli-dev --repository cli  %s`, suite.defaultKosliArguments),
			golden: "github pull request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest github pr against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest github testdata/file1 --artifact-type file --name bar 
				--github-org kosli-dev --repository cli  %s`, suite.defaultKosliArguments),
			golden: "github pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "can attest github pr against an artifact using --fingerprint",
			cmd: fmt.Sprintf(`attest pullrequest github --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
				--github-org kosli-dev --repository cli  %s`, suite.defaultKosliArguments),
			golden: "github pull request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest github pr against a trail",
			cmd: fmt.Sprintf(`attest pullrequest github --name bar 
				--github-org kosli-dev --repository cli  %s`, suite.defaultKosliArguments),
			golden: "github pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "can attest github pr against a trail when name is not found in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest github --name additional 
				--github-org kosli-dev --repository cli  %s`, suite.defaultKosliArguments),
			golden: "github pull request attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name: "can attest github pr against an artifact it is created using dot syntax in --name",
			cmd: fmt.Sprintf(`attest pullrequest github --name cli.foo 
				--github-org kosli-dev --repository cli  %s`, suite.defaultKosliArguments),
			golden: "github pull request attestation 'foo' is reported to trail: test-123\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestGithubPRCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestGithubPRCommandTestSuite))
}
