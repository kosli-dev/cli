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
			name:   "works when --name does not match artifact name in the template (extra artifact)",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --artifact-type file --name bar --commit HEAD --build-url example.com --commit-url example.com  %s", suite.defaultKosliArguments),
			golden: "artifact file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
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
		{
			name:   "can attest an artifact with external urls",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url example.com --commit-url example.com --external-url jira=https://jira.kosli.com  %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest an artifact with external urls and fingerprints",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url example.com --commit-url example.com --external-url file=https://kosli.com/file --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9  %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			wantError: true,
			name:      "fails when --external-fingerprint has more items than external urls",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url example.com --commit-url example.com --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9  %s", suite.defaultKosliArguments),
			golden:    "Error: --external-fingerprints have labels that don't have a URL in --external-url\n",
		},
		{
			wantError: true,
			name:      "fails (from server) when --external-fingerprint has invalid fingerprint",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url example.com --commit-url example.com --external-url file=https://example.com --external-fingerprint file=7509e5bda0  %s", suite.defaultKosliArguments),
			golden:    "Error: Input payload validation failed: map[external_urls.file.fingerprint:'7509e5bda0' does not match '^[a-f0-9]{64}$']\n",
		},
		{
			name:   "can attest with annotations against a trail",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url example.com --commit-url example.com --annotate foo=bar --annotate baz=\"data with spaces\" %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			wantError: true,
			name:      "fails when annotation is not valid",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url example.com --commit-url example.com --annotate foo.baz=bar %s", suite.defaultKosliArguments),
			golden:    "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.baz'. Key can only contain [A-Za-z0-9_].\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestArtifactCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestArtifactCommandTestSuite))
}
