package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// RepoInfoWiringTestSuite guards that the CI-detected repo name flows into
// repo_info.name and that an explicit --repository overrides it. It covers the
// three distinct run() paths that build repo_info: the shared
// CommonAttestationOptions.run() (via `attest generic`), `begin trail`, and
// `attest artifact`.
//
// It does NOT cover the `attest pr *` commands: those call the live GitHub /
// GitLab API to gather PR evidence before the dry-run guard, so they can't run
// here without provider tokens (see attestPRGitlab_test.go, which is token-gated).
//
// The original production bug (a short CI *default* clobbering the fuller
// CI-detected name) can't be reproduced under KOSLI_TESTS, which zeroes flag
// defaults; that logic is covered directly by TestMergeGitRepoInfo. This suite
// instead guards the command wiring: an explicit --repository must reach
// repo_info.name. Because the CI base name here is non-empty, a missing
// `repoNameExplicit` assignment would leave the base name in place and fail the
// explicit-override cases.
type RepoInfoWiringTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	origEnv               map[string]*string
}

func (suite *RepoInfoWiringTestSuite) SetupSuite() {
	// Simulate GitHub Actions so getGitRepoInfoFromEnvironment() returns a base
	// with the full-path name "kosli-dev/cli". Save originals for restoration.
	ciEnv := map[string]string{
		"GITHUB_RUN_NUMBER":    "1",
		"GITHUB_SERVER_URL":    "https://github.com",
		"GITHUB_REPOSITORY":    "kosli-dev/cli",
		"GITHUB_REPOSITORY_ID": "123456",
	}
	suite.origEnv = map[string]*string{}
	for k, v := range ciEnv {
		if orig, ok := os.LookupEnv(k); ok {
			o := orig
			suite.origEnv[k] = &o
		} else {
			suite.origEnv[k] = nil
		}
		suite.Require().NoError(os.Setenv(k, v))
	}
}

func (suite *RepoInfoWiringTestSuite) TearDownSuite() {
	for k, v := range suite.origEnv {
		if v == nil {
			suite.Require().NoError(os.Unsetenv(k))
		} else {
			suite.Require().NoError(os.Setenv(k, *v))
		}
	}
}

func (suite *RepoInfoWiringTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "DRY_RUN",
		Org:      "test-org",
		Host:     "http://localhost:8001",
		DryRun:   true,
	}
	// --repo-root ../.. points git resolution at the real repo root.
	suite.defaultKosliArguments = " --repo-root ../.. --dry-run --host http://localhost:8001 --org test-org --api-token DRY_RUN"
}

func (suite *RepoInfoWiringTestSuite) TestRepoInfoNameWiring() {
	const fingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	tests := []cmdTestCase{
		{
			name:        "attest generic: CI-detected full-path name is used by default",
			cmd:         "attest generic --fingerprint " + fingerprint + " --name foo --flow f --trail t" + suite.defaultKosliArguments,
			goldenRegex: `"name": "kosli-dev/cli"`,
		},
		{
			name:        "attest generic: explicit --repository overrides the CI-detected name",
			cmd:         "attest generic --fingerprint " + fingerprint + " --name foo --flow f --trail t --repository my/explicit-repo" + suite.defaultKosliArguments,
			goldenRegex: `"name": "my/explicit-repo"`,
		},
		{
			name:        "begin trail: CI-detected full-path name is used by default",
			cmd:         "begin trail t --flow f" + suite.defaultKosliArguments,
			goldenRegex: `"name": "kosli-dev/cli"`,
		},
		{
			name:        "begin trail: explicit --repository overrides the CI-detected name",
			cmd:         "begin trail t --flow f --repository my/explicit-repo" + suite.defaultKosliArguments,
			goldenRegex: `"name": "my/explicit-repo"`,
		},
		{
			name:        "attest artifact: CI-detected full-path name is used by default",
			cmd:         "attest artifact foo --fingerprint " + fingerprint + " --name n --flow f --trail t --commit HEAD --build-url b --commit-url c" + suite.defaultKosliArguments,
			goldenRegex: `"name": "kosli-dev/cli"`,
		},
		{
			name:        "attest artifact: explicit --repository overrides the CI-detected name",
			cmd:         "attest artifact foo --fingerprint " + fingerprint + " --name n --flow f --trail t --commit HEAD --build-url b --commit-url c --repository my/explicit-repo" + suite.defaultKosliArguments,
			goldenRegex: `"name": "my/explicit-repo"`,
		},
	}
	runTestCmd(suite.T(), tests)
}

func TestRepoInfoWiringTestSuite(t *testing.T) {
	suite.Run(t, new(RepoInfoWiringTestSuite))
}
