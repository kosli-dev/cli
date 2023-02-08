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
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
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

func (suite *ArtifactEvidenceSnykCommandTestSuite) TestArtifactEvidenceSnykCmd() {

	tests := []cmdTestCase{
		{
			name: "report Snyk test evidence works (using --fingerprint)",
			cmd: `pipeline artifact report evidence snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --pipeline ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk test evidence works (using --artifact-type)",
			cmd: `pipeline artifact report evidence snyk testdata/file1 --artifact-type file --name snyk-result --pipeline ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk scan evidence with non-existing scan-results",
			cmd: `pipeline artifact report evidence snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --pipeline ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/foo.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: open testdata/foo.json: no such file or directory\n",
		},
		{
			name: "report Snyk scan evidence with missing scan-results flag",
			cmd: `pipeline artifact report evidence snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --pipeline ` + suite.pipelineName + `
			          --build-url example.com` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"scan-results\" not set\n",
		},
		{
			name: "report Snyk scan evidence with missing name flag",
			cmd: `pipeline artifact report evidence snyk --fingerprint ` + suite.artifactFingerprint + ` --pipeline ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing pipeline",
			cmd: `pipeline artifact report evidence snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"pipeline\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing build-url",
			cmd: `pipeline artifact report evidence snyk --fingerprint ` + suite.artifactFingerprint + ` --pipeline ` + suite.pipelineName + `
			         --name snyk-result --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"build-url\" not set\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidenceSnykCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidenceSnykCommandTestSuite))
}
