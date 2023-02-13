package github

import (
	"context"
	"os"
	"testing"

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
		name  string
		token string
	}{
		{
			name:  "when provided a token, a client is created.",
			token: "some_fake_token",
		},
	} {
		suite.Run(t.name, func() {
			client, err := NewGithubClientFromToken(context.Background(), t.token, "")
			require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
			require.NotNilf(suite.T(), client, "client should not be nil")
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
		ghOwner    string
		repository string
		commit     string
		result     result
	}{
		{
			name:       "can list pull requests for a commit.",
			ghOwner:    "kosli-dev",
			repository: "cli",
			commit:     "73d7fee2f31ade8e1a9c456c324255212c30c2a6",
			result: result{
				wantError:   false,
				numberOfPRs: 1,
			},
		},
		{
			name:       "non-existing commit will cause an error.",
			ghOwner:    "kosli-dev",
			repository: "cli",
			commit:     "73d7fee2f31ade8e1a9c456c324255212c3tf45a",
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			token, ok := os.LookupEnv("KOSLI_GITHUB_TOKEN")
			if !ok {
				suite.T().Logf("skipping %s as KOSLI_GITHUB_TOKEN is unset in environment", suite.T().Name())
				suite.T().Skip("requires github token")
			}
			prs, err := PullRequestsForCommit(token, t.ghOwner, t.repository, t.commit, "")
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
		ghOwner    string
		repository string
		number     int
		result     result
	}{
		{
			name:       "get an empty list for a PR without approvers",
			ghOwner:    "kosli-dev",
			repository: "cli",
			number:     8,
			result: result{
				approvers: []string{},
			},
		},
		{
			name:       "get the list of approvers for an approved PR",
			ghOwner:    "kosli-dev",
			repository: "cli",
			number:     6,
			result: result{
				approvers: []string{"sami-alajrami"},
			},
		},
		{
			name:       "non-existing PR causes an error",
			ghOwner:    "kosli-dev",
			repository: "cli",
			number:     666,
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			token, ok := os.LookupEnv("KOSLI_GITHUB_TOKEN")
			if !ok {
				suite.T().Logf("skipping %s as KOSLI_GITHUB_TOKEN is unset in environment", suite.T().Name())
				suite.T().Skip("requires github token")
			}
			approvers, err := GetPullRequestApprovers(token, t.ghOwner, t.repository, t.number, "")
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGithubTestSuite(t *testing.T) {
	suite.Run(t, new(GithubTestSuite))
}
