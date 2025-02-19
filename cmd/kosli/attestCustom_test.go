package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AttestCustomCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	typeName            string
	schemaFilePath      string
	jqRules             []string
	attestationDataFile string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestCustomCommandTestSuite) SetupTest() {
	suite.flowName = "attest-custom"
	suite.trailName = "test-321"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	suite.typeName = "person"
	suite.schemaFilePath = ""
	suite.attestationDataFile = "testdata/person-type-data-example.json"
	suite.jqRules = []string{}
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --type %s --attestation-data %s --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.typeName, suite.attestationDataFile, suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)

	CreateCustomAttestationType(suite.typeName, suite.schemaFilePath, suite.jqRules, suite.Suite.T())
	CreateFlow(suite.flowName, suite.Suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.Suite.T())
}

func (suite *AttestCustomCommandTestSuite) TestAttestCustomCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest custom foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flag",
			cmd:       fmt.Sprintf("attest custom foo -t file %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when artifact-name is provided and there is no --artifact-type",
			cmd:       fmt.Sprintf("attest custom wibble %s", suite.defaultKosliArguments),
			golden:    "Error: --artifact-type or --fingerprint must be specified when artifact name ('wibble') argument is supplied.\nUsage: kosli attest custom [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest custom foo --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest custom --name foo --fingerprint xxxx --commit HEAD --origin-url http://example.com %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest custom [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest custom --fingerprint 3214e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint 3214e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-321\" of flow \"attest-custom\" belonging to organization \"docs-cmd-test-user\"\n",
		},
		{
			wantError: true,
			name:      "fails when --name is passed as empty string",
			cmd:       fmt.Sprintf("attest custom --name \"\" --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: flag '--name' is required, but empty string was provided\n",
		},
		{
			name:   "can attest custom against an artifact using artifact-name and --fingerprint",
			cmd:    fmt.Sprintf("attest custom testdata/file1 %s --name foo --fingerprint %s", suite.defaultKosliArguments, suite.artifactFingerprint),
			golden: "custom:person attestation 'foo' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest custom testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'foo' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest custom testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'bar' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom against an artifact using --fingerprint and no artifact-name",
			cmd:    fmt.Sprintf("attest custom --fingerprint %s --name foo --commit HEAD --origin-url http://example.com  %s", suite.artifactFingerprint, suite.defaultKosliArguments),
			golden: "custom:person attestation 'foo' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom against a trail",
			cmd:    fmt.Sprintf("attest custom --name bar --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'bar' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest custom --name additional --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'additional' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest custom --name cli.foo --commit HEAD --origin-url http://example.com  %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'foo' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom attestation with attachment against a trail",
			cmd:    fmt.Sprintf("attest custom --name bar --attachments testdata/file1 %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'bar' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom attestation with external-url against a trail",
			cmd:    fmt.Sprintf("attest custom --name bar --external-url foo=https://foo.com %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'bar' is reported to trail: test-321\n",
		},
		{
			name:   "can attest custom attestation with external-url and external-fingerprint against a trail",
			cmd:    fmt.Sprintf("attest custom --name bar --external-url file=https://foo.com/file --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'bar' is reported to trail: test-321\n",
		},
		{
			wantError: true,
			name:      "fails when external-fingerprint has more items more than external-url",
			cmd:       fmt.Sprintf("attest custom --name bar --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 %s", suite.defaultKosliArguments),
			golden:    "Error: --external-fingerprints have labels that don't have a URL in --external-url\n",
		},
		{
			name:   "can attest custom attestation with description against a trail",
			cmd:    fmt.Sprintf("attest custom --name bar --description 'foo bar foo' %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'bar' is reported to trail: test-321\n",
		},
		{
			name:   "can attest with annotations against a trail",
			cmd:    fmt.Sprintf("attest custom --name bar --annotate foo=bar --annotate baz=\"data with spaces\" %s", suite.defaultKosliArguments),
			golden: "custom:person attestation 'bar' is reported to trail: test-321\n",
		},
		{
			wantError: true,
			name:      "fails when annotation is not valid",
			cmd:       fmt.Sprintf("attest custom --name bar --annotate foo.baz=bar %s", suite.defaultKosliArguments),
			golden:    "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.baz'. Key can only contain [A-Za-z0-9_].\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestCustomCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestCustomCommandTestSuite))
}
