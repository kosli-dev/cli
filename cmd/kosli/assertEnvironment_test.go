package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AssertEnvCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

func (suite *AssertEnvCommandTestSuite) SetupTest() {
	suite.envName = "env-to-assert"
	suite.flowName = "some-flow"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	kosliClient = requests.NewKosliClient(1, false, logger)
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)

	CreateEnv(global.Owner, suite.envName, "server", suite.T())
	CreateFlow(suite.flowName, suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifact(suite.flowName, suite.fingerprint, suite.artifactName, suite.T())
}

func (suite *AssertEnvCommandTestSuite) TestAssertEnvCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "asserting an empty env results in non-zero exit",
			cmd:       fmt.Sprintf(`assert env %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: Org: 'docs-cmd-test-user'. Snapshot 'env-to-assert#-1' resolves to 'env-to-assert#0'. len(snapshots) == 0. Indexes are 1-based\n",
		},
		{
			wantError: true,
			name:      "asserting a non existing env fails",
			cmd:       `assert env non-existing` + suite.defaultKosliArguments,
			golden:    "Error: Environment named 'non-existing' does not exist for Organization 'docs-cmd-test-user'\n",
		},
		{
			name:   "asserting a non-empty env results in OK and zero exit",
			cmd:    fmt.Sprintf(`assert env %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: "COMPLIANT\n",
		},
	}

	for _, t := range tests {
		if !t.wantError {
			ExpectDeployment(suite.flowName, suite.fingerprint, suite.envName, suite.T())
			ReportServerArtifactToEnv([]string{suite.artifactPath}, suite.envName, suite.T())
			runTestCmd(suite.T(), []cmdTestCase{t})
		}
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertEnvCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertEnvCommandTestSuite))
}
