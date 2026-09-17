package main

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/suite"
)

// AttestationCommitInfoTestSuite guards how a failed commit lookup is reported.
// A --commit defaulted from the CI environment must not fail a command that
// asked for no repository (kosli-dev/server#6094, kosli-dev/server#5615), while
// anything asked for explicitly, or needed by the command, still fails.
//
// The CI default exists only while KOSLI_TESTS is unset, because DefaultValue
// returns "" under it. inCI unsets it around a command run, as TestDefaultValue
// does, and simulates a GitHub Actions job whose GITHUB_SHA is the given commit.
type AttestationCommitInfoTestSuite struct {
	suite.Suite
	headHash              string
	defaultKosliArguments string
}

const (
	commitInfoTestFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	// A well-formed SHA that is not in this repository, as in a shallow clone.
	absentSHA = "0d4c1e1b7f5c2a9e8b3d6f0a1c4e7b2d5a8f3c60"
)

func (suite *AttestationCommitInfoTestSuite) SetupTest() {
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

// inCI runs f with the CI defaults live, as in a GitHub Actions job whose
// GITHUB_SHA is sha.
func (suite *AttestationCommitInfoTestSuite) inCI(sha string, f func()) {
	if value, set := os.LookupEnv("KOSLI_TESTS"); set {
		suite.Require().NoError(os.Unsetenv("KOSLI_TESTS"))
		defer func() { suite.Require().NoError(os.Setenv("KOSLI_TESTS", value)) }()
	}
	suite.T().Setenv("GITHUB_RUN_NUMBER", "1")
	suite.T().Setenv("GITHUB_SHA", sha)
	f()
}

func (suite *AttestationCommitInfoTestSuite) attestGeneric(extraFlags string) string {
	return "attest generic --fingerprint " + commitInfoTestFingerprint + " --name foo --flow f --trail t " + extraFlags + suite.defaultKosliArguments
}

func (suite *AttestationCommitInfoTestSuite) beginTrail(extraFlags string) string {
	return "begin trail t --flow f " + extraFlags + suite.defaultKosliArguments
}

func (suite *AttestationCommitInfoTestSuite) TestCIDefaultedCommitWithoutRepositoryWarns() {
	// --repo-root defaults to ".", so run where there is no repository at all.
	suite.T().Chdir(suite.T().TempDir())
	suite.inCI(suite.headHash, func() {
		for _, cmd := range []string{suite.attestGeneric(""), suite.beginTrail("")} {
			_, out, _, _, err := executeCommandC(cmd)
			suite.Require().NoError(err, cmd)
			suite.Contains(out, "[warning] proceeding without commit info", cmd)
			suite.Contains(out, "--commit "+suite.headHash+" (defaulted from the CI environment)", cmd)
			suite.Contains(out, "repository does not exist", cmd)
			suite.Contains(out, "THIS IS A DRY-RUN", cmd)
			suite.NotContains(out, "git_commit_info", cmd)
		}
	})
}

func (suite *AttestationCommitInfoTestSuite) TestCIDefaultedCommitNotInRepositoryWarns() {
	suite.T().Chdir("../..")
	suite.inCI(absentSHA, func() {
		_, out, _, _, err := executeCommandC(suite.attestGeneric(""))
		suite.Require().NoError(err)
		suite.Contains(out, "[warning] proceeding without commit info")
		suite.Contains(out, "--commit "+absentSHA+" (defaulted from the CI environment)")
		suite.NotContains(out, "git_commit_info")
	})
}

func (suite *AttestationCommitInfoTestSuite) TestCIDefaultedCommitWithExplicitRepoRootFails() {
	suite.inCI(suite.headHash, func() {
		for _, cmd := range []string{suite.attestGeneric("--repo-root testdata"), suite.beginTrail("--repo-root testdata")} {
			_, out, _, _, err := executeCommandC(cmd)
			suite.Require().Error(err, cmd)
			suite.Contains(err.Error(), "failed to get commit info for --commit "+suite.headHash+" (defaulted from the CI environment)", cmd)
			suite.Contains(err.Error(), "repository does not exist", cmd)
			suite.Contains(err.Error(), "Point --repo-root at a repository containing it", cmd)
			suite.NotContains(out, "[warning] proceeding without commit info", cmd)
		}
	})
}

func (suite *AttestationCommitInfoTestSuite) TestExplicitCommitWithoutRepositoryFails() {
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "attest generic: an explicit --commit fails when --repo-root has no repository",
			cmd:         suite.attestGeneric("--commit " + suite.headHash + " --repo-root testdata"),
			goldenRegex: "Error: failed to get commit info for --commit " + suite.headHash + ": .*repository does not exist\\. Point --repo-root at a repository containing it\n",
		},
		{
			wantError:   true,
			name:        "begin trail: an explicit --commit fails when --repo-root has no repository",
			cmd:         suite.beginTrail("--commit " + suite.headHash + " --repo-root testdata"),
			goldenRegex: "Error: failed to get commit info for --commit " + suite.headHash + ": .*repository does not exist\\. Point --repo-root at a repository containing it\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

func (suite *AttestationCommitInfoTestSuite) TestExplicitCommitIsAttached() {
	tests := []cmdTestCase{
		{
			name:        "attest generic: an explicit --commit is resolved and sent",
			cmd:         suite.attestGeneric("--commit " + suite.headHash + " --repo-root ../.."),
			goldenRegex: `(?s)"git_commit_info": \{.*"sha1": "` + suite.headHash + `"`,
		},
		{
			name:        "begin trail: an explicit --commit is resolved and sent",
			cmd:         suite.beginTrail("--commit " + suite.headHash + " --repo-root ../.."),
			goldenRegex: `(?s)"git_commit_info": \{.*"sha1": "` + suite.headHash + `"`,
		},
	}
	runTestCmd(suite.T(), tests)
}

// The commands that do their work from the commit get one error naming what
// needs it, not a warning followed by a nil dereference.
func (suite *AttestationCommitInfoTestSuite) TestCommandsNeedingTheCommitFail() {
	suite.T().Chdir(suite.T().TempDir())
	suite.inCI(suite.headHash, func() {
		for cmd, need := range map[string]string{
			"attest pullrequest github --name foo --flow f --trail t --github-token tok --github-org o --repository r" + suite.defaultKosliArguments:                 "find pull requests",
			"attest jira --name foo --flow f --trail t --jira-base-url https://x.atlassian.net --jira-username u --jira-api-token tok" + suite.defaultKosliArguments: "search for Jira issue keys",
		} {
			_, out, _, _, err := executeCommandC(cmd)
			suite.Require().Error(err, cmd)
			suite.Contains(err.Error(), "failed to get commit info for --commit "+suite.headHash+" (defaulted from the CI environment)", cmd)
			suite.Contains(err.Error(), "The commit is required to "+need, cmd)
			suite.NotContains(out, "[warning] proceeding without commit info", cmd)
		}
	})
}

// RequireFlags keeps --commit non-empty for these commands, so this guards the
// nil dereference that would follow if that ever changed.
func (suite *AttestationCommitInfoTestSuite) TestCommandsNeedingTheCommitFailWithoutOne() {
	o := &CommonAttestationOptions{
		fingerprintOptions:      &fingerprintOptions{},
		attestationNameTemplate: "foo",
		commitRequiredFor:       "find pull requests",
	}
	err := o.run([]string{}, &CommonAttestationPayload{})
	suite.Require().Error(err)
	suite.Contains(err.Error(), "the commit is required to find pull requests")
}

func TestAttestationCommitInfoTestSuite(t *testing.T) {
	suite.Run(t, new(AttestationCommitInfoTestSuite))
}
