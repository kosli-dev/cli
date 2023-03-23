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
type SnapshotS3TestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
	bucketName            string
}

type snapshotS3TestConfig struct {
	requireAuthToBeSet bool
}

func (suite *SnapshotS3TestSuite) SetupTest() {
	suite.envName = "snapshot-s3-env"
	suite.bucketName = "kosli-cli-public"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "S3", suite.T())
}

func (suite *SnapshotS3TestSuite) TestSnapshotS3Cmd() {
	tests := []cmdTestCase{
		{
			name: "snapshot s3 works with --bucket",
			cmd:  fmt.Sprintf(`snapshot s3 %s %s --bucket %s`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			additionalConfig: snapshotS3TestConfig{
				requireAuthToBeSet: true,
			},
			golden: "bucket " + suite.bucketName + " was reported to environment " + suite.envName + "\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails without --bucket",
			cmd:       fmt.Sprintf(`snapshot s3 %s %s`, suite.envName, suite.defaultKosliArguments),
			additionalConfig: snapshotS3TestConfig{
				requireAuthToBeSet: true,
			},
			golden: "Error: required flag \"bucket\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails if no args are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "snapshot s3 fails two args are set",
			cmd:       fmt.Sprintf(`snapshot s3 %s xxx %s --bucket %s`, suite.envName, suite.defaultKosliArguments, suite.bucketName),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(snapshotS3TestConfig).requireAuthToBeSet {
			testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"})
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotS3TestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotS3TestSuite))
}
