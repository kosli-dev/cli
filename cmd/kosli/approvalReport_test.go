package main

import (
	"fmt"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/suite"
)

type ApprovalReportTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	pipelineName          string
}

func (suite *ApprovalReportTestSuite) SetupTest() {

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0"
	suite.pipelineName = "approval-test"

	logger = log.NewStandardLogger()
	kosliClient = requests.NewKosliClient(1, false, logger)

	CreatePipeline(suite.pipelineName, suite.T())
	CreateArtifact(suite.pipelineName, suite.artifactFingerprint, "foobar", suite.T())

}

func (suite *ApprovalReportTestSuite) TestApprovalReportCmd() {
	tests := []cmdTestCase{
		{
			name: "report approval with a range of commits works ",
			cmd: `pipeline approval report --sha256 ` + suite.artifactFingerprint + ` --pipeline ` + suite.pipelineName + ` --repo-root ../.. 
			--newest-commit HEAD --oldest-commit HEAD~3` + suite.defaultKosliArguments,
			golden: fmt.Sprintf("approval created for artifact: %s\n", suite.artifactFingerprint),
		},
	}
	runTestCmd(suite.T(), tests)
}

func TestApprovalReportCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ApprovalReportTestSuite))
}