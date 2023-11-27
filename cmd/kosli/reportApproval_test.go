package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/stretchr/testify/require"
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
	t := suite.T()

	gitView, err := gitview.New("../..")
	require.NoError(t, err, "Failed to create gitview")

	suite.gitCommit, err = gitView.ResolveRevision("HEAD~5")
	require.NoError(t, err, "Failed to get HEAD~5")

	suite.artifactPath = "testdata/report.xml"
	// We cannot get the digest of the file by running the 'kosli fingerprint' command
	// by using executeCommandC() because this function overwrites the global options
	suite.artifactFingerprint, err = digest.FileSha256(suite.artifactPath)
	require.NoError(t, err, "Failed to calculate fingerprint")

	CreateFlow(suite.flowName, t)
	CreateArtifactWithCommit(suite.flowName, suite.artifactFingerprint, suite.artifactPath, suite.gitCommit, t)
	CreateEnv(global.Org, suite.envName, "server", t)
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
			name:      "report approval with no environment name and no oldest commit fails",
			cmd: `report approval --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + ` --repo-root ../.. ` +
				suite.defaultKosliArguments,
			golden: "Error: at least one of --environment, --oldest-commit is required\n",
		},
		{
			name: "report approval with an environment name and no oldest-commit and no newest-commit works",
			cmd: `report approval --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + ` --repo-root ../.. ` +
				` --environment ` + suite.envName + suite.defaultKosliArguments,
			golden: fmt.Sprintf("approval created for artifact: %s\n", suite.artifactFingerprint),
			additionalConfig: reportApprovalTestConfig{
				createSnapshot: true,
			},
		},
	}
	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(reportApprovalTestConfig).createSnapshot {
			ReportServerArtifactToEnv([]string{suite.artifactPath}, suite.envName, suite.T())
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

func TestApprovalReportCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ApprovalReportTestSuite))
}
