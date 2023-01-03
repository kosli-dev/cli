package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArtifactEvidenceSnykCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	pipelineName          string
}

func (suite *ArtifactEvidenceSnykCommandTestSuite) SetupTest() {
	suite.defaultKosliArguments = " -H http://localhost:8001 --owner docs-cmd-test-user -a eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa2"
	suite.pipelineName = "snyk-test"
	tests := []cmdTestCase{
		{
			name: "create a pipeline",
			cmd:  "pipeline declare --pipeline " + suite.pipelineName + " --description \"my snyk pipeline\" " + suite.defaultKosliArguments,
		},
		{
			name: "create an artifact",
			cmd: `pipeline artifact report creation FooBar_1 --git-commit HEAD --sha256 ` + suite.artifactFingerprint + `
			          --pipeline ` + suite.pipelineName + ` --build-url www.yr.no --commit-url www.nrk.no --repo-root ../..` + suite.defaultKosliArguments,
		},
	}
	runTestCmd(suite.T(), tests)
}

func (suite *ArtifactEvidenceSnykCommandTestSuite) TestArtifactEvidenceSnykCommandCmd() {

	tests := []cmdTestCase{
		{
			name: "report Snyk test evidence works",
			cmd: `pipeline artifact report evidence snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --pipeline ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		// {
		// 	name: "report JUnit test evidence with non-existing results dir",
		// 	cmd: `pipeline artifact report evidence junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result --pipeline ` + suite.pipelineName + `
		// 	          --build-url example.com --results-dir foo` + suite.defaultKosliArguments,
		// 	wantError: true,
		// },
		// {
		// 	name: "report JUnit test evidence with a results dir that does not contain any results",
		// 	cmd: `pipeline artifact report evidence junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result --pipeline ` + suite.pipelineName + `
		// 	          --build-url example.com --results-dir testdata/folder1` + suite.defaultKosliArguments,
		// 	wantError: true,
		// 	golden:    "Error: no tests found in testdata/folder1 directory\n",
		// },
		// {
		// 	name: "report JUnit test evidence with missing name flag",
		// 	cmd: `pipeline artifact report evidence junit --fingerprint ` + suite.artifactFingerprint + ` --pipeline ` + suite.pipelineName + `
		// 	          --build-url example.com --results-dir testdata/folder1` + suite.defaultKosliArguments,
		// 	wantError: true,
		// },
		// {
		// 	name: "report JUnit test evidence with missing build-url",
		// 	cmd: `pipeline artifact report evidence junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result --pipeline ` + suite.pipelineName + `
		// 	          --results-dir testdata/folder1` + suite.defaultKosliArguments,
		// 	wantError: true,
		// },
		// {
		// 	name: "report JUnit test evidence with a missing pipeline",
		// 	cmd: `pipeline artifact report evidence junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result
		// 	          --build-url example.com --results-dir testdata/folder1` + suite.defaultKosliArguments,
		// 	wantError: true,
		// },
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidenceSnykCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidenceSnykCommandTestSuite))
}
