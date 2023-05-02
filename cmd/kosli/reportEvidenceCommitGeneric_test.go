package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CommitEvidenceGenericCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowNames             string
}

func (suite *CommitEvidenceGenericCommandTestSuite) SetupTest() {
	suite.flowNames = "generic-evidence"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowNames, suite.T())
}

func (suite *CommitEvidenceGenericCommandTestSuite) TestCommitEvidenceGenericCommandCmd() {
	evidenceName := "manual-test"
	tests := []cmdTestCase{
		{
			name: "report Generic test evidence works without files",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant --description "some description" %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works when --evidence-url and --evidence-fingerprint are provided",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant --description "some description" 
					  --evidence-url https://example.com --evidence-fingerprint 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0 %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works when --evidence-paths is provided and contains a single file",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant --description "some description" 
					  --evidence-paths testdata/file1 %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works when --evidence-paths is provided and contains a single directory",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant --description "some description" 
					  --evidence-paths testdata/folder1 %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works when --evidence-paths is provided and contains a file and a dir",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant --description "some description" 
					  --evidence-paths testdata/folder1,testdata/file1 %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			wantError: true,
			name:      "report Generic test evidence fails when --evidence-paths is provided and contains a non-existing file",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant --description "some description" 
					  --evidence-paths non-existing.txt %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: "Error: stat non-existing.txt: no such file or directory\n",
		},
		{
			name: "report Generic test evidence works when neither of --description nor --user-data provided",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works when neither of --description, --user-data or --compliant is provided",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com %s`, evidenceName, suite.flowNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence fails if --name is missing",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --flows %s
			          --build-url example.com %s`, suite.flowNames, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Generic test evidence fails if --commit is missing",
			cmd: fmt.Sprintf(`report evidence commit generic --name %s --flows %s
			          --build-url example.com --compliant --description "some description" %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"commit\" not set\n",
		},
		{
			name: "report Generic test evidence works if --flows flag is missing",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s
			          --build-url example.com --compliant --description "some description" %s`,
				evidenceName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence fails if --build-url is missing",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
					--compliant --description "some description" %s`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"build-url\" not set\n",
		},
		{
			name: "report Generic test evidence fails if user-data is non-existing file",
			cmd: fmt.Sprintf(`report evidence commit generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --flows %s
			          --build-url example.com --compliant --description "some description" %s --user-data non-existing-file`,
				evidenceName, suite.flowNames, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: open non-existing-file: no such file or directory\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidenceGenericCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidenceGenericCommandTestSuite))
}
