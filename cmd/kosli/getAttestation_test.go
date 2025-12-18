package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetAttestationCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
	trailName             string
	attestationId         string
}

func (suite *GetAttestationCommandTestSuite) SetupTest() {
	suite.flowName = "get-attestation"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	suite.trailName = "cli-build-1"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.Suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.Suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.Suite.T(), err)
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.fingerprint, suite.artifactName, suite.Suite.T())
	CreateGenericArtifactAttestation(suite.flowName, suite.trailName, suite.fingerprint, "first-artifact-attestation", true, suite.Suite.T())
	CreateGenericTrailAttestation(suite.flowName, suite.trailName, "first-trail-attestation", suite.Suite.T())
	CreateGenericArtifactAttestation(suite.flowName, suite.trailName, suite.fingerprint, "second-artifact-attestation", true, suite.Suite.T())
	CreateGenericTrailAttestation(suite.flowName, suite.trailName, "second-trail-attestation", suite.Suite.T())

	suite.attestationId = GetAttestationId(suite.flowName, suite.trailName, "first-trail-attestation", suite.Suite.T())
}

func (suite *GetAttestationCommandTestSuite) TestGetAttestationCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "if no attestation found, say so",
			cmd:       fmt.Sprintf(`get attestation non-existent-attestation --flow %s --trail %s %s`, suite.flowName, suite.trailName, suite.defaultKosliArguments),
			golden:    "No attestations found.\n",
		},
		{
			wantError:  false,
			name:       "if no attestation found return empty list in json format",
			cmd:        fmt.Sprintf(`get attestation non-existent-attestation --flow %s --trail %s %s --output json`, suite.flowName, suite.trailName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"", "[]"}},
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get attestation first-attestation second-attestation --flow %s --trail %s %s`, suite.flowName, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "missing --flow fails when ATTESTATION-NAME is provided",
			cmd:       fmt.Sprintf(`get attestation first-artifact-attestation --trail %s %s`, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`get attestation first-artifact-attestation --flow %s --org orgX`, suite.flowName),
			golden:    "Error: --api-token is not set\nUsage: kosli get attestation [ATTESTATION-NAME] [flags]\n",
		},
		{
			name: "getting an existing trail attestation works",
			cmd:  fmt.Sprintf(`get attestation first-trail-attestation --flow %s --trail %s %s`, suite.flowName, suite.trailName, suite.defaultKosliArguments),
		},
		{
			name:       "getting an existing trail attestation with --output json works",
			cmd:        fmt.Sprintf(`get attestation first-trail-attestation --flow %s --trail %s --output json %s`, suite.flowName, suite.trailName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"", "non-empty"}},
		},
		{
			name: "getting an existing artifact attestation works",
			cmd:  fmt.Sprintf(`get attestation first-artifact-attestation --flow %s --fingerprint %s %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
		},
		{
			name:       "getting an existing artifact attestation with --output json works",
			cmd:        fmt.Sprintf(`get attestation first-artifact-attestation --flow %s --fingerprint %s --output json %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"", "non-empty"}},
		},
		{
			wantError: true,
			name:      "missing both trail and fingerprint fails",
			cmd:       fmt.Sprintf(`get attestation first-artifact-attestation --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: at least one of --trail, --fingerprint is required\n",
		},
		{
			wantError: true,
			name:      "providing both trail and fingerprint fails",
			cmd:       fmt.Sprintf(`get attestation first-artifact-attestation --flow %s --trail %s --fingerprint %s %s`, suite.flowName, suite.trailName, suite.fingerprint, suite.defaultKosliArguments),
			golden:    "Error: only one of --trail, --fingerprint is allowed\n",
		},
		{
			name: "can get an attestation from its id",
			cmd:  fmt.Sprintf(`get attestation --attestation-id %s %s`, suite.attestationId, suite.defaultKosliArguments),
		},
		{
			wantError:  false,
			name:       "if no attestation found when getting by id return empty list in json format",
			cmd:        fmt.Sprintf(`get attestation --attestation-id %s --output json %s`, "non-existent-attestation-id", suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "length:0"}},
		},
		{
			wantError: true,
			name:      "providing both attestation id and attestation name fails",
			cmd:       fmt.Sprintf(`get attestation %s --attestation-id %s %s`, "first-artifact-attestation", suite.attestationId, suite.defaultKosliArguments),
			golden:    "Error: --attestation-id cannot be used when ATTESTATION-NAME is provided\n",
		},
		{
			wantError: true,
			name:      "providing both attestation id and trail fails",
			cmd:       fmt.Sprintf(`get attestation --attestation-id %s --trail %s %s`, suite.attestationId, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: --flow, --trail, and --fingerprint flags cannot be used with --attestation-id\n",
		},
		{
			wantError: true,
			name:      "providing both attestation id and fingerprint fails",
			cmd:       fmt.Sprintf(`get attestation --attestation-id %s --fingerprint %s %s`, suite.attestationId, suite.fingerprint, suite.defaultKosliArguments),
			golden:    "Error: --flow, --trail, and --fingerprint flags cannot be used with --attestation-id\n",
		},
		{
			wantError: true,
			name:      "providing both attestation id and flow fails",
			cmd:       fmt.Sprintf(`get attestation --attestation-id %s --flow %s %s`, suite.attestationId, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: --flow, --trail, and --fingerprint flags cannot be used with --attestation-id\n",
		},
		{
			wantError: true,
			name:      "providing neither attestation id or flow fails",
			cmd:       fmt.Sprintf(`get attestation  %s`, suite.defaultKosliArguments),
			golden:    "Error: one of ATTESTATION-NAME argument or --attestation-id flag is required\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetAttestationCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetAttestationCommandTestSuite))
}
