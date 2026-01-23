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
			          --build-url http://www.example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			goldenRegex: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk test evidence works when --evidence-url and --evidence-fingerprint are provided",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url http://www.example.com --scan-results testdata/snyk_scan_example.json
					  --evidence-url https://example.com --evidence-fingerprint ` + suite.artifactFingerprint + suite.defaultKosliArguments,
			goldenRegex: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk test evidence works (using --artifact-type)",
			cmd: `report evidence artifact snyk testdata/file1 --artifact-type file --name snyk-result --flow ` + suite.flowName + `
			          --build-url http://www.example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			goldenRegex: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk scan evidence with non-existing scan-results fails",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url http://www.example.com --scan-results testdata/foo.json` + suite.defaultKosliArguments,
			wantError:   true,
			goldenRegex: "Error: failed to parse Snyk results file \\[testdata/foo.json\\]. Failed to parse as Sarif: open testdata/foo.json: no such file or directory. Fallen back to parse Snyk Json, but also failed: the provided file path doesn't have a file\n",
		},
		{
			name: "report Snyk scan evidence with missing scan-results flag fails",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url http://www.example.com` + suite.defaultKosliArguments,
			wantError:   true,
			goldenRegex: "Error: required flag\\(s\\) \"scan-results\" not set\n",
		},
		{
			name: "report Snyk scan evidence with missing name flag fails",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
			          --build-url http://www.example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError:   true,
			goldenRegex: "Error: required flag\\(s\\) \"name\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing flow fails",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result
			          --build-url http://www.example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError:   true,
			goldenRegex: "Error: required flag\\(s\\) \"flow\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing build-url fails",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
			         --name snyk-result --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError:   true,
			goldenRegex: "Error: required flag\\(s\\) \"build-url\" not set\n",
		},
		{
			name: "report Snyk test evidence works with sarif snyk results and uploading is enabled",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url http://www.example.com --scan-results testdata/snyk_sarif.json` + suite.defaultKosliArguments,
			goldenRegex: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Snyk test evidence works with sarif snyk results and uploading is disabled",
			cmd: `report evidence artifact snyk --fingerprint ` + suite.artifactFingerprint + ` --name snyk-result --flow ` + suite.flowName + `
			          --build-url http://www.example.com --scan-results testdata/snyk_sarif.json --upload-results=false` + suite.defaultKosliArguments,
			goldenRegex: "snyk scan evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidenceSnykCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidenceSnykCommandTestSuite))
}
