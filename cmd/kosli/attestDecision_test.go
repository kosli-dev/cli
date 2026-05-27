package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AttestDecisionCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestDecisionCommandTestSuite) SetupTest() {
	suite.flowName = "attest-decision"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateControl(global.Org, "RCTL-043", "Test Control", suite.T())
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.T())
}

func (suite *AttestDecisionCommandTestSuite) TestAttestDecisionCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest decision foo bar %s --control RCTL-043 --compliant=true", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when missing --name flag",
			cmd:       fmt.Sprintf("attest decision %s --control RCTL-043 --compliant=true", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when missing --control flag",
			cmd:       fmt.Sprintf("attest decision --name foo %s --compliant=true", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"control\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --compliant is not set",
			cmd:       fmt.Sprintf("attest decision --name foo --control RCTL-043 %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"compliant\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type are set",
			cmd:       fmt.Sprintf("attest decision testdata/file1 --fingerprint xxxx --artifact-type file --name foo --control RCTL-043 --compliant=true %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not a valid SHA256",
			cmd:       fmt.Sprintf("attest decision --name foo --fingerprint xxxx --control RCTL-043 --compliant=true %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest decision [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "fails when --name is passed as empty string",
			cmd:       fmt.Sprintf("attest decision --name \"\" --control RCTL-043 --compliant=true %s", suite.defaultKosliArguments),
			golden:    "Error: flag '--name' is required, but empty string was provided\n",
		},
		{
			name:   "can record a compliant decision against a trail",
			cmd:    fmt.Sprintf("attest decision --name foo --control RCTL-043 --compliant=true %s", suite.defaultKosliArguments),
			golden: "decision attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can record a non-compliant decision against a trail",
			cmd:    fmt.Sprintf("attest decision --name foo --control RCTL-043 --compliant=false %s", suite.defaultKosliArguments),
			golden: "decision attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can record a decision linked to a specific artifact by fingerprint",
			cmd:    fmt.Sprintf("attest decision --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --control RCTL-043 --compliant=true %s", suite.defaultKosliArguments),
			golden: "decision attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can record a decision with an attachment",
			cmd:    fmt.Sprintf("attest decision --name foo --control RCTL-043 --compliant=true --attachments testdata/file1 %s", suite.defaultKosliArguments),
			golden: "decision attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can record a decision with description",
			cmd:    fmt.Sprintf("attest decision --name foo --control RCTL-043 --compliant=true --description 'evaluation passed' %s", suite.defaultKosliArguments),
			golden: "decision attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can record a decision with annotations",
			cmd:    fmt.Sprintf("attest decision --name foo --control RCTL-043 --compliant=true --annotate key=value %s", suite.defaultKosliArguments),
			golden: "decision attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can record a decision with user data",
			cmd:    fmt.Sprintf("attest decision --name foo --control RCTL-043 --compliant=true --user-data testdata/person-type-data-example.json %s", suite.defaultKosliArguments),
			golden: "decision attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "fails when annotation key is invalid",
			cmd:       fmt.Sprintf("attest decision --name foo --control RCTL-043 --compliant=true --annotate foo.bar=baz %s", suite.defaultKosliArguments),
			golden:    "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.bar'. Key can only contain [A-Za-z0-9_]\n",
		},
		{
			wantError: true,
			name:      "fails when --name has invalid dot format",
			cmd:       fmt.Sprintf("attest decision --name .foo --control RCTL-043 --compliant=true %s", suite.defaultKosliArguments),
			golden:    "Error: failed to parse attestation name: invalid attestation name format: .foo\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestAttestDecisionCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestDecisionCommandTestSuite))
}
