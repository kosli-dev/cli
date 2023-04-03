package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CommitEvidenceSnykCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
}

func (suite *CommitEvidenceSnykCommandTestSuite) SetupTest() {
	suite.flowName = "snyk-test"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
}

func (suite *CommitEvidenceSnykCommandTestSuite) TestCommitEvidenceSnykCmd() {
	tests := []cmdTestCase{
		{
			name: "report Snyk test evidence works",
			cmd: `report evidence commit snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result --flows ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to commit: 239d7cee00ca341f124fa710fc694b67cdf8011b\n",
		},
		{
			name: "report Snyk test evidence works  when --evidence-url and --evidence-fingerprint are provided",
			cmd: `report evidence commit snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result --flows ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json 
					  --evidence-url https://example.com --evidence-fingerprint 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to commit: 239d7cee00ca341f124fa710fc694b67cdf8011b\n",
		},
		{
			name: "report Snyk scan evidence with non-existing scan-results",
			cmd: `report evidence commit snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result --flows ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/foo.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: open testdata/foo.json: no such file or directory\n",
		},
		{
			name: "report Snyk scan evidence with missing scan-results flag",
			cmd: `report evidence commit snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result --flows ` + suite.flowName + `
			          --build-url example.com` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"scan-results\" not set\n",
		},
		{
			name: "report Snyk scan evidence with missing name flag",
			cmd: `report evidence commit snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --flows ` + suite.flowName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing --flows flag",
			cmd: `report evidence commit snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to commit: 239d7cee00ca341f124fa710fc694b67cdf8011b\n",
		},
		{
			name: "report Snyk scan evidence with a missing build-url",
			cmd: `report evidence commit snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --flows ` + suite.flowName + `
			         --name snyk-result --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"build-url\" not set\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidenceSnykCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidenceSnykCommandTestSuite))
}
