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
type GetDeploymentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	envName               string
	fingerprint           string
	artifactPath          string
}

func (suite *GetDeploymentCommandTestSuite) SetupTest() {
	suite.flowName = "get-deployment"
	suite.envName = "get-deployment-env"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
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
	CreateArtifact(suite.flowName, suite.fingerprint, "arti-name", suite.T())
	CreateEnv(global.Org, suite.envName, "server", suite.T())
	ExpectDeployment(suite.flowName, suite.fingerprint, suite.envName, suite.T())
	ExpectDeployment(suite.flowName, suite.fingerprint, suite.envName, suite.T())
}

func (suite *GetDeploymentCommandTestSuite) TestGetDeploymentCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get deployment %s xxx %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "providing no arguments fails",
			cmd:       fmt.Sprintf(`get deployment %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "get deployment fails when --api-token flag is missing",
			cmd:       `get deployment ` + suite.flowName + `#1 --org foo --host bar`,
			golden: "Error: --api-token is not set\n" +
				"Usage: kosli get deployment EXPRESSION [flags]\n",
		},
		{
			wantError: true,
			name:      "get deployment fails when flow does not exist",
			cmd:       `get deployment foo#1` + suite.defaultKosliArguments,
			golden:    "Error: Flow named 'foo' does not exist for organization 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "get deployment fails when deployment does not exist",
			cmd:       fmt.Sprintf(`get deployment %s#20 %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: Deployment number '20' does not exist in flow 'get-deployment' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:       "get deployment works with the # expression",
			cmd:        fmt.Sprintf(`get deployment %s#1 %s`, suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-deployment.txt",
		},
		{
			name:       "get deployment works with the ~ expression",
			cmd:        fmt.Sprintf(`get deployment %s~1 %s`, suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-deployment.txt",
		},
		{
			name:       "get deployment works with just the flow name",
			cmd:        fmt.Sprintf(`get deployment %s %s`, suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-deployment-latest.txt",
		},
		{
			name: "get deployment works with --output json",
			cmd:  fmt.Sprintf(`get deployment %s --output json %s`, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetDeploymentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetDeploymentCommandTestSuite))
}
