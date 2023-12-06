package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AttestGenericCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestGenericCommandTestSuite) SetupTest() {
	suite.flowName = "attest-generic"
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

func (suite *AttestGenericCommandTestSuite) TestAttestGenericCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest generic foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flag",
			cmd:       fmt.Sprintf("attest generic foo %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest generic testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest generic --name foo --fingerprint xxxx --commit HEAD --url example.com %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest generic [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest generic --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-generic' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:   "can attest generic against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest generic testdata/file1 --artifact-type file --name foo --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden: "generic attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest generic against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest generic testdata/file1 --artifact-type file --name bar --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden: "generic attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest generic against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest generic --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden: "generic attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest generic against a trail",
			cmd:    fmt.Sprintf("attest generic --name bar --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden: "generic attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest generic against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest generic --name additional --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden: "generic attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest generic against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest generic --name cli.foo --commit HEAD --url example.com  %s", suite.defaultKosliArguments),
			golden: "generic attestation 'foo' is reported to trail: test-123\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestGenericCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestGenericCommandTestSuite))
}
