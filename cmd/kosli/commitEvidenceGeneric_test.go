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
type CommitEvidenceGenericCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	pipelineNames         string
}

func (suite *CommitEvidenceGenericCommandTestSuite) SetupTest() {
	suite.pipelineNames = "generic-evidence"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	kosliClient = requests.NewKosliClient(1, false, log.NewStandardLogger())

	CreatePipeline(suite.pipelineNames, suite.T())
}

func (suite *CommitEvidenceGenericCommandTestSuite) TestCommitEvidenceGenericCommandCmd() {
	evidenceName := "manual-test"
	tests := []cmdTestCase{
		{
			name: "report Generic test evidence works",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --pipelines %s
			          --build-url example.com --compliant --description "some description" %s`,
				evidenceName, suite.pipelineNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works with evidence-url and evidence-fingerprint flags",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --pipelines %s
			          --build-url example.com --compliant --description "some description"
					  --evidence-url example.com --evidence-fingerprint 1234 %s`,
				evidenceName, suite.pipelineNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works when neither of --description nor --user-data provided",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --pipelines %s
			          --build-url example.com --compliant %s`,
				evidenceName, suite.pipelineNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence works when neither of --description, --user-data or --compliant is provided",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --pipelines %s
			          --build-url example.com %s`, evidenceName, suite.pipelineNames, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence fails if --name is missing",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --pipelines %s
			          --build-url example.com %s`, suite.pipelineNames, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Generic test evidence fails if --commit is missing",
			cmd: fmt.Sprintf(`commit report evidence generic --name %s --pipelines %s
			          --build-url example.com --compliant --description "some description" %s`,
				evidenceName, suite.pipelineNames, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"commit\" not set\n",
		},
		{
			name: "report Generic test evidence works if --pipelines flag is missing",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s
			          --build-url example.com --compliant --description "some description" %s`,
				evidenceName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to commit: af28ccdeffdfa67f5c5a88be209e94cc4742de3c\n", evidenceName),
		},
		{
			name: "report Generic test evidence fails if --build-url is missing",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --pipelines %s
					--compliant --description "some description" %s`,
				evidenceName, suite.pipelineNames, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"build-url\" not set\n",
		},
		{
			name: "report Generic test evidence fails if user-data is non-existing file",
			cmd: fmt.Sprintf(`commit report evidence generic --commit af28ccdeffdfa67f5c5a88be209e94cc4742de3c --name %s --pipelines %s
			          --build-url example.com --compliant --description "some description" %s --user-data non-existing-file`,
				evidenceName, suite.pipelineNames, suite.defaultKosliArguments),
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
