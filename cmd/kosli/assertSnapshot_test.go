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
	envName               string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

type assertSnapshotTestConfig struct {
	reportToEnv      bool
	expectDeployment bool
}

func (suite *AssertSnapshotCommandTestSuite) SetupTest() {
	suite.envName = "env-to-assert"
	suite.flowName = "assert-snapshot"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/assert_snapshot_artifact.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "server", suite.T())
	CreateFlow(suite.flowName, suite.T())
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
			name:      "missing --org fails",
			cmd:       fmt.Sprintf(`assert snapshot %s --api-token secret`, suite.envName),
			golden:    "Error: --org is not set\nUsage: kosli assert snapshot [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "asserting an empty env results in non-zero exit",
			cmd:       fmt.Sprintf(`assert snapshot %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: Org: 'docs-cmd-test-user'. Snapshot 'env-to-assert#-1' resolves to 'env-to-assert#0'. len(snapshots) == 0. Indexes are 1-based\n",
		},
		{
			wantError: true,
			name:      "asserting a non existing env fails",
			cmd:       `assert snapshot non-existing` + suite.defaultKosliArguments,
			golden:    "Error: Environment named 'non-existing' does not exist for organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "asserting a non-compliant env results in INCOMPLIANT and non-zero exit",
			cmd:       fmt.Sprintf(`assert snapshot %s %s`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: assertSnapshotTestConfig{
				reportToEnv:      true,
				expectDeployment: false,
			},
			golden: "Error: INCOMPLIANT\n",
		},
		{
			name: "asserting a compliant env results in COMPLIANT and zero exit",
			cmd:  fmt.Sprintf(`assert snapshot %s %s`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: assertSnapshotTestConfig{
				reportToEnv:      true,
				expectDeployment: true,
			},
			golden: "COMPLIANT\n",
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(assertSnapshotTestConfig).reportToEnv {
			if t.additionalConfig.(assertSnapshotTestConfig).expectDeployment {
				ExpectDeployment(suite.flowName, suite.fingerprint, suite.envName, suite.T())
			}
			ReportServerArtifactToEnv([]string{suite.artifactPath}, suite.envName, suite.T())
			runTestCmd(suite.T(), []cmdTestCase{t})
		}
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertSnapshotCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertSnapshotCommandTestSuite))
}
