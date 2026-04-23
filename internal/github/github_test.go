package github

import (
	"context"
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
