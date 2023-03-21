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
type SnapshotLambdaTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
	functionName          string
}

type snapshotLambdaTestConfig struct {
	requireAuthToBeSet bool
}

func (suite *SnapshotLambdaTestSuite) SetupTest() {
	suite.envName = "snapshot-lambda-env"
	suite.functionName = "reporter-kosli-prod"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "lambda", suite.T())
}

func (suite *SnapshotLambdaTestSuite) TestSnapshotLambdaCmd() {
	tests := []cmdTestCase{
		{
			name: "snapshot lambda works with --function-name",
			cmd:  fmt.Sprintf(`snapshot lambda %s %s --function-name %s`, suite.envName, suite.defaultKosliArguments, suite.functionName),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: suite.functionName + " lambda function was reported to environment " + suite.envName + "\n",
		},
		{
			name: "snapshot lambda works with --function-name and --function-version",
			cmd:  fmt.Sprintf(`snapshot lambda %s %s --function-name %s --function-version 317`, suite.envName, suite.defaultKosliArguments, suite.functionName),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: suite.functionName + " lambda function was reported to environment " + suite.envName + "\n",
		},
		{
			wantError: true,
			name:      "snapshot lambda fails without --function-name",
			cmd:       fmt.Sprintf(`snapshot lambda %s %s --function-version 317`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: "Error: required flag(s) \"function-name\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot lambda fails if no args are set",
			cmd:       fmt.Sprintf(`snapshot lambda %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "snapshot lambda fails two args are set",
			cmd:       fmt.Sprintf(`snapshot lambda %s xxx %s --function-name %s`, suite.envName, suite.defaultKosliArguments, suite.functionName),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(snapshotLambdaTestConfig).requireAuthToBeSet {
			testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotLambdaTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotLambdaTestSuite))
}
