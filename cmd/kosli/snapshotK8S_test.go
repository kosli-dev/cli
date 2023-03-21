package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotK8STestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *SnapshotK8STestSuite) SetupTest() {
	suite.envName = "snapshot-k8s-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "K8S", suite.T())
}

func (suite *SnapshotK8STestSuite) TestSnapshotK8SCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "snapshot K8S fails if both --namespaces and --exclude-namespaces are set",
			cmd:       fmt.Sprintf(`snapshot k8s %s --namespaces default --exclude-namespaces default %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --namespaces, --exclude-namespaces is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if 2 args are provided",
			cmd:       fmt.Sprintf(`snapshot k8s %s xxx %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if no args are set",
			cmd:       fmt.Sprintf(`snapshot k8s %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotK8STestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotK8STestSuite))
}
