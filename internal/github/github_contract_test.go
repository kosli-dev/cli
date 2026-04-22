package github

import (
	"os"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// runGitHubContractTests exercises the types.PRRetriever contract against any
// implementation. commitWithPR must be a commit SHA that has at least one
// associated pull request. commitUnknown must be a validly-formatted SHA that
// does not exist in the repository.
//
// V1 and V2 have different contracts for unknown commits:
//   - V2 (GraphQL) returns empty with no error — the GraphQL API returns null
//     for objects that don't exist.
//   - V1 (REST) returns an error — the REST API returns 422 for unknown commits.
//
// Any implementation that passes this suite is a valid stand-in for the real
// GitHub API as far as this codebase is concerned.
func runGitHubContractTests(t *testing.T, provider types.PRRetriever, commitWithPR, commitUnknown string) {
	t.Helper()

	t.Run("V2 returns PRs for commit with PRs", func(t *testing.T) {
		prs, err := provider.PREvidenceForCommitV2(commitWithPR)
		require.NoError(t, err)
		require.NotEmpty(t, prs)
		require.NotEmpty(t, prs[0].URL, "URL should be present")
		require.NotEmpty(t, prs[0].State, "State should be present")
		require.Equal(t, commitWithPR, prs[0].MergeCommit, "V2 sets MergeCommit to the queried commit SHA")
	})

	t.Run("V2 returns empty with no error for unknown commit", func(t *testing.T) {
		prs, err := provider.PREvidenceForCommitV2(commitUnknown)
		require.NoError(t, err)
		require.Empty(t, prs)
	})

	t.Run("V1 returns PRs for commit with PRs", func(t *testing.T) {
		prs, err := provider.PREvidenceForCommitV1(commitWithPR)
		require.NoError(t, err)
		require.NotEmpty(t, prs)
		require.NotEmpty(t, prs[0].URL, "URL should be present")
		require.NotEmpty(t, prs[0].State, "State should be present")
		require.NotEmpty(t, prs[0].MergeCommit, "MergeCommit should be present")
	})

	t.Run("V1 returns error for unknown commit", func(t *testing.T) {
		_, err := provider.PREvidenceForCommitV1(commitUnknown)
		require.Error(t, err)
	})

	t.Run("ProviderAndLabel returns github and pull request", func(t *testing.T) {
		providerName, label := provider.ProviderAndLabel()
		require.Equal(t, "github", providerName)
		require.Equal(t, "pull request", label)
	})
}

func TestGitHubContract_Fake(t *testing.T) {
	commitWithPR := "abc123"
	commitUnknown := "0000000000000000000000000000000000000000"

	pr := &types.PREvidence{
		URL:         "https://github.com/kosli-dev/cli/pull/1",
		State:       "MERGED",
		MergeCommit: commitWithPR,
	}

	client := &FakeGitHubClient{
		PRsByCommit: map[string][]*types.PREvidence{
			commitWithPR: {pr},
		},
	}

	runGitHubContractTests(t, client, commitWithPR, commitUnknown)

	// Error injection is a fake-specific mechanism with no real-API equivalent.
	// These tests verify the fake itself, not the contract.
	t.Run("V2 returns error when Err is injected", func(t *testing.T) {
		client.Err = errInjected
		defer func() { client.Err = nil }()
		_, err := client.PREvidenceForCommitV2(commitWithPR)
		require.Error(t, err)
	})

	t.Run("V1 returns error when Err is injected", func(t *testing.T) {
		client.Err = errInjected
		defer func() { client.Err = nil }()
		_, err := client.PREvidenceForCommitV1(commitWithPR)
		require.Error(t, err)
	})
}

func TestGitHubContract_RealGitHub(t *testing.T) {
	testHelpers.SkipIfEnvVarUnset(t, []string{"KOSLI_GITHUB_TOKEN"})

	config := NewGithubConfig(
		os.Getenv("KOSLI_GITHUB_TOKEN"),
		"",
		"kosli-dev",
		"cli",
	)

	// commitUnknown is a validly-formatted SHA that does not exist in kosli-dev/cli.
	commitUnknown := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	runGitHubContractTests(t, config, testHelpers.GithubCommitWithPR(), commitUnknown)
}
