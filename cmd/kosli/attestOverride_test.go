package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AttestOverrideCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestOverrideCommandTestSuite) SetupTest() {
	suite.flowName = "attest-override"
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

func (suite *AttestOverrideCommandTestSuite) TestAttestOverrideCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest override foo bar --name foo --reason r %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when missing required --name flag",
			cmd:       fmt.Sprintf("attest override --reason r --original-attestation-type generic --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when missing required --reason flag",
			cmd:       fmt.Sprintf("attest override --name foo --original-attestation-type generic --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"reason\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when missing required --original-attestation-type flag",
			cmd:       fmt.Sprintf("attest override --name foo --reason r --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"original-attestation-type\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when missing required --new-compliance-status flag",
			cmd:       fmt.Sprintf("attest override --name foo --reason r --original-attestation-type generic %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"new-compliance-status\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --name is passed as empty string",
			cmd:       fmt.Sprintf("attest override --name \"\" --reason r --original-attestation-type generic --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden:    "Error: flag '--name' is required, but empty string was provided\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type are set",
			cmd:       fmt.Sprintf("attest override testdata/file1 --fingerprint xxxx --artifact-type file --name bar --reason r --original-attestation-type generic --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest override --name foo --fingerprint xxxx --reason r --original-attestation-type generic --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest override [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "fails when --repo-url is not a valid URL",
			cmd:       fmt.Sprintf("attest override --name foo --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --reason r --original-attestation-type generic --new-compliance-status=true --repo-url not-a-url %s", suite.defaultKosliArguments),
			golden:    "Error: --repo-url 'not-a-url' is not a valid URL\n",
		},
		{
			wantError: true,
			name:      "fails when --repo-provider is not an allowed value",
			cmd:       fmt.Sprintf("attest override --name foo --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --reason r --original-attestation-type generic --new-compliance-status=true --repo-provider jenkins %s", suite.defaultKosliArguments),
			golden:    "Error: --repo-provider 'jenkins' is not allowed. Must be one of: github, gitlab, bitbucket, bitbucket_cloud, bitbucket_dc, azure-devops, azure_devops_services, azure_devops_server, git, subversion\n",
		},
		{
			name:   "can override an attestation against a trail (compliant)",
			cmd:    fmt.Sprintf("attest override --name bar --reason 'manual review' --original-attestation-type generic --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden: "attestation 'bar' has been overridden in trail: test-123\n",
		},
		{
			name:   "can override an attestation against a trail (non-compliant)",
			cmd:    fmt.Sprintf("attest override --name bar --reason 'failed audit' --new-compliance-status=false --original-attestation-type generic %s", suite.defaultKosliArguments),
			golden: "attestation 'bar' has been overridden in trail: test-123\n",
		},
		{
			name:   "can override an attestation against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest override --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --reason 'approved out of band' --original-attestation-type generic --new-compliance-status=true %s", suite.defaultKosliArguments),
			golden: "attestation 'foo' has been overridden in trail: test-123\n",
		},
		{
			name:   "can override an attestation with description, annotation and external-url",
			cmd:    fmt.Sprintf("attest override --name bar --reason r --original-attestation-type generic --new-compliance-status=true --description 'some desc' --annotate foo=bar --external-url ref=https://example.com %s", suite.defaultKosliArguments),
			golden: "attestation 'bar' has been overridden in trail: test-123\n",
		},
		{
			name:   "can override an attestation with --user-data sent as JSON in the payload",
			cmd:    fmt.Sprintf("attest override --name bar --reason r --original-attestation-type generic --new-compliance-status=true --user-data testdata/person-type-data-example.json %s", suite.defaultKosliArguments),
			golden: "attestation 'bar' has been overridden in trail: test-123\n",
		},
		{
			name:        "dry-run prints the override URL and the JSON payload without contacting the server",
			cmd:         fmt.Sprintf("attest override --name bar --reason r --original-attestation-type generic --new-compliance-status=true --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `(?s)THIS IS A DRY-RUN.*attestations/docs-cmd-test-user/attest-override/trail/test-123/override.*"attestation_name": "bar".*"reason": "r".*"new_compliance_status": true.*"original_attestation_type": "generic"`,
		},
		{
			wantError: true,
			name:      "fails when annotation key is invalid",
			cmd:       fmt.Sprintf("attest override --name bar --reason r --original-attestation-type generic --new-compliance-status=true --annotate foo.baz=bar %s", suite.defaultKosliArguments),
			golden:    "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.baz'. Key can only contain [A-Za-z0-9_]\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestAttestOverrideCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestOverrideCommandTestSuite))
}
