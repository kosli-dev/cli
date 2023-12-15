package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AttestArtifactCommandTestSuite struct {
	flowName  string
	trailName string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestArtifactCommandTestSuite) SetupTest() {
	suite.flowName = "attest-artifact"
	suite.trailName = "test-123"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
}

func (suite *AttestArtifactCommandTestSuite) TestAttestArtifactCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest artifact foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flag",
			cmd:       fmt.Sprintf("attest artifact foo --artifact-type file --name bar --commit HEAD --build-url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"commit-url\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is invalid sha256 digest",
			cmd:       fmt.Sprintf("attest artifact foo --fingerprint xxxx --name bar --commit HEAD --build-url example.com --commit-url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]\n",
		},
		{
			wantError: true,
			name:      "fails when --name does not match artifact name in the template",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --artifact-type file --name bar --commit HEAD --build-url example.com --commit-url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact 'bar' does not exist in trail template 'test-123'.\nAvailable artifacts: cli\n",
		},
		{
			name:   "can attest a file artifact",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --artifact-type file --name cli --commit HEAD --build-url example.com --commit-url example.com  %s", suite.defaultKosliArguments),
			golden: "artifact file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest an artifact with --fingerprint",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url example.com --commit-url example.com  %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestArtifactCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestArtifactCommandTestSuite))
}
