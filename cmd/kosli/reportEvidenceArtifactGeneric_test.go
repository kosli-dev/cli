package main

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArtifactEvidenceGenericCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	flowName              string
}

func (suite *ArtifactEvidenceGenericCommandTestSuite) SetupTest() {
	suite.flowName = "generic-evidence"
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	t := suite.T()
	CreateFlow(suite.flowName, t)
	CreateArtifact(suite.flowName, suite.artifactFingerprint, "FooBar_1", t)

	repo, err := git.PlainOpen("../..")
	require.NoError(t, err, "failed to open git repository at %s: %v", "../..", err)
	commitPointer, err := repo.ResolveRevision(plumbing.Revision("HEAD~1"))
	require.NoError(t, err, "failed to resolve revision %s: %v", "HEAD~1", err)
	commitHash := commitPointer.String()
	tests := []cmdTestCase{
		{
			name: "create second artifact",
			cmd: `report artifact testdata --git-commit ` + commitHash + ` --artifact-type dir ` + `
			          --flow ` + suite.flowName + ` --build-url www.yr.no --commit-url www.nrk.no --repo-root ../..` + suite.defaultKosliArguments,
		},
	}
	runTestCmd(suite.T(), tests)
}

func (suite *ArtifactEvidenceGenericCommandTestSuite) TestArtifactEvidenceGenericCommandCmd() {
	evidenceName := "manual-test"
	tests := []cmdTestCase{
		{
			name: "report Generic test evidence works without --evidence-paths",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description" %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when --evidence-url and --evidence-fingerprint are provided",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description" 
					  --evidence-url https://example.com --evidence-fingerprint %s %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.artifactFingerprint, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works with --evidence-paths that contains a single file",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description"
					  --evidence-paths testdata/file1 %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works with --evidence-paths that contains a single dir",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description"
					  --evidence-paths testdata/folder1 %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works with --evidence-paths that contains multiple paths",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description"
					  --evidence-paths testdata/file1,testdata/folder1 %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when multiple --evidence-paths include duplicates",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description" 
					  --evidence-paths testdata/file1,testdata/folder1,testdata/file1 %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when neither of --description nor --user-data provided",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			name: "report Generic test evidence works when neither of --description, --user-data or --compliant is provided",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("generic evidence '%s' is reported to artifact: %s\n", evidenceName, suite.artifactFingerprint),
		},
		{
			wantError: true,
			name:      "report Generic test evidence fails when providing --evidence-paths that does not exist",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --name %s --flow %s
			          --build-url example.com --compliant --description "some description"
					  --evidence-paths non-existing %s`,
				suite.artifactFingerprint, evidenceName, suite.flowName, suite.defaultKosliArguments),
			golden: "Error: stat non-existing: no such file or directory\n",
		},
		{
			name: "report Generic test evidence fails if --name is missing",
			cmd: fmt.Sprintf(`report evidence artifact generic --fingerprint %s --flow %s
			          --build-url example.com %s`,
				suite.artifactFingerprint, suite.flowName, suite.defaultKosliArguments),
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Generic test evidence fails if --fingerprint and --artifact-type are missing ",
			cmd: fmt.Sprintf(`report evidence artifact generic --name %s --flow %s
			          --build-url example.com %s`,
				evidenceName, suite.flowName, suite.defaultKosliArguments),
			wantError: true,
		},
		{
			name: "report Generic test evidence works when --artifact-type is provided",
			cmd: fmt.Sprintf(`report evidence artifact generic testdata --artifact-type dir --name %s --flow %s
			          --build-url example.com %s`,
				evidenceName, suite.flowName, suite.defaultKosliArguments),
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidenceGenericCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidenceGenericCommandTestSuite))
}
