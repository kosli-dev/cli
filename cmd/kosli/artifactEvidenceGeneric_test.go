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
	kosliClient = requests.NewKosliClient(1, false, log.NewStandardLogger())

	CreatePipeline(suite.pipelineName, suite.T())
	CreateArtifact(suite.pipelineName, suite.artifactFingerprint, "FooBar_1", suite.T())

	tests := []cmdTestCase{
		{
			name: "create second artifact",
			cmd: `pipeline artifact report creation testdata --git-commit 6ef6fc37c373922eecd4e823cf2633326790cfe8 --artifact-type dir ` + `
			          --pipeline ` + suite.pipelineName + ` --build-url www.yr.no --commit-url www.nrk.no --repo-root ../..` + suite.defaultKosliArguments,
		},
	}
	runTestCmd(suite.T(), tests)
}

func (suite *ArtifactEvidenceGenericCommandTestSuite) TestArtifactEvidenceGenericCommandCmd() {
	evidenceName := "manual-test"
	tests := []cmdTestCase{
		{
			name: "report Generic test evidence works",
			cmd: fmt.Sprintf(`pipeline artifact report evidence generic --fingerprint %s --name %s --pipeline %s
			          --build-url example.com --compliant --description "some description" %s`,
				suite.artifactFingerprint, evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when neither of --description nor --user-data provided",
			cmd: fmt.Sprintf(`pipeline artifact report evidence generic --fingerprint %s --name %s --pipeline %s
			          --build-url example.com --compliant %s`,
				suite.artifactFingerprint, evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when neither of --description, --user-data or --compliant is provided",
			cmd: fmt.Sprintf(`pipeline artifact report evidence generic --fingerprint %s --name %s --pipeline %s
			          --build-url example.com %s`,
				suite.artifactFingerprint, evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence fails if both --name and --evidence-type are missing",
			cmd: fmt.Sprintf(`pipeline artifact report evidence generic --fingerprint %s --pipeline %s
			          --build-url example.com %s`,
				suite.artifactFingerprint, suite.pipelineName, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: at least one of --name, --evidence-type is required\n",
		},
		{
			name: "report Generic test evidence fails if --sha256, --fingerprint and --artifact-type are missing ",
			cmd: fmt.Sprintf(`pipeline artifact report evidence generic --name %s --pipeline %s
			          --build-url example.com %s`,
				evidenceName, suite.pipelineName, suite.defaultKosliArguments),
			wantError: true,
		},
		{
			name: "report Generic test evidence works when --artifact-type is provided",
			cmd: fmt.Sprintf(`pipeline artifact report evidence generic testdata --artifact-type dir --name %s --pipeline %s
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