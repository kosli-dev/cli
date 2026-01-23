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
	commitWithPR        string
	commitWithNoPR      string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestGithubPRCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})

	suite.flowName = "attest-github-pr"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	suite.commitWithPR = "a72d2b5cfae42cb95700b3645de0c8ba3129a2ae"
	suite.commitWithNoPR = "13c900483c17b6ca5e0b26984ed74a6120838cad"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.T())
}

func (suite *AttestGithubPRCommandTestSuite) TestAttestGithubPRCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "01 fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest pullrequest github foo bar --commit %s %s", suite.commitWithPR, suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "02 fails when missing a required flags",
			cmd:       fmt.Sprintf("attest pullrequest github foo -t file --commit %s %s", suite.commitWithPR, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"github-org\", \"name\", \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "03 fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest pullrequest github testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit %s %s", suite.commitWithPR, suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "04 fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest pullrequest github --name foo --fingerprint xxxx --commit %s %s", suite.commitWithPR, suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "05 fails when --commit is provided as empty string",
			cmd: fmt.Sprintf(`attest pullrequest github --commit "" --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
			    --github-org kosli-dev --repository cli %s`, suite.defaultKosliArguments),
			golden: "Error: flag '--commit' is required, but empty string was provided\n",
		},
		{
			wantError: true,
			name:      "06 attesting against an artifact that does not exist fails",
			cmd: fmt.Sprintf(`attest pullrequest github --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nError: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-github-pr\" belonging to organization \"docs-cmd-test-user\"\n",
		},
		{
			name: "07 can attest github pr against an artifact using artifact name and --artifact-type",
			cmd: fmt.Sprintf(`attest pullrequest github testdata/file1 --artifact-type file --name foo 
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "08 can attest github pr against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest github testdata/file1 --artifact-type file --name bar 
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "09 can attest github pr against an artifact using --fingerprint",
			cmd: fmt.Sprintf(`attest pullrequest github --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "10 can attest github pr against a trail",
			cmd: fmt.Sprintf(`attest pullrequest github --name bar 
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "11 can attest github pr against a trail when name is not found in the trail template",
			cmd: fmt.Sprintf(`attest pullrequest github --name additional 
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name: "12 can attest github pr against an artifact it is created using dot syntax in --name",
			cmd: fmt.Sprintf(`attest pullrequest github --name cli.foo 
				--github-org kosli-dev --repository cli  --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "13 can attest github pr with external-url and external-fingerprint against a trail ",
			cmd: fmt.Sprintf(`attest pullrequest github --name bar 
				--external-url file=https://example.com/file --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "14 assert fails with non-zero exit code when commit has no PRs",
			cmd: fmt.Sprintf(`attest pullrequest github --name bar 
				--github-org kosli-dev --repository cli --commit %s --assert %s`, suite.commitWithNoPR, suite.defaultKosliArguments),
			goldenRegex: "found 0 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'bar' is reported to trail: test-123\nError: assert failed: no pull request found for the given commit: .*\n",
		},
		{
			name: "15 assert works and has zero exit code when commit has PR(s)",
			cmd: fmt.Sprintf(`attest pullrequest github --name bar 
				--github-org kosli-dev --repository cli --commit %s --assert %s`, suite.commitWithPR, suite.defaultKosliArguments),
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'bar' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "16 if there is a server error, this is output even when assert fails",
			cmd: fmt.Sprintf(`attest pullrequest github --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo 
				--github-org kosli-dev --repository cli --commit %s --assert %s`, suite.commitWithNoPR, suite.defaultKosliArguments),
			goldenRegex: "found 0 pull request\\(s\\) for commit: .*\nError: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-github-pr\" belonging to organization \"docs-cmd-test-user\"\nError: assert failed: no pull request found for the given commit: .*\n",
		},
		{
			name: "17 can attest github pr even if commit has no PR",
			cmd: fmt.Sprintf(`attest pullrequest github --name bar 
				--github-org kosli-dev --repository cli --commit %s %s`, suite.commitWithNoPR, suite.defaultKosliArguments),
			goldenRegex: "found 0 pull request\\(s\\) for commit: .*\ngithub pull request attestation 'bar' is reported to trail: test-123\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestGithubPRCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestGithubPRCommandTestSuite))
}
