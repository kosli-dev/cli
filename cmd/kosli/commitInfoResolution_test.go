package main

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/suite"
)

// CommitInfoResolutionTestSuite guards that a --commit which was defaulted from
// the CI environment does not fail the command when git cannot supply its info,
// while an explicitly passed --commit still does.
//
// The production trigger (a CI-defaulted --commit in a job with no checked-out
// repository) cannot be reproduced through the command harness, because
// DefaultValue returns "" whenever KOSLI_TESTS is set. resolveCommitInfo is
// therefore exercised directly, and the command cases below guard only that
// each command assigns commitSHAExplicit.
type CommitInfoResolutionTestSuite struct {
	suite.Suite
	headHash              string
	defaultKosliArguments string
}

func (suite *CommitInfoResolutionTestSuite) SetupTest() {
	repo, err := git.PlainOpen("../..")
	suite.Require().NoError(err)
	head, err := repo.Head()
	suite.Require().NoError(err)
	suite.headHash = head.Hash().String()

	global = &GlobalOpts{
		ApiToken: "DRY_RUN",
		Org:      "test-org",
		Host:     "http://localhost:8001",
		DryRun:   true,
	}
	suite.defaultKosliArguments = " --dry-run --host http://localhost:8001 --org test-org --api-token DRY_RUN"
}

func (suite *CommitInfoResolutionTestSuite) TestResolveCommitInfoWithoutRepository() {
	const noRepo = "testdata"

	info, err := resolveCommitInfo(noRepo, suite.headHash, false, []string{})
	suite.Require().NoError(err, "a CI-defaulted commit must not fail when there is no repository")
	suite.Nil(info)

	_, err = resolveCommitInfo(noRepo, suite.headHash, true, []string{})
	suite.Require().Error(err, "an explicit --commit must still fail when there is no repository")
	suite.Contains(err.Error(), "repository does not exist")
}

func (suite *CommitInfoResolutionTestSuite) TestResolveCommitInfoWithUnresolvableCommit() {
	// A well-formed SHA that is not in this repository, as in a shallow clone.
	const absentSHA = "0d4c1e1b7f5c2a9e8b3d6f0a1c4e7b2d5a8f3c60"

	info, err := resolveCommitInfo("../..", absentSHA, false, []string{})
	suite.Require().NoError(err, "a CI-defaulted commit must not fail when it cannot be resolved")
	suite.Nil(info)

	_, err = resolveCommitInfo("../..", absentSHA, true, []string{})
	suite.Require().Error(err, "an explicit --commit must still fail when it cannot be resolved")
}

func (suite *CommitInfoResolutionTestSuite) TestResolveCommitInfoSucceeds() {
	info, err := resolveCommitInfo("../..", suite.headHash, false, []string{})
	suite.Require().NoError(err)
	suite.Require().NotNil(info)
	suite.Equal(suite.headHash, info.Sha1)
}

func (suite *CommitInfoResolutionTestSuite) TestExplicitCommitWiring() {
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "attest generic: an explicit --commit fails when --repo-root has no repository",
			cmd:         fmt.Sprintf("attest generic --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --flow f --trail t --commit %s --repo-root testdata%s", suite.headHash, suite.defaultKosliArguments),
			goldenRegex: "Error: failed to get commit info\\. .*repository does not exist\n",
		},
		{
			wantError:   true,
			name:        "begin trail: an explicit --commit fails when --repo-root has no repository",
			cmd:         fmt.Sprintf("begin trail t --flow f --commit %s --repo-root testdata%s", suite.headHash, suite.defaultKosliArguments),
			goldenRegex: "Error: failed to get commit info\\. .*repository does not exist\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// commitRequiredOptions builds the shared attestation options for a command run
// whose --commit came from the CI default and cannot be resolved, which is the
// only way payload.Commit reaches these commands as nil.
func (suite *CommitInfoResolutionTestSuite) commitRequiredOptions() *CommonAttestationOptions {
	return &CommonAttestationOptions{
		fingerprintOptions:      &fingerprintOptions{},
		attestationNameTemplate: "foo",
		flowName:                "f",
		trailName:               "t",
		commitSHA:               suite.headHash,
		srcRepoRoot:             "testdata",
		commitSHAExplicit:       false,
	}
}

func (suite *CommitInfoResolutionTestSuite) TestCommandsNeedingCommitReportIt() {
	pr := &attestPROptions{
		CommonAttestationOptions: suite.commitRequiredOptions(),
		payload:                  PRAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{}},
	}
	err := pr.run([]string{})
	suite.Require().Error(err)
	suite.Contains(err.Error(), "required to find pull requests")

	jira := &attestJiraOptions{
		CommonAttestationOptions: suite.commitRequiredOptions(),
		payload:                  JiraAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{}},
	}
	err = jira.run([]string{})
	suite.Require().Error(err)
	suite.Contains(err.Error(), "required to search for Jira issue keys")
}

func TestCommitInfoResolutionTestSuite(t *testing.T) {
	suite.Run(t, new(CommitInfoResolutionTestSuite))
}
