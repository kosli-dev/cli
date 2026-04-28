package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SnapshotCloudRunTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *SnapshotCloudRunTestSuite) SetupTest() {
	suite.envName = "snapshot-cloud-run-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "snapshot cloud-run fails if no args are provided",
			cmd:       fmt.Sprintf(`snapshot cloud-run --project p --region r %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "snapshot cloud-run fails if 2 args are provided",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s xxx --project p --region r %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "snapshot cloud-run fails if --project is missing",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --region r %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"project\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot cloud-run fails if --region is missing",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --project p %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"region\" not set\n",
		},
		{
			name:   "snapshot cloud-run succeeds with required args and prints the placeholder",
			cmd:    fmt.Sprintf(`snapshot cloud-run %s --project p --region r %s`, suite.envName, suite.defaultKosliArguments),
			golden: "cloud-run snapshot: not yet implemented (forced dry-run)\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestSnapshotCloudRunCommandTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotCloudRunTestSuite))
}
