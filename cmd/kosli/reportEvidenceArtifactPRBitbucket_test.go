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
type ArtifactEvidencePRBitbucketCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	flowName              string
}

func (suite *ArtifactEvidencePRBitbucketCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{"KOSLI_BITBUCKET_PASSWORD"})

	suite.flowName = "bitbucket-pr"
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.Suite.T())
	CreateArtifact(suite.flowName, suite.artifactFingerprint, "foobar", suite.Suite.T())
}

func (suite *ArtifactEvidencePRBitbucketCommandTestSuite) TestArtifactEvidencePRBitbucketCmd() {
	tests := []cmdTestCase{
		{
			name: "report Bitbucket PR evidence works with new flags (fingerprint, name ...)",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request evidence is reported to artifact: .*",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --org is missing",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f --api-token foo --host bar`,
			goldenRegex: "Error: --org is not set\n" +
				"Usage: kosli report evidence artifact pullrequest bitbucket \\[IMAGE-NAME | FILE-PATH | DIR-PATH\\] \\[flags\\]\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --name is missing",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: required flag\\(s\\) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --repository is missing",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: required flag\\(s\\) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --commit is missing",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test` + suite.defaultKosliArguments,
			goldenRegex: "Error: required flag\\(s\\) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when neither --fingerprint nor --artifact-type are set",
			cmd: `report evidence artifact pullrequest bitbucket artifactNameArg --name bb-pr --flow ` + suite.flowName + `
					  --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: either --artifact-type or --fingerprint must be specified\n" +
				"Usage: kosli report evidence artifact pullrequest bitbucket \\[IMAGE-NAME | FILE-PATH | DIR-PATH\\] \\[flags\\]\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when commit does not exist",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit 73d7fee2f31ade8e1a9c456c324255212c3123ab` + suite.defaultKosliArguments,
			goldenRegex: "Error: map\\[error:map\\[message:Resource not found\\] type:error\\]\n",
		},
		{
			name: "report Bitbucket PR evidence works when --assert is used and commit has a PR",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
					  --assert
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nbitbucket pull request evidence is reported to artifact: .*\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --assert is used and commit has no PRs",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
					  --assert
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit 3dce097040987c4693d2e4be817474d9d0063c93` + suite.defaultKosliArguments,
			goldenRegex: "found 0 pull request\\(s\\) for commit: .*\nbitbucket pull request evidence is reported to artifact: .*\nError: assert failed: no pull request found for the given commit: .*\n",
		},
		{
			name: "report Bitbucket PR evidence does not fail when commit has no PRs",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit 3dce097040987c4693d2e4be817474d9d0063c93` + suite.defaultKosliArguments,
			goldenRegex: "found 0 pull request\\(s\\) for commit: .*\nbitbucket pull request evidence is reported to artifact: .*\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when the artifact does not exist in the server",
			cmd: `report evidence artifact pullrequest bitbucket testdata/file1 --artifact-type file --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "found 1 pull request\\(s\\) for commit: .*\nError: Artifact with fingerprint '7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'bitbucket-pr' belonging to organization 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --artifact-type is unsupported",
			cmd: `report evidence artifact pullrequest bitbucket testdata/file1 --artifact-type unsupported --name bb-pr --flow ` + suite.flowName + `
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: unsupported is not a supported artifact type\n",
		},
		{
			wantError: true,
			name:      "report Bitbucket PR evidence fails when --user-data is not found",
			cmd: `report evidence artifact pullrequest bitbucket --fingerprint ` + suite.artifactFingerprint + ` --name bb-pr --flow ` + suite.flowName + `
					  --user-data non-existing.json
			          --build-url http://www.example.com --bitbucket-workspace kosli-dev --repository cli-test --commit fd54040fc90e7e83f7b152619bfa18917b72c34f` + suite.defaultKosliArguments,
			goldenRegex: "Error: open non-existing.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidencePRBitbucketCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidencePRBitbucketCommandTestSuite))
}
