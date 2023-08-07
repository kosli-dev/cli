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
type ExpectDeploymentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	envName               string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

func (suite *ExpectDeploymentCommandTestSuite) SetupTest() {
	suite.flowName = "expect-deploy"
	suite.envName = "expect-deploy-env"
	suite.artifactName = "arti"
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
	CreateArtifact(suite.flowName, suite.fingerprint, suite.artifactName, suite.T())
	CreateEnv(global.Org, suite.envName, "server", suite.T())
}

func (suite *ExpectDeploymentCommandTestSuite) TestExpectDeploymentCmd() {
	tests := []cmdTestCase{
		{
			name: "expect deployment works (with --fingerprint)",
			cmd: fmt.Sprintf(`expect deployment --flow %s --fingerprint %s --environment %s --build-url example.com %s`,
				suite.flowName, suite.fingerprint, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("expect deployment of artifact %s was reported to: expect-deploy-env\n", suite.fingerprint),
		},
		{
			name: "expect deployment works (with --artifact-type)",
			cmd: fmt.Sprintf(`expect deployment %s --artifact-type file --flow %s --fingerprint %s --environment %s --build-url example.com %s`,
				suite.artifactPath, suite.flowName, suite.fingerprint, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("expect deployment of artifact %s was reported to: expect-deploy-env\n", suite.fingerprint),
		},
		{
			name: "expect deployment works with --user-data",
			cmd: fmt.Sprintf(`expect deployment --flow %s --fingerprint %s --environment %s --build-url example.com
								--user-data testdata/snyk_scan_example.json %s`,
				suite.flowName, suite.fingerprint, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("expect deployment of artifact %s was reported to: expect-deploy-env\n", suite.fingerprint),
		},
		{
			wantError: true,
			name:      "missing --org flag causes an error",
			cmd: fmt.Sprintf(`expect deployment --flow %s --fingerprint %s --environment %s --build-url example.com
			 		--api-token secret`,
				suite.flowName, suite.fingerprint, suite.envName),
			golden: "Error: --org is not set\nUsage: kosli expect deployment [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd: fmt.Sprintf(`expect deployment --flow %s --fingerprint %s --environment %s --build-url example.com
			 		--org orgX`,
				suite.flowName, suite.fingerprint, suite.envName),
			golden: "Error: --api-token is not set\nUsage: kosli expect deployment [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "expect deployment fails when --user-data is a non-existing file",
			cmd: fmt.Sprintf(`expect deployment --flow %s --fingerprint %s --environment %s --build-url example.com
								--user-data non-existing.json %s`,
				suite.flowName, suite.fingerprint, suite.envName, suite.defaultKosliArguments),
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
		{
			wantError: true,
			name:      "expect deployment fails if --flow is missing",
			cmd: fmt.Sprintf(`expect deployment --fingerprint %s --environment %s --build-url example.com %s`,
				suite.fingerprint, suite.envName, suite.defaultKosliArguments),
			golden: "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError: true,
			name:      "expect deployment fails if --environment is missing",
			cmd: fmt.Sprintf(`expect deployment --flow %s --fingerprint %s --build-url example.com %s`,
				suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
			golden: "Error: required flag(s) \"environment\" not set\n",
		},
		{
			wantError: true,
			name:      "expect deployment fails if --build-url is missing",
			cmd: fmt.Sprintf(`expect deployment --flow %s --fingerprint %s --environment %s %s`,
				suite.flowName, suite.fingerprint, suite.envName, suite.defaultKosliArguments),
			golden: "Error: required flag(s) \"build-url\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExpectDeploymentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ExpectDeploymentCommandTestSuite))
}
