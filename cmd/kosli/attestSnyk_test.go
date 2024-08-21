package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AttestSnykCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestSnykCommandTestSuite) SetupTest() {
	suite.flowName = "attest-snyk"
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

func (suite *AttestSnykCommandTestSuite) TestAttestSnykCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest snyk foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flags",
			cmd:       fmt.Sprintf("attest snyk foo -t file %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\", \"scan-results\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest snyk testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest snyk --name foo --fingerprint xxxx --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest snyk [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest snyk --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --scan-results testdata/snyk_sarif.json %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-snyk\" belonging to organization \"docs-cmd-test-user\"\n",
		},
		{
			wantError: true,
			name:      "fails when --snyk-results is missing",
			cmd:       fmt.Sprintf("attest snyk testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"scan-results\" not set\n",
		},
		{
			name:   "can attest snyk against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest snyk testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url http://www.example.com --scan-results testdata/snyk_sarif.json  %s", suite.defaultKosliArguments),
			golden: "snyk attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest snyk against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest snyk testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com --scan-results testdata/snyk_sarif.json  %s", suite.defaultKosliArguments),
			golden: "snyk attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest snyk against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest snyk --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --scan-results testdata/snyk_sarif.json %s", suite.defaultKosliArguments),
			golden: "snyk attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest snyk against a trail",
			cmd:    fmt.Sprintf("attest snyk --name bar --commit HEAD --origin-url http://www.example.com --scan-results testdata/snyk_sarif.json %s", suite.defaultKosliArguments),
			golden: "snyk attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest snyk against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest snyk --name additional --commit HEAD --origin-url http://www.example.com --scan-results testdata/snyk_sarif.json %s", suite.defaultKosliArguments),
			golden: "snyk attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest snyk against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest snyk --name cli.foo --commit HEAD --origin-url http://www.example.com --scan-results testdata/snyk_sarif.json %s", suite.defaultKosliArguments),
			golden: "snyk attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name: "can attest snyk against with external-url and external-fingerprint a trail",
			cmd: fmt.Sprintf(`attest snyk --name bar --commit HEAD --origin-url http://www.example.com
				--external-url file=https://http://www.example.com/file  --external-url other=https://other.com
				--external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 
				--scan-results testdata/snyk_sarif.json %s`, suite.defaultKosliArguments),
			golden: "snyk attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name: "can attest with annotations against a trail",
			cmd: fmt.Sprintf(`attest snyk --name bar --commit HEAD --origin-url http://www.example.com
				--annotate foo=bar --annotate baz=qux
				--scan-results testdata/snyk_sarif.json %s`, suite.defaultKosliArguments),
			golden: "snyk attestation 'bar' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "fails when annotation is not valid",
			cmd: fmt.Sprintf(`attest snyk --name bar --commit HEAD --origin-url http://www.example.com
				--annotate foo.baz=bar
				--scan-results testdata/snyk_sarif.json %s`, suite.defaultKosliArguments),
			golden: "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.baz'. Key can only contain [A-Za-z0-9_].\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestSnykCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSnykCommandTestSuite))
}
