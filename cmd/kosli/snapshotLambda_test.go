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
	zipFunctionName       string
	imageFunctionName     string
}

type snapshotLambdaTestConfig struct {
	requireAuthToBeSet bool
}

func (suite *SnapshotLambdaTestSuite) SetupTest() {
	suite.envName = "snapshot-lambda-env"
	suite.zipFunctionName = "ewelina-test"
	suite.imageFunctionName = "lambda-docker-test"
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
			name: "snapshot lambda works with deprecated --function-name for Zip package type",
			cmd:  fmt.Sprintf(`snapshot lambda %s %s --function-name %s`, suite.envName, suite.defaultKosliArguments, suite.zipFunctionName),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: fmt.Sprintf("Flag --function-name has been deprecated, use --function-names instead\n1 lambda functions were reported to environment %s\n", suite.envName),
		},
		{
			name: "snapshot lambda works with --function-names for Zip package type",
			cmd:  fmt.Sprintf(`snapshot lambda %s %s --function-names %s`, suite.envName, suite.defaultKosliArguments, suite.zipFunctionName),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: fmt.Sprintf("1 lambda functions were reported to environment %s\n", suite.envName),
		},
		{
			name: "snapshot lambda works with --function-names taking a list of functions",
			cmd:  fmt.Sprintf(`snapshot lambda %s %s --function-names %s,%s`, suite.envName, suite.defaultKosliArguments, suite.zipFunctionName, suite.imageFunctionName),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: fmt.Sprintf("2 lambda functions were reported to environment %s\n", suite.envName),
		},
		{
			name: "snapshot lambda works with --function-names for Image package type",
			cmd:  fmt.Sprintf(`snapshot lambda %s %s --function-names %s`, suite.envName, suite.defaultKosliArguments, suite.imageFunctionName),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: fmt.Sprintf("1 lambda functions were reported to environment %s\n", suite.envName),
		},
		{
			name: "snapshot lambda works with --function-names and deprecated --function-version which is ignored",
			cmd:  fmt.Sprintf(`snapshot lambda %s %s --function-names %s --function-version 317`, suite.envName, suite.defaultKosliArguments, suite.zipFunctionName),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: fmt.Sprintf("Flag --function-version has been deprecated, --function-version is no longer supported. It will be removed in a future release.\n1 lambda functions were reported to environment %s\n", suite.envName),
		},
		{
			wantError: false,
			name:      "snapshot lambda without --function-names will report all lambdas in the AWS account",
			cmd:       fmt.Sprintf(`snapshot lambda %s %s`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			goldenRegex: fmt.Sprintf("[0-9]+ lambda functions were reported to environment %s\n", suite.envName),
		},
		{
			wantError: true,
			name:      "snapshot lambda fails when both of --function-name and --function-names are set",
			cmd:       fmt.Sprintf(`snapshot lambda %s --function-name foo --function-names foo %s`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: snapshotLambdaTestConfig{
				requireAuthToBeSet: true,
			},
			golden: "Flag --function-name has been deprecated, use --function-names instead\nError: only one of --function-name, --function-names is allowed\n",
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
			cmd:       fmt.Sprintf(`snapshot lambda %s xxx %s --function-names %s`, suite.envName, suite.defaultKosliArguments, suite.zipFunctionName),
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
