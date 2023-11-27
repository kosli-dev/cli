package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ApprovalReportTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	flowName              string
	envName               string
	gitCommit             string
	artifactPath          string
}

type reportApprovalTestConfig struct {
	createSnapshot bool
}

func (suite *ApprovalReportTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	suite.flowName = "approval-test"
	suite.envName = "staging"
	suite.gitCommit = "993a9a6be532ed4e7a87aab4df90a7f1b3168d63"
	suite.artifactPath = "testdata/report.xml"
	suite.artifactFingerprint = "c6b02fb1708fbc46e63f5074c38d17852bfa0c7bcfcdd1f877e80476e3fcb74f"
	// _, suite.artifactFingerprint, _ = executeCommandC(fmt.Sprintf("kosli fingerprint" + suite.artifactPath + "--artifact-type file" + suite.defaultKosliArguments))

	CreateFlow(suite.flowName, suite.T())
	CreateArtifactWithCommit(suite.flowName, suite.artifactFingerprint, suite.artifactPath, suite.gitCommit, suite.T())
	CreateEnv(global.Org, suite.envName, "server", suite.T())
}

func (suite *ApprovalReportTestSuite) TestApprovalReportCmd() {
	tests := []cmdTestCase{
		{
			name: "report approval with a range of commits works ",
			cmd: `report approval --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + ` --repo-root ../.. 
			--newest-commit HEAD --oldest-commit HEAD~3` + suite.defaultKosliArguments,
			golden: fmt.Sprintf("approval created for artifact: %s\n", suite.artifactFingerprint),
		},
		{
			name: "report approval with an environment name works",
			cmd: `report approval --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + ` --repo-root ../.. 
			--newest-commit HEAD --oldest-commit HEAD~3` + ` --environment staging` + suite.defaultKosliArguments,
			golden: fmt.Sprintf("approval created for artifact: %s\n", suite.artifactFingerprint),
		},
		{
			wantError: true,
			name:      "report approval with no environment name or oldest commit fails",
			cmd: `report approval --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + ` --repo-root ../.. ` +
				suite.defaultKosliArguments,
			golden: "Error: at least one of --environment, --oldest-commit is required\n",
		},
		// For the next test we have to
		// - create a flow
		// - create an artifact in that flow with git commit HEAD~5
		// - (create an environment)
		// - create snapshot that contains this artifact
		{
			name: "report approval with an environment name and no oldest-commit and no newest-commit works",
			cmd: `report approval --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + ` --repo-root ../.. ` +
				` --environment ` + suite.envName + suite.defaultKosliArguments,
			golden: fmt.Sprintf("approval created for artifact: %s\n", suite.artifactFingerprint),
			additionalConfig: reportApprovalTestConfig{
				createSnapshot: true,
			},
		},

		// Here is a case we need to investigate how to test:
		// - Create approval with '--newest-commit HEAD~5', '--oldest-commit HEAD~7' and '--environment staging',
		//   then create approval only with '--environment staging',
		// 	 the resulting payload should contain a commit list of 4 commits, and an oldest_commit of HEAD~5
	}
	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(reportApprovalTestConfig).createSnapshot {
			ReportServerArtifactToEnv([]string{suite.artifactPath}, suite.envName, suite.T())
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
	// runTestCmd(suite.T(), tests)
}

func TestApprovalReportCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ApprovalReportTestSuite))
}
