package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetSnapshotCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
	emptyEnvName          string
}

func (suite *GetSnapshotCommandTestSuite) SetupTest() {
	suite.envName = "get-snapshot-env"
	suite.emptyEnvName = "get-snapshot-empty-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "server", suite.T())
	ReportServerArtifactToEnv([]string{"testdata/folder1/hello.txt"}, suite.envName, suite.T())
	ReportServerArtifactToEnv([]string{"testdata/file1"}, suite.envName, suite.T())
	ReportServerArtifactToEnv([]string{"testdata/report.xml"}, suite.envName, suite.T())
	CreateEnv(global.Org, suite.emptyEnvName, "server", suite.T())
}

// TODO: Add test for a snappish of the environemnt name
func (suite *GetSnapshotCommandTestSuite) TestGetSnapshotCmd() {
	tests := []cmdTestCase{
		{
			name: "can get the first snapshot with # snappish",
			cmd:  fmt.Sprintf(`get snapshot %s#1 %s`, suite.envName, suite.defaultKosliArguments),
		},
		{
			name: "can get the first snapshot with # snappish with --output json",
			cmd:  fmt.Sprintf(`get snapshot %s#1 --output json %s`, suite.envName, suite.defaultKosliArguments),
		},
		{
			name: "can get the second snapshot with ~ snappish",
			cmd:  fmt.Sprintf(`get snapshot %s~1 %s`, suite.envName, suite.defaultKosliArguments),
		},
		{
			name: "can get the second snapshot with it's datetime snappish",
			cmd:  fmt.Sprintf(`get snapshot %s@{%s} %s`, suite.envName, currentTimeForSnappish(), suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "getting snapshot for an empty environment exist fails",
			cmd:       fmt.Sprintf(`get snapshot %s %s`, suite.emptyEnvName, suite.defaultKosliArguments),
			golden:    "Error: Org: 'docs-cmd-test-user'. Snapshot 'get-snapshot-empty-env#-1' resolves to 'get-snapshot-empty-env#0'. len(snapshots) == 0. Indexes are 1-based\n",
		},
		{
			wantError: true,
			name:      "getting non-existing snapshot fails",
			cmd:       fmt.Sprintf(`get snapshot %s#23 %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: Org: 'docs-cmd-test-user'. Snapshot 'get-snapshot-env#23' resolves to 'get-snapshot-env#23'. len(snapshots) == 3. Indexes are 1-based\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func currentTimeForSnappish() string {
	// Add a little time as we will lose ms resolution
	timeIn1Second := time.Now().Local().Add(time.Second * 1).UTC()

	return timeIn1Second.Format("2006-01-02T15:04:05")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetSnapshotCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetSnapshotCommandTestSuite))
}
