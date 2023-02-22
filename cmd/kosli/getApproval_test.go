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
type GetApprovalCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments    string
	defaultArtifactArguments string
	flowName                 string
}

func (suite *GetApprovalCommandTestSuite) SetupTest() {
	suite.flowName = "approval-42"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	suite.defaultArtifactArguments = " --flow " + suite.flowName + " --build-url www.yr.no --commit-url www.nrk.no"

	CreateFlow(suite.flowName, suite.T())
}

func (suite *GetApprovalCommandTestSuite) TestGetDeploymentCmd() {
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
			wantError: false,
			name:      "report artifact with fingerprint",
			cmd:       "report artifact FooBar_1 --git-commit " + headHash + " --fingerprint 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0" + suite.defaultArtifactArguments + suite.defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: false,
			name:      "report approval",
			cmd:       "report approval --flow " + suite.flowName + " --oldest-commit HEAD~1 --fingerprint 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0" + suite.defaultKosliArguments + defaultRepoRoot,
			golden:    "",
		},
		{
			wantError: false,
			name:      "get an approval",
			cmd:       "get approval " + suite.flowName + " " + suite.defaultKosliArguments,
			golden:    "",
		},
		{
			wantError: true,
			name:      "get a non-existing approval fails",
			cmd:       "get approval newFlow#20" + suite.defaultKosliArguments,
			golden:    "",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetApprovalCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetApprovalCommandTestSuite))
}
