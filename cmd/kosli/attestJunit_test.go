package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AttestJunitCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestJunitCommandTestSuite) SetupTest() {
	suite.flowName = "attest-junit"
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

func (suite *AttestJunitCommandTestSuite) TestAttestJunitCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest junit foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flags",
			cmd:       fmt.Sprintf("attest junit foo %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest junit testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest junit --name foo --fingerprint xxxx --commit HEAD --origin-url example.com %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest junit [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest junit --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --results-dir testdata  %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-junit\" belonging to organization \"docs-cmd-test-user\"\n",
		},
		{
			name:   "can attest junit against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest junit testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url example.com --results-dir testdata %s", suite.defaultKosliArguments),
			golden: "junit attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest junit against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest junit testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url example.com --results-dir testdata  %s", suite.defaultKosliArguments),
			golden: "junit attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest junit against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest junit --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --results-dir testdata %s", suite.defaultKosliArguments),
			golden: "junit attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest junit against a trail",
			cmd:    fmt.Sprintf("attest junit --name bar --commit HEAD --origin-url example.com --results-dir testdata %s", suite.defaultKosliArguments),
			golden: "junit attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest junit against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest junit --name additional --commit HEAD --origin-url example.com --results-dir testdata %s", suite.defaultKosliArguments),
			golden: "junit attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest junit against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest junit --name cli.foo --commit HEAD --origin-url example.com --results-dir testdata %s", suite.defaultKosliArguments),
			golden: "junit attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest junit with external-url and external-fingerprint against a trail",
			cmd: fmt.Sprintf(`attest junit --name bar --commit HEAD --origin-url example.com --results-dir testdata
					 --external-url file=https://foo.com/file 
					 --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 %s`, suite.defaultKosliArguments),
			golden: "junit attestation 'bar' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "fails when external-fingerprint has more items than external-url",
			cmd: fmt.Sprintf(`attest junit --name bar --commit HEAD --origin-url example.com --results-dir testdata
					 --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 %s`, suite.defaultKosliArguments),
			golden: "Error: --external-fingerprints have labels that don't have a URL in --external-url\n",
		},
		{
			name: "can attest with annotations against a trail",
			cmd: fmt.Sprintf(`attest junit --name bar --commit HEAD --origin-url example.com --results-dir testdata
					 --annotate foo=bar --annotate baz=qux %s`, suite.defaultKosliArguments),
			golden: "junit attestation 'bar' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "fails when annotation is not valid",
			cmd: fmt.Sprintf(`attest junit --name bar --commit HEAD --origin-url example.com --results-dir testdata
					       --annotate foo.bar=bar %s`, suite.defaultKosliArguments),
			golden: "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.bar'. Key can only contain [A-Za-z0-9_].\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestJunitCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestJunitCommandTestSuite))
}
