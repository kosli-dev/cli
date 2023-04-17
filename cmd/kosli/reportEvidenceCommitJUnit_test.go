package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CommitEvidenceJUnitCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowNames             string
}

func (suite *CommitEvidenceJUnitCommandTestSuite) SetupTest() {
	suite.flowNames = "junit-test"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowNames, suite.T())
}

func (suite *CommitEvidenceJUnitCommandTestSuite) TestCommitEvidenceJUnitCommandCmd() {
	tests := []cmdTestCase{
		{
			name: "report JUnit test evidence works",
			cmd: `report evidence commit junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result --flows ` + suite.flowNames + `
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n",
		},
		{
			name: "report JUnit test evidence works when --evidence-url and --evidence-fingerprint are provided",
			cmd: `report evidence commit junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result --flows ` + suite.flowNames + `
			          --build-url example.com --results-dir testdata 
					  --evidence-url https://example.com --evidence-fingerprint 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n",
		},
		{
			name: "report JUnit test evidence with non-existing results dir",
			cmd: `report evidence commit junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result --flows ` + suite.flowNames + `
			          --build-url example.com --results-dir foo` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: lstat foo: no such file or directory\n",
		},
		{
			name: "report JUnit test evidence with a results dir that does not contain any results",
			cmd: `report evidence commit junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result --flows ` + suite.flowNames + `
			          --build-url example.com --results-dir testdata/folder1` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: no tests found in testdata/folder1 directory\n",
		},
		{
			name: "report JUnit test evidence with missing name flag",
			cmd: `report evidence commit junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --flows ` + suite.flowNames + `
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report JUnit test evidence with a missing --flows flag",
			cmd: `report evidence commit junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name junit-result
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n",
		},
		{
			name: "report JUnit test evidence with a missing build-url",
			cmd: `report evidence commit junit --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --flows ` + suite.flowNames + `
					--name junit-result --results-dir testdata` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"build-url\" not set\n",
		},
		{
			name: "report JUnit test evidence with a missing commit flag",
			cmd: `report evidence commit junit --flows ` + suite.flowNames + `
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
