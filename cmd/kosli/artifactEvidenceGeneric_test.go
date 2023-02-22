package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArtifactEvidenceGenericCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	pipelineName          string
}

func (suite *ArtifactEvidenceGenericCommandTestSuite) SetupTest() {
	suite.pipelineName = "generic-evidence"
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)

	CreateFlow(suite.pipelineName, suite.T())
	CreateArtifact(suite.pipelineName, suite.artifactFingerprint, "FooBar_1", suite.T())

	tests := []cmdTestCase{
		{
			name: "create second artifact",
			cmd: `report artifact testdata --git-commit 0fc1ba9876f91b215679f3649b8668085d820ab5 --artifact-type dir ` + `
			          --flow ` + suite.pipelineName + ` --build-url www.yr.no --commit-url www.nrk.no --repo-root ../..` + suite.defaultKosliArguments,
		},
	}
	runTestCmd(suite.T(), tests)
}

func (suite *ArtifactEvidenceGenericCommandTestSuite) TestArtifactEvidenceGenericCommandCmd() {
	evidenceName := "manual-test"
	tests := []cmdTestCase{
		{
			name: "report Generic test evidence works",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description" %s`,
				suite.artifactFingerprint, evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when neither of --description nor --user-data provided",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant %s`,
				suite.artifactFingerprint, evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when neither of --description, --user-data or --compliant is provided",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com %s`,
				suite.artifactFingerprint, evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence fails if --name is missing",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --flow %s
			          --build-url example.com %s`,
				suite.artifactFingerprint, suite.pipelineName, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Generic test evidence fails if --fingerprint and --artifact-type are missing ",
			cmd: fmt.Sprintf(`report evidence artifact generic --name %s --flow %s
			          --build-url example.com %s`,
				evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			wantError: true,
		},
		{
			name: "report Generic test evidence works when --artifact-type is provided",
			cmd: fmt.Sprintf(`report evidence artifact generic testdata --artifact-type dir --name %s --flow %s
			          --build-url example.com %s`,
				evidenceName, suite.pipelineName, suite.defaultKosliArguments),
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidenceGenericCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidenceGenericCommandTestSuite))
}
