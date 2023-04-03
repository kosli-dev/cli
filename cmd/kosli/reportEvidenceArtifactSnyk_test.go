package main

import (
	"fmt"
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
	flowName              string
}

func (suite *ArtifactEvidenceSnykCommandTestSuite) SetupTest() {
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	suite.flowName = "snyk-test"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	CreateArtifact(suite.flowName, suite.artifactFingerprint, "FooBar_1", suite.T())
}

func (suite *ArtifactEvidenceSnykCommandTestSuite) TestArtifactEvidenceSnykCmd() {
	tests := []cmdTestCase{
		{
			name: "report Snyk test evidence works (using --fingerprint)",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk test evidence works when --evidence-url and --evidence-fingerprint are provided",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json
					  --evidence-url https://example.com --evidence-fingerprint ` + suite.artifactFingerprint + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk test evidence works (using --artifact-type)",
			cmd: `report evidence artifact snyk testdata/file1 --artifact-type file --name snyk-result --flow ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk scan evidence with non-existing scan-results",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/foo.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: open testdata/foo.json: no such file or directory\n",
		},
		{
			name: "report Snyk scan evidence with missing scan-results flag",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url example.com` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"scan-results\" not set\n",
		},
		{
			name: "report Snyk scan evidence with missing name flag",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing flow",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing build-url",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
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
