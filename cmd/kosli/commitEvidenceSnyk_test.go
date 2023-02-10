package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CommitEvidenceSnykCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	pipelineName          string
}

func (suite *CommitEvidenceSnykCommandTestSuite) SetupTest() {
	suite.pipelineName = "snyk-test"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)

	kosliClient = requests.NewKosliClient(1, false, logger)

	CreatePipeline(suite.pipelineName, suite.T())
}

func (suite *CommitEvidenceSnykCommandTestSuite) TestCommitEvidenceSnykCmd() {
	tests := []cmdTestCase{
		{
			name: "report Snyk test evidence works",
			cmd: `commit report evidence snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result --pipelines ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to commit: 239d7cee00ca341f124fa710fc694b67cdf8011b\n",
		},
		{
			name: "report Snyk scan evidence with non-existing scan-results",
			cmd: `commit report evidence snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result --pipelines ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/foo.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: open testdata/foo.json: no such file or directory\n",
		},
		{
			name: "report Snyk scan evidence with missing scan-results flag",
			cmd: `commit report evidence snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result --pipelines ` + suite.pipelineName + `
			          --build-url example.com` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"scan-results\" not set\n",
		},
		{
			name: "report Snyk scan evidence with missing name flag",
			cmd: `commit report evidence snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --pipelines ` + suite.pipelineName + `
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report Snyk scan evidence with a missing pipelines flag",
			cmd: `commit report evidence snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --name snyk-result
			          --build-url example.com --scan-results testdata/snyk_scan_example.json` + suite.defaultKosliArguments,
			golden: "snyk scan evidence is reported to commit: 239d7cee00ca341f124fa710fc694b67cdf8011b\n",
		},
		{
			name: "report Snyk scan evidence with a missing build-url",
			cmd: `commit report evidence snyk --commit 239d7cee00ca341f124fa710fc694b67cdf8011b --pipelines ` + suite.pipelineName + `
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
