package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArtifactEvidencePRGithubCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	flowName              string
	commitWithPR          string
}

func (suite *ArtifactEvidencePRGithubCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})

	suite.flowName = "github-pr"
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	suite.commitWithPR = testHelpers.GithubCommitWithPR()

	CreateFlow(suite.flowName, suite.T())
	CreateArtifact(suite.flowName, suite.artifactFingerprint, "foobar", suite.T())
}

func (suite *ArtifactEvidencePRGithubCommandTestSuite) TestArtifactEvidencePRGithubCmd() {
	tests := []cmdTestCase{
		{
			name: "report Github PR evidence works with new flags (fingerprint, name ...)",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit ` + suite.commitWithPR + suite.defaultKosliArguments,
			golden: "github pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --org is missing",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6 --api-token foo --host bar`,
			golden: "Error: --org is not set\n" +
				"Usage: kosli report evidence artifact pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --name is missing",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --github-org is missing",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"github-org\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --repository is missing",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --commit is missing",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when neither --fingerprint nor --artifact-type are set",
			cmd: `report evidence artifact pullrequest github artifactNameArg --name gh-pr --flow ` + suite.flowName + `
					  --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: either --artifact-type or --fingerprint must be specified\n" +
				"Usage: kosli report evidence artifact pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when commit does not exist",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c3123ab` + suite.defaultKosliArguments,
			golden: "Error: GET https://api.github.com/repos/kosli-dev/cli/commits/73d7fee2f31ade8e1a9c456c324255212c3123ab/pulls: 422 No commit found for SHA: 73d7fee2f31ade8e1a9c456c324255212c3123ab []\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --assert is used and commit has no PRs",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
					  --assert
			          --build-url example.com --github-org kosli-dev --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n",
		},
		{
			name: "report Github PR evidence does not fail when commit has no PRs",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "no pull requests found for given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n" +
				"github pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when the artifact does not exist in the server",
			cmd: `report evidence artifact pullrequest github testdata/file1 --artifact-type file --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit ` + suite.commitWithPR + suite.defaultKosliArguments,
			golden: "Error: Artifact with fingerprint '7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'github-pr' belonging to organization 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --artifact-type is unsupported",
			cmd: `report evidence artifact pullrequest github testdata/file1 --artifact-type unsupported --name gh-pr --flow ` + suite.flowName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: unsupported is not a supported artifact type\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --user-data is not found",
			cmd: `report evidence artifact pullrequest github --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --flow ` + suite.flowName + `
					  --user-data non-existing.json
			          --build-url example.com --github-org kosli-dev --repository cli --commit ` + suite.commitWithPR + suite.defaultKosliArguments,
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidencePRGithubCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidencePRGithubCommandTestSuite))
}
