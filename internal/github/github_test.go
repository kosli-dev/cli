package github

import (
	"context"
	"os"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GithubTestSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *GithubTestSuite) TestNewGithubClientFromToken() {
	for _, t := range []struct {
		name    string
		token   string
		baseURL string
	}{
		{
			name:  "when provided a token, a client is created.",
			token: "some_fake_token",
		},
		{
			name:    "when baseURL and token are provided, a client is created.",
			token:   "some_fake_token",
			baseURL: "https://github.example.com",
		},
	} {
		suite.Run(t.name, func() {
			client, err := NewGithubClientFromToken(context.Background(), t.token, t.baseURL)
			require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
			require.NotNilf(suite.T(), client, "client should not be nil")
		})
	}
}

func (suite *GithubTestSuite) TestPREvidenceForCommit() {
	type result struct {
		wantError   bool
		numberOfPRs int
	}
	for _, t := range []struct {
		name           string
		config         *GithubConfig
		commit         string
		requireEnvVars bool // indicates that a test case needs real credentials from env vars
		result         result
	}{
		{
			name: "invalid token causes an error",
			config: &GithubConfig{
				Token:      "some_fake_token",
				Org:        "kosli-dev",
				Repository: "cli",
			},
			result: result{
				wantError: true,
			},
		},
		{
			name: "can list pull requests for a commit.",
			config: &GithubConfig{
				Org:        "kosli-dev",
				Repository: "cli",
			},
			requireEnvVars: true,
			result: result{
				numberOfPRs: 1,
			},
		},
		{
			name: "non-existing commit will cause an error.",
			config: &GithubConfig{
				Org:        "kosli-dev",
				Repository: "cli",
			},
			commit:         "73d7fee2f31ade8e1a9c456c324255212c3tf45a",
			requireEnvVars: true,
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			if t.requireEnvVars {
				testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})
				t.config.Token = os.Getenv("KOSLI_GITHUB_TOKEN")
			}
			if t.commit == "" {
				t.commit = testHelpers.GithubCommitWithPR()
			}
			prs, err := t.config.PREvidenceForCommit(t.commit)
			if t.result.wantError {
				require.Errorf(suite.T(), err, "expected an error but got: %s", err)
			} else {
				require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
				require.Len(suite.T(), prs, t.result.numberOfPRs)
			}
		})
	}
}

func (suite *GithubTestSuite) TestPullRequestsForCommit() {
	type result struct {
		wantError   bool
		numberOfPRs int
	}
	for _, t := range []struct {
		name       string
		ghOrg      string
		repository string
		commit     string
		result     result
	}{
		{
			name:       "can list pull requests for a commit.",
			ghOrg:      "kosli-dev",
			repository: "cli",
			result: result{
				wantError:   false,
				numberOfPRs: 1,
			},
		},
		{
			name:       "non-existing commit will cause an error.",
			ghOrg:      "kosli-dev",
			repository: "cli",
			commit:     "73d7fee2f31ade8e1a9c456c324255212c3tf45a",
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})
			token := os.Getenv("KOSLI_GITHUB_TOKEN")
			c := &GithubConfig{
				Token:      token,
				Org:        t.ghOrg,
				Repository: t.repository,
			}
			if t.commit == "" {
				t.commit = testHelpers.GithubCommitWithPR()
			}
			prs, err := c.PullRequestsForCommit(t.commit)
			if t.result.wantError {
				require.Errorf(suite.T(), err, "expected an error but got: %s", err)
			} else {
				require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
				require.Lenf(suite.T(), prs, t.result.numberOfPRs, "expected %d PRs but got %d", t.result.numberOfPRs, len(prs))
			}
		})
	}
}

func (suite *GithubTestSuite) TestGetPullRequestApprovers() {
	type result struct {
		wantError bool
		approvers []string
	}
	for _, t := range []struct {
		name       string
		ghOrg      string
		repository string
		number     int
		result     result
	}{
		{
			name:       "get an empty list for a PR without approvers",
			ghOrg:      "kosli-dev",
			repository: "cli",
			number:     8,
			result: result{
				approvers: []string{},
			},
		},
		{
			name:       "get the list of approvers for an approved PR",
			ghOrg:      "kosli-dev",
			repository: "cli",
			number:     6,
			result: result{
				approvers: []string{"sami-alajrami"},
			},
		},
		{
			name:       "non-existing PR causes an error",
			ghOrg:      "kosli-dev",
			repository: "cli",
			number:     666,
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITHUB_TOKEN"})
			token := os.Getenv("KOSLI_GITHUB_TOKEN")
			c := &GithubConfig{
				Token:      token,
				Org:        t.ghOrg,
				Repository: t.repository,
			}
			approvers, err := c.GetPullRequestApprovers(t.number)
			if t.result.wantError {
				require.Errorf(suite.T(), err, "expected an error but got: %s", err)
			} else {
				require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
				require.ElementsMatchf(suite.T(), t.result.approvers, approvers, "want approvers: %v, got approvers: %v",
					t.result.approvers, approvers)
			}
		})
	}
}

func (suite *GithubTestSuite) TestExtractRepoName() {
	for _, t := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "full repo name (including org) is separated",
			input: "kosli-dev/cli",
			want:  "cli",
		},
		{
			name:  "short repo name is returned as is",
			input: "cli",
			want:  "cli",
		},
	} {
		suite.Run(t.name, func() {
			repo := extractRepoName(t.input)
			require.Equalf(suite.T(), t.want, repo, "expected %s but got %s", t.want, repo)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGithubTestSuite(t *testing.T) {
	suite.Run(t, new(GithubTestSuite))
}
