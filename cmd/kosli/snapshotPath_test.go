package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotPathTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *SnapshotPathTestSuite) SetupSuite() {
	suite.envName = "snapshot-path-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "server", suite.Suite.T())
}

func (suite *SnapshotPathTestSuite) TestSnapshotPathCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when --path is missing",
			cmd:       fmt.Sprintf(`snapshot path --name foo %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"path\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --name is missing",
			cmd:       fmt.Sprintf(`snapshot path --path testdata/paths-files %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --path does not exist",
			cmd:       fmt.Sprintf(`snapshot path --path testdata/paths-files/does-not-exist --name foo %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: failed to calculate fingerprint for artifact [foo]: failed to open path testdata/paths-files/does-not-exist with error: stat testdata/paths-files/does-not-exist: no such file or directory\n",
		},
		{
			name:   "can report artifact data with --path and --name",
			cmd:    fmt.Sprintf(`snapshot path --path testdata/file1 --name foo %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.envName),
		},
		{
			name:   "can report artifact data with --path and --exclude",
			cmd:    fmt.Sprintf(`snapshot path --path testdata/server --name foo --exclude app.app,"**/logs.txt" %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.envName),
		},
	}

	runTestCmd(suite.Suite.T(), tests)

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotPathTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotPathTestSuite))
}
