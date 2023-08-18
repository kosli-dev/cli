package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type DiffSnapshotsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName1              string
	envName2              string
	artifactPath          string
}

type diffSnapshotsTestConfig struct {
	reportToEnv1 bool
	reportToEnv2 bool
}

func (suite *DiffSnapshotsCommandTestSuite) SetupTest() {
	suite.envName1 = "env1-to-diff"
	suite.envName2 = "env2-to-diff"

	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName1, "server", suite.T())
	CreateEnv(global.Org, suite.envName2, "server", suite.T())
}

func (suite *DiffSnapshotsCommandTestSuite) TestDiffSnapshotsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "diffing empty envs gives an error",
			cmd:       fmt.Sprintf(`diff snapshots %s %s %s`, suite.envName1, suite.envName1, suite.defaultKosliArguments),
			golden:    "Error: Org: 'docs-cmd-test-user'. Snapshot 'env1-to-diff' resolves to 'env1-to-diff#0'. len(snapshots) == 0. Indexes are 1-based\n",
		},
		{
			wantError: true,
			name:      "diffing empty env against a non-empty env gives an error",
			cmd:       fmt.Sprintf(`diff snapshots %s %s %s`, suite.envName1, suite.envName2, suite.defaultKosliArguments),
			additionalConfig: diffSnapshotsTestConfig{
				reportToEnv1: true,
			},
			golden: "Error: Org: 'docs-cmd-test-user'. Snapshot 'env2-to-diff' resolves to 'env2-to-diff#0'. len(snapshots) == 0. Indexes are 1-based\n",
		},
		{
			name: "diffing the same snapshot works",
			cmd:  fmt.Sprintf(`diff snapshots %s %s %s`, suite.envName1, suite.envName1, suite.defaultKosliArguments),
			additionalConfig: diffSnapshotsTestConfig{
				reportToEnv1: true,
			},
			golden: "",
		},
		{
			name: "diffing the same snapshot with --output json works",
			cmd:  fmt.Sprintf(`diff snapshots %s %s --output json %s`, suite.envName1, suite.envName1, suite.defaultKosliArguments),
			additionalConfig: diffSnapshotsTestConfig{
				reportToEnv1: true,
			},
			golden: "",
		},
		{
			name: "diffing two envs with the same snapshot works",
			cmd:  fmt.Sprintf(`diff snapshots %s %s %s`, suite.envName1, suite.envName2, suite.defaultKosliArguments),
			additionalConfig: diffSnapshotsTestConfig{
				reportToEnv1: true,
				reportToEnv2: true,
			},
			golden: "",
		},
		{
			name: "diffing two envs with the same snapshot with --show-unchanged enabled works",
			cmd:  fmt.Sprintf(`diff snapshots %s %s --show-unchanged %s`, suite.envName1, suite.envName2, suite.defaultKosliArguments),
			additionalConfig: diffSnapshotsTestConfig{
				reportToEnv1: true,
				reportToEnv2: true,
			},
			golden: "",
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil {
			if t.additionalConfig.(diffSnapshotsTestConfig).reportToEnv1 {
				ReportServerArtifactToEnv([]string{suite.artifactPath}, suite.envName1, suite.T())
			}
			if t.additionalConfig.(diffSnapshotsTestConfig).reportToEnv2 {
				ReportServerArtifactToEnv([]string{suite.artifactPath}, suite.envName2, suite.T())
			}
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDiffSnapshotsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(DiffSnapshotsCommandTestSuite))
}
