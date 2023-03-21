package main

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PipelineCommandTestSuite struct {
	suite.Suite
}

func (suite *PipelineCommandTestSuite) TestPipelineCommandCmd() {
	defaultKosliArguments := " -H http://localhost:8001 --org docs-cmd-test-user -a eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"
	defaultArtifactArguments := " --flow newFlow --build-url www.yr.no --commit-url www.nrk.no"
	defaultRepoRoot := " --repo-root ../.. "

	repo, err := git.PlainOpen("../..")
	if err != nil {
		suite.T().Fatal(fmt.Errorf("failed to open git repository at %s: %v", "../..", err))
	}
	// headHash, err := repo.ResolveRevision(plumbing.Revision("HEAD"))
	repoHead, err := repo.Head()
	if err != nil {
		suite.T().Fatal(fmt.Errorf("failed to resolve revision %s: %v", "HEAD", err))
	}
	headHash := repoHead.Hash().String()

	tests := []cmdTestCase{

		// Report artifacts
		{
			wantError: false,
			name:      "report artifact with fingerprint",
			cmd:       "report artifact FooBar_1 --git-commit " + headHash + " --fingerprint 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0" + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: false,
			name:      "report different artifact with same git commit",
			cmd:       "report artifact FooBar_2 --git-commit " + headHash + " --fingerprint 4f09b9f4e4d354a42fd4599d0ef8e04daf278c967dea68741d127f21eaa1eeaf" + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: false,
			name:      "report artifact file",
			cmd:       "report artifact testdata/file1 --artifact-type file --git-commit " + headHash + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: false,
			name:      "report artifact dir",
			cmd:       "report artifact testdata/folder1 --artifact-type dir --git-commit " + headHash + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: true,
			name:      "report artifact missing --org",
			cmd:       "report artifact testdata/folder1 --artifact-type dir --git-commit " + headHash + defaultArtifactArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: true,
			name:      "report artifact missing --artifact-type",
			cmd:       "report artifact testdata/folder1 --git-commit " + headHash + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: true,
			name:      "report artifact missing --git-commit",
			cmd:       "report artifact testdata/folder1 --artifact-type dir " + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "Error: required flag(s) \"git-commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report artifact file with non existing file name",
			cmd:       "report artifact thisIsNotAFile --artifact-type file --git-commit " + headHash + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "Error: open thisIsNotAFile: no such file or directory\n",
		},
		{
			wantError: true,
			name:      "report artifact wrong --repo-root",
			cmd:       "report artifact testdata/file1 --repo-root . --artifact-type file --git-commit " + headHash + defaultArtifactArguments + defaultKosliArguments,
			golden:    "Error: failed to open git repository at .: repository does not exist\n",
		},

		// TODO: decouple approval tests and make them independent
		// Report approval
		{
			wantError: false,
			name:      "report approval",
			cmd:       "report approval --flow newFlow --oldest-commit HEAD~1 --fingerprint 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0" + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},

		// Request approval
		{
			wantError: false,
			name:      "request approval",
			cmd:       "request approval --flow newFlow --oldest-commit HEAD --fingerprint 4f09b9f4e4d354a42fd4599d0ef8e04daf278c967dea68741d127f21eaa1eeaf" + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPipelineCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PipelineCommandTestSuite))
}
