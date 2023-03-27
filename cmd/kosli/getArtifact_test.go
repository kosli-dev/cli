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
type GetArtifactCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

func (suite *GetArtifactCommandTestSuite) SetupTest() {
	suite.flowName = "get-artifact"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifact(suite.flowName, suite.fingerprint, suite.artifactName, suite.T())
}

func (suite *GetArtifactCommandTestSuite) TestGetArtifactCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "getting a non existing artifact fails",
			cmd:       fmt.Sprintf(`get artifact %s@8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c' does not exist in flow 'get-artifact' belonging to organization 'docs-cmd-test-user-shared'\n",
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get artifact %s@8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c xxx %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`get artifact %s@%s --org orgX`, suite.flowName, suite.fingerprint),
			golden:    "Error: --api-token is not set\nUsage: kosli get artifact EXPRESSION [flags]\n",
		},
		{
			wantError: true,
			name:      "getting an existing artifact using an invalid expression fails",
			cmd:       fmt.Sprintf(`get artifact %s#%s %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: invalid expression: %s#%s\n", suite.flowName, suite.fingerprint),
		},
		{
			name:       "getting an existing artifact using fingerprint works",
			cmd:        fmt.Sprintf(`get artifact %s@%s %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
			goldenFile: "output/get/get-artifact.txt",
		},
		{
			name: "getting an existing artifact using fingerprint with --output json works",
			cmd:  fmt.Sprintf(`get artifact %s@%s --output json %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
		},
		{
			name:       "get an existing artifact using commit works",
			cmd:        fmt.Sprintf(`get artifact %s:0fc1ba9876f91b215679f3649b8668085d820ab5 %s`, suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-artifact.txt",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetArtifactCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetArtifactCommandTestSuite))
}
