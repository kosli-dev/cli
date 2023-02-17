package main

import (
	"fmt"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CommitEvidenceJUnitCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	pipelineNames         string
}

func (suite *CommitEvidenceJUnitCommandTestSuite) SetupTest() {
	suite.pipelineNames = "junit-test"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	kosliClient = requests.NewKosliClient(1, false, log.NewStandardLogger())

	CreateFlow(suite.pipelineNames, suite.T())
}

func (suite *CommitEvidenceJUnitCommandTestSuite) TestCommitEvidenceJUnitCommandCmd() {
	tests := []cmdTestCase{
		{
			name: "report JUnit test evidence works",
			cmd: `commit report evidence junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result --pipelines ` + suite.pipelineNames + `
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n",
		},
		{
			name: "report JUnit test evidence with non-existing results dir",
			cmd: `commit report evidence junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result --pipelines ` + suite.pipelineNames + `
			          --build-url example.com --results-dir foo` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: lstat foo: no such file or directory\n",
		},
		{
			name: "report JUnit test evidence with a results dir that does not contain any results",
			cmd: `commit report evidence junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result --pipelines ` + suite.pipelineNames + `
			          --build-url example.com --results-dir testdata/folder1` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: no tests found in testdata/folder1 directory\n",
		},
		{
			name: "report JUnit test evidence with missing name flag",
			cmd: `commit report evidence junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --pipelines ` + suite.pipelineNames + `
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report JUnit test evidence with a missing pipelines flag",
			cmd: `commit report evidence junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n",
		},
		{
			name: "report JUnit test evidence with a missing build-url",
			cmd: `commit report evidence junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --pipelines ` + suite.pipelineNames + `
					--name junit-result --results-dir testdata` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"build-url\" not set\n",
		},
		{
			name: "report JUnit test evidence with a missing commit flag",
			cmd: `commit report evidence junit --pipelines ` + suite.pipelineNames + `
					--build-url example.com --name junit-result --results-dir testdata` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"commit\" not set\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidenceJUnitCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidenceJUnitCommandTestSuite))
}
