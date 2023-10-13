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
type ListDeploymentsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName1             string
	flowName2             string
	artifactName          string
	artifactPath          string
	fingerprint           string
	envName               string
}

func (suite *ListDeploymentsCommandTestSuite) SetupTest() {
	suite.flowName1 = "list-deployments-empty"
	suite.flowName2 = "list-deployments"
	suite.envName = "list-deployments-env"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlow(suite.flowName1, suite.T())
	CreateFlow(suite.flowName2, suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifact(suite.flowName2, suite.fingerprint, suite.artifactName, suite.T())
	CreateEnv(global.Org, suite.envName, "server", suite.T())
	ExpectDeployment(suite.flowName2, suite.fingerprint, suite.envName, suite.T())
}

func (suite *ListDeploymentsCommandTestSuite) TestListDeploymentsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing flow flag causes an error",
			cmd:       fmt.Sprintf(`list deployments %s`, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError: true,
			name:      "non-existing flow causes an error",
			cmd:       fmt.Sprintf(`list deployments --flow non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Flow named 'non-existing' does not exist for organization 'docs-cmd-test-user'\n",
		},
		// TODO: the correct error is overwritten by the hack flag value check in root.go
		{
			wantError: true,
			name:      "negative page number causes an error",
			cmd:       fmt.Sprintf(`list deployments --flow foo --page -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "negative page limit causes an error",
			cmd:       fmt.Sprintf(`list deployments --flow foo --page-limit -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page-limit' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`list deployments --flow %s --org orgX`, suite.flowName1),
			golden:    "Error: --api-token is not set\nUsage: kosli list deployments [flags]\n",
		},
		{
			name:   "listing deployments on an empty flow works",
			cmd:    fmt.Sprintf(`list deployments --flow %s %s`, suite.flowName1, suite.defaultKosliArguments),
			golden: "No deployments were found.\n",
		},
		{
			name:   "listing second page of deployments on an empty flow works",
			cmd:    fmt.Sprintf(`list deployments --flow %s --page 2 %s`, suite.flowName1, suite.defaultKosliArguments),
			golden: "No deployments were found at page number 2.\n",
		},
		{
			name:   "listing deployments on an empty flow with --output json works",
			cmd:    fmt.Sprintf(`list deployments --flow %s --output json %s`, suite.flowName1, suite.defaultKosliArguments),
			golden: "[]\n",
		},
		{
			name:       "listing deployments on a flow works",
			cmd:        fmt.Sprintf(`list deployments --flow %s %s`, suite.flowName2, suite.defaultKosliArguments),
			goldenFile: "output/list/list-deployments.txt",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListDeploymentsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListDeploymentsCommandTestSuite))
}
