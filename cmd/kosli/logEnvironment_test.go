package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type LogEnvironmentCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	eventsEnvName         string
	firstArtifactPath     string
	secondArtifactPath    string
}

func (suite *LogEnvironmentCommandTestSuite) SetupTest() {
	suite.eventsEnvName = "list-events-env"
	suite.firstArtifactPath = "testdata/report.xml"
	suite.secondArtifactPath = "testdata/file1"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateEnv(global.Org, suite.eventsEnvName, "server", suite.T())
}

func (suite *LogEnvironmentCommandTestSuite) TestLogEnvironmentCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "listing events fails when env does not exist",
			cmd:       fmt.Sprintf(`log env non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Environment named 'non-existing' does not exist for organization 'docs-cmd-test-user'. \n",
		},
		// TODO: the correct error is overwritten by the hack flag value check in root.go
		{
			wantError: true,
			name:      "listing events fails when --page is negative",
			cmd:       fmt.Sprintf(`log env %s --page -1 %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "listing events fails when --page-limit is negative",
			cmd:       fmt.Sprintf(`log env %s --page-limit -1 %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			golden:    "Error: flag '--page-limit' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "listing events fails when 2 args are provided",
			cmd:       fmt.Sprintf(`log env %s arg2 %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "listing events fails when no args are provided",
			cmd:       fmt.Sprintf(`log env %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			name:   "listing events works when env is empty",
			cmd:    fmt.Sprintf(`log env %s %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			golden: "No environment events were found.\n",
		},
		{
			name: "listing events works when env contains snapshots",
			cmd:  fmt.Sprintf(`log env %s %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			additionalConfig: listSnapshotsTestConfig{
				reportToEnv: true,
			},
		},
		{
			name: "listing events works with --output json when env contains snapshots",
			cmd:  fmt.Sprintf(`log env %s --output json %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			additionalConfig: listSnapshotsTestConfig{
				reportToEnv: true,
			},
		},
		{
			name: "listing events works when env contains snapshots and NOW is provided as interval",
			cmd:  fmt.Sprintf(`log env %s --interval NOW %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			additionalConfig: listSnapshotsTestConfig{
				reportToEnv: true,
			},
		},
		{
			name: "listing events works when env contains snapshots and 1..2 is provided as interval",
			cmd:  fmt.Sprintf(`log env %s --interval 1..2 %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			additionalConfig: listSnapshotsTestConfig{
				reportToEnv: true,
			},
		},
		{
			name: "listing events in interval 1..2 with --reverse works",
			cmd:  fmt.Sprintf(`log env %s --interval 1..2 --reverse %s`, suite.eventsEnvName, suite.defaultKosliArguments),
			additionalConfig: listSnapshotsTestConfig{
				reportToEnv: true,
			},
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil {
			if t.additionalConfig.(listSnapshotsTestConfig).reportToEnv {
				// send 2 reports to create 2 snapshots
				// every time this is called, will add 2 more snapshots and 2 more events
				ReportServerArtifactToEnv([]string{suite.firstArtifactPath}, suite.eventsEnvName, suite.T())
				ReportServerArtifactToEnv([]string{suite.firstArtifactPath, suite.secondArtifactPath}, suite.eventsEnvName, suite.T())
			}
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestLogEnvironmentCommandTestSuite(t *testing.T) {
	suite.Run(t, new(LogEnvironmentCommandTestSuite))
}
