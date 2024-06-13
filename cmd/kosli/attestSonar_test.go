package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

type AttestSonarCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestSonarCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_SONAR_API_TOKEN"})
	suite.flowName = "attest-sonar"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
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

func (suite *AttestSonarCommandTestSuite) TestAttestSonarCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest sonar foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flags",
			cmd:       fmt.Sprintf("attest sonar foo %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\", \"sonar-project-key\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest sonar testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest sonar --name foo-s --fingerprint xxxx --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ  %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-sonar' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ  %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ  %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest sonar --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ  %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail",
			cmd:    fmt.Sprintf("attest sonar --name bar --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest sonar --name additional --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarCloud for a non-existent project gives error",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key cyberdojo_differ  %s", suite.defaultKosliArguments),
			golden:    "Error: No data retrieved from Sonarcloud/Sonarqube - check your project key and API token are correct\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestSonarCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSonarCommandTestSuite))
}
