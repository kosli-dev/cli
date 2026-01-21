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
type AssertSnapshotCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	nonCompliantEnvName   string
	compliantEnvName      string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

type assertSnapshotTestConfig struct {
	reportToEnv bool
	envName     string
}

func (suite *AssertSnapshotCommandTestSuite) SetupTest() {
	suite.nonCompliantEnvName = "env-to-assert"
	suite.compliantEnvName = "compliant-env-to-assert"
	suite.flowName = "assert-snapshot"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/assert_snapshot_artifact.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	// Non-compliant environment
	CreateEnv(global.Org, suite.nonCompliantEnvName, "server", suite.T())
	CreateFlow(suite.flowName, suite.T())

	// In order for an environment to be compliant instead of unknown (which we convert to false), it must have
	// a polict attached.
	CreateEnv(global.Org, suite.compliantEnvName, "server", suite.T())
	CreatePolicy(global.Org, "server-policy", suite.T())
	AttachPolicy([]string{suite.compliantEnvName}, "server-policy", suite.T())

	//Create artifact to report to environments
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifact(suite.flowName, suite.fingerprint, suite.artifactName, suite.T())

}

func (suite *AssertSnapshotCommandTestSuite) TestAssertSnapshotCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "01 missing --org fails",
			cmd:       fmt.Sprintf(`assert snapshot %s --api-token secret`, suite.nonCompliantEnvName),
			golden:    "Error: --org is not set\nUsage: kosli assert snapshot ENVIRONMENT-NAME-OR-EXPRESSION [flags]\n",
		},
		{
			wantError: true,
			name:      "02 asserting an empty env results in non-zero exit",
			cmd:       fmt.Sprintf(`assert snapshot %s %s`, suite.nonCompliantEnvName, suite.defaultKosliArguments),
			golden:    "Error: Org: 'docs-cmd-test-user'. Snapshot 'env-to-assert#-1' resolves to 'env-to-assert#0'. len(snapshots) == 0. Indexes are 1-based\n",
		},
		{
			wantError: true,
			name:      "03 asserting a non existing env fails",
			cmd:       `assert snapshot non-existing` + suite.defaultKosliArguments,
			golden:    "Error: Environment named 'non-existing' does not exist for organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "04 asserting a non compliant env results in INCOMPLIANT and non-zero exit",
			cmd:       fmt.Sprintf(`assert snapshot %s %s`, suite.nonCompliantEnvName, suite.defaultKosliArguments),
			additionalConfig: assertSnapshotTestConfig{
				reportToEnv: true,
				envName:     suite.nonCompliantEnvName,
			},
			golden: "Error: INCOMPLIANT\n",
		},
		{
			wantError: false,
			name:      "05 asserting a compliant env results in COMPLIANT and zero exit",
			cmd:       fmt.Sprintf(`assert snapshot %s %s`, suite.compliantEnvName, suite.defaultKosliArguments),
			additionalConfig: assertSnapshotTestConfig{
				reportToEnv: true,
				envName:     suite.compliantEnvName,
			},
			golden: "COMPLIANT\n",
		},
		{
			wantError: true,
			name:      "06 asserting an env using expression with both # and ~ fails",
			cmd:       fmt.Sprintf(`assert snapshot %s#~ %s`, suite.compliantEnvName, suite.defaultKosliArguments),
			golden:    "Error: invalid expression: compliant-env-to-assert#~. Both '~' and '#' are present\n",
		},
		{
			wantError: true,
			name:      "07 asserting an env using expression with a non-integer fails",
			cmd:       fmt.Sprintf(`assert snapshot %s#five %s`, suite.compliantEnvName, suite.defaultKosliArguments),
			golden:    "Error: invalid expression: compliant-env-to-assert#five. 'five' is not an integer\n",
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(assertSnapshotTestConfig).reportToEnv {
			ReportServerArtifactToEnv([]string{suite.artifactPath}, t.additionalConfig.(assertSnapshotTestConfig).envName, suite.T())
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertSnapshotCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertSnapshotCommandTestSuite))
}
