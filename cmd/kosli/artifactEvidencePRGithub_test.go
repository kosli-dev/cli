package main

import (
	"fmt"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
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
	pipelineName          string
}

func (suite *ArtifactEvidencePRGithubCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})

	suite.pipelineName = "github-pr"
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	kosliClient = requests.NewKosliClient(1, false, log.NewStandardLogger())

	CreatePipeline(suite.pipelineName, suite.T())
	CreateArtifact(suite.pipelineName, suite.artifactFingerprint, "foobar", suite.T())
}

func (suite *ArtifactEvidencePRGithubCommandTestSuite) TestArtifactEvidencePRGithubCmd() {
	tests := []cmdTestCase{
		{
			name: "report Github PR evidence works with new flags (fingerprint, name ...)",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "github pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Github PR evidence works with evidence-url and evidence-fingerprint",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6
					  --evidence-url yr.no --evidence-fingerprint deadbeef ` + suite.defaultKosliArguments,
			golden: "github pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report Github PR evidence works with deprecated flags",
			cmd: `pipeline artifact report evidence github-pullrequest --sha256 ` + suite.artifactFingerprint + ` --evidence-type gh-pr --pipeline ` + suite.pipelineName + `
			          --description text --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Flag --sha256 has been deprecated, use --fingerprint instead\n" +
				"Flag --evidence-type has been deprecated, use --name instead\n" +
				"Flag --description has been deprecated, description is no longer used\n" +
				"github pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --owner is missing",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6 --api-token foo --host bar`,
			golden: "Error: --owner is not set\n" +
				"Usage: kosli pipeline artifact report evidence github-pullrequest [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when both --name and --evidence-type are missing",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --pipeline ` + suite.pipelineName + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: at least one of --name, --evidence-type is required\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --github-org is missing",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"github-org\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --repository is missing",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --commit is missing",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository cli` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when neither --fingerprint nor --artifact-type are set",
			cmd: `pipeline artifact report evidence github-pullrequest artifactNameArg --name gh-pr --pipeline ` + suite.pipelineName + `
					  --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: either --artifact-type or --sha256 must be specified\n" +
				"Usage: kosli pipeline artifact report evidence github-pullrequest [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when commit does not exist",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c3123ab` + suite.defaultKosliArguments,
			golden: "Error: GET https://api.github.com/repos/kosli-dev/cli/commits/73d7fee2f31ade8e1a9c456c324255212c3123ab/pulls: 422 No commit found for SHA: 73d7fee2f31ade8e1a9c456c324255212c3123ab []\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --assert is used and commit has no PRs",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
					  --assert
			          --build-url example.com --github-org kosli-dev --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n",
		},
		{
			name: "report Github PR evidence does not fail when commit has no PRs",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "no pull requests found for given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n" +
				"github pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when the artifact does not exist in the server",
			cmd: `pipeline artifact report evidence github-pullrequest testdata/file1 --artifact-type file --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: Artifact with fingerprint '7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in pipeline 'github-pr' belonging to 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --artifact-type is unsupported",
			cmd: `pipeline artifact report evidence github-pullrequest testdata/file1 --artifact-type unsupported --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: unsupported is not a supported artifact type\n",
		},
		{
			wantError: true,
			name:      "report Github PR evidence fails when --user-data is not found",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
					  --user-data non-existing.json
			          --build-url example.com --github-org kosli-dev --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
		{
			name: "report Github PR evidence works with --repository=owner/repo",
			cmd: `pipeline artifact report evidence github-pullrequest --fingerprint ` + suite.artifactFingerprint + ` --name gh-pr --pipeline ` + suite.pipelineName + `
			          --build-url example.com --github-org kosli-dev --repository kosli-dev/cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "github pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ArtifactEvidencePRGithubCommandTestSuite) TestAssertPRGithubCmd() {
	tests := []cmdTestCase{
		{
			name: "assert Github PR evidence passes when commit has a PR in github",
			cmd: `assert github-pullrequest --github-org kosli-dev --repository cli 
			--commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Github for commit: 73d7fee2f31ade8e1a9c456c324255212c30c2a6\n",
		},
		{
			wantError: true,
			name:      "assert Github PR evidence fails when commit has no PRs in github",
			cmd: `assert github-pullrequest --github-org kosli-dev --repository cli 
			--commit 19aab7f063147614451c88969602a10afbabb43d` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 19aab7f063147614451c88969602a10afbabb43d\n",
		},
		{
			wantError: true,
			name:      "assert Github PR evidence fails when commit does not exist",
			cmd: `assert github-pullrequest --github-org kosli-dev --repository cli 
			--commit 19aab7f063147614451c88969602a10afba123ab` + suite.defaultKosliArguments,
			golden: "Error: GET https://api.github.com/repos/kosli-dev/cli/commits/19aab7f063147614451c88969602a10afba123ab/pulls: 422 No commit found for SHA: 19aab7f063147614451c88969602a10afba123ab []\n",
		},
		{
			name: "assert Github PR evidence works with --repository=owner/repo",
			cmd: `assert github-pullrequest --github-org kosli-dev --repository kosli-dev/cli 
			--commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "found [1] pull request(s) in Github for commit: 73d7fee2f31ade8e1a9c456c324255212c30c2a6\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidencePRGithubCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidencePRGithubCommandTestSuite))
}
