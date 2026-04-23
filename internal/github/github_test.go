package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// graphqlNullPRResponse is a minimal valid GraphQL response where the PR is null.
// PREvidenceByPRNumber returns nil, nil in this case.
const graphqlNullPRResponse = `{"data":{"repository":{"pullRequest":null}}}`

func newRetryTestServer(t *testing.T, failCount int) (*httptest.Server, *int) {
	t.Helper()
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graphql" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		attempts++
		if attempts <= failCount {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, graphqlNullPRResponse)
	}))
	return ts, &attempts
}

func newRetryConfig(serverURL string, sleepFn func(time.Duration)) *GithubConfig {
	return &GithubConfig{
		Token:      "fake-token",
		BaseURL:    serverURL,
		Org:        "test-org",
		Repository: "test-repo",
		Sleep:      sleepFn,
	}
}

func TestPREvidenceByPRNumber_RetriesOnGraphQLError(t *testing.T) {
	ts, attempts := newRetryTestServer(t, 2)
	defer ts.Close()

	var sleptDurations []time.Duration
	config := newRetryConfig(ts.URL, func(d time.Duration) { sleptDurations = append(sleptDurations, d) })

	pr, err := config.PREvidenceByPRNumber(1)
	require.NoError(t, err)
	require.Nil(t, pr)
	require.Equal(t, 3, *attempts, "should have retried twice before succeeding")
	require.Len(t, sleptDurations, 2, "should have slept between retries")
}

func TestPREvidenceByPRNumber_ReturnsErrorAfterAllRetriesExhausted(t *testing.T) {
	ts, attempts := newRetryTestServer(t, 999)
	defer ts.Close()

	config := newRetryConfig(ts.URL, func(time.Duration) {})

	_, err := config.PREvidenceByPRNumber(1)
	require.Error(t, err)
	require.Equal(t, 4, *attempts, "should have made 1 initial attempt + 3 retries")
}
