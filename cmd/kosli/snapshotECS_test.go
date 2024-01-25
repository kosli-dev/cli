package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotECSTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

type snapshotECSTestConfig struct {
	requireAuthToBeSet bool
}

func (suite *SnapshotECSTestSuite) SetupTest() {
	suite.envName = "snapshot-ecs-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "ECS", suite.T())
}

func (suite *SnapshotECSTestSuite) TestSnapshotECSCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "snapshot ECS fails if --cluster is missing",
			cmd:       fmt.Sprintf(`snapshot ecs %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"cluster\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot ECS fails if 2 args are provided",
			cmd:       fmt.Sprintf(`snapshot ecs %s xxx --cluster sss --service-name sss %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "snapshot ECS fails if no args are set",
			cmd:       fmt.Sprintf(`snapshot ecs %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			name: "snapshot ECS works with --cluster",
			cmd:  fmt.Sprintf(`snapshot ecs %s %s --cluster merkely`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: snapshotECSTestConfig{
				requireAuthToBeSet: true,
			},
			golden: "[2] containers were reported to environment snapshot-ecs-env\n",
		},
		{
			name: "snapshot ECS works with --service-name",
			cmd:  fmt.Sprintf(`snapshot ecs %s %s --cluster merkely --service-name merkely`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: snapshotECSTestConfig{
				requireAuthToBeSet: true,
			},
			golden: "[2] containers were reported to environment snapshot-ecs-env\n",
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(snapshotECSTestConfig).requireAuthToBeSet {
			testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotECSTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotECSTestSuite))
}
