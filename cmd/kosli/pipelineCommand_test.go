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
	defaultKosliArguments := " -H http://localhost:8001 --owner docs-cmd-test-user -a eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"
	defaultArtifactArguments := " --pipeline newPipe --build-url www.yr.no --commit-url www.nrk.no"
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
		{
			name:   "declare pipeline",
			cmd:    "pipeline declare --pipeline newPipe --description \"my new pipeline\" " + defaultKosliArguments,
			golden: "",
		},
		{
			name:   "re-declaring a pipeline updates its metadata",
			cmd:    "pipeline declare --pipeline newPipe --description \"changed description\" " + defaultKosliArguments,
			golden: "",
		},
		{
			wantError: true,
			name:      "missing --owner flag causes an error",
			cmd:       "pipeline declare --pipeline newPipe --description \"my new pipeline\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --owner is not set\nUsage: kosli pipeline declare [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "pipeline declare --pipeline newPipe --description \"my new pipeline\" --owner cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli pipeline declare [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --pipeline causes an error",
			cmd:       "pipeline declare --description \"my new pipeline\" -H http://localhost:8001 --owner cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --pipeline is required when you are not using --pipefile\nUsage: kosli pipeline declare [flags]\n",
		},
		// Pipeline ls tests
		{
			wantError: false,
			name:      "kosli pipeline ls command does not return error",
			cmd:       "pipeline ls" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli pipeline ls --output json command does not return error",
			cmd:       "pipeline ls --output json" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli pipeline ls --output table command does not return error",
			cmd:       "pipeline ls --output table" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: true,
			name:      "kosli pipeline ls --output text command does return error",
			cmd:       "pipeline ls --output text" + defaultKosliArguments,
			golden:    "",
		},

		// Pipeline pipeline get tests
		{
			wantError: false,
			name:      "kosli pipeline inspect newPipe command does not return error",
			cmd:       "pipeline inspect newPipe" + defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: false,
			name:      "kosli pipeline inspect newPipe --output json command does not return error",
			cmd:       "pipeline inspect newPipe --output json" + defaultKosliArguments,
			golden:    "",
		},

		// Report artifacts
		{
			wantError: false,
			name:      "report artifact 1",
			cmd:       "pipeline artifact report creation FooBar_1 --git-commit " + headHash + " --sha256 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0" + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			// Git commit SHA has to be the same as the previous one so we test that an empty commit list is reported correctly
			wantError: false,
			name:      "report artifact 2",
			cmd:       "pipeline artifact report creation FooBar_2 --git-commit " + headHash + " --sha256 4f09b9f4e4d354a42fd4599d0ef8e04daf278c967dea68741d127f21eaa1eeaf" + defaultArtifactArguments + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},

		// List artifacts
		{
			wantError: false,
			name:      "list artifacts",
			cmd:       "artifact ls newPipe" + defaultKosliArguments,
			golden:    "",
		},

		// Get artifact
		{
			wantError: false,
			name:      "get artifact",
			cmd:       "artifact get newPipe@4f09b9f4e4d354a42fd4599d0ef8e04daf278c967dea68741d127f21eaa1eeaf" + defaultKosliArguments,
			golden:    "",
		},

		// TODO: decouple approval tests and make them independent
		// Report approval
		{
			wantError: false,
			name:      "report approval",
			cmd:       "pipeline approval report --pipeline newPipe --sha256 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0 --oldest-commit HEAD~1" + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},

		// Request approval
		{
			wantError: false,
			name:      "request approval",
			cmd:       "pipeline approval request --pipeline newPipe --sha256 4f09b9f4e4d354a42fd4599d0ef8e04daf278c967dea68741d127f21eaa1eeaf --oldest-commit HEAD~1" + defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},

		// Assert approval
		{
			wantError: false,
			name:      "assert an approved approval does not fail",
			cmd:       "pipeline approval assert --pipeline newPipe --sha256 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0" + defaultKosliArguments,
			golden:    "",
		},

		{
			wantError: true,
			name:      "assert a pending approval fails",
			cmd:       "pipeline approval assert --pipeline newPipe --sha256 4f09b9f4e4d354a42fd4599d0ef8e04daf278c967dea68741d127f21eaa1eeaf" + defaultKosliArguments,
			golden:    "",
		},

		// list approvals
		{
			wantError: false,
			name:      "list approvals",
			cmd:       "approval ls newPipe" + defaultKosliArguments,
			golden:    "",
		},

		// Get an approval
		{
			wantError: false,
			name:      "get an approval",
			cmd:       "approval get newPipe#2" + defaultKosliArguments,
			golden:    "",
		},

		{
			wantError: true,
			name:      "get a non-existing approval fails",
			cmd:       "approval get newPipe#20" + defaultKosliArguments,
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
