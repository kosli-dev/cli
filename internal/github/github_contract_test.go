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

	t.Run("Hybrid returns PRs for commit with PRs", func(t *testing.T) {
		prs, err := provider.PREvidenceForCommitHybrid(commitWithPR)
		require.NoError(t, err)
		require.NotEmpty(t, prs)
		require.NotEmpty(t, prs[0].URL, "URL should be present")
		require.NotEmpty(t, prs[0].State, "State should be present")
		require.NotEmpty(t, prs[0].MergeCommit, "MergeCommit should be present")
	})

	t.Run("ProviderAndLabel returns github and pull request", func(t *testing.T) {
		providerName, label := provider.ProviderAndLabel()
		require.Equal(t, "github", providerName)
		require.Equal(t, "pull request", label)
	})
}

// prByNumberRetriever is a local interface used to test PREvidenceByPRNumber
// independently of the PRRetriever contract.
type prByNumberRetriever interface {
	PREvidenceByPRNumber(int) (*types.PREvidence, error)
}

func runPRByNumberContractTests(t *testing.T, provider prByNumberRetriever, knownPRNumber int) {
	t.Helper()

	t.Run("returns evidence for known PR number", func(t *testing.T) {
		pr, err := provider.PREvidenceByPRNumber(knownPRNumber)
		require.NoError(t, err)
		require.NotNil(t, pr)
		require.NotEmpty(t, pr.URL, "URL should be present")
		require.NotEmpty(t, pr.State, "State should be present")
		require.NotEmpty(t, pr.MergeCommit, "MergeCommit should be present")
	})

	t.Run("returns error for unknown PR number", func(t *testing.T) {
		pr, err := provider.PREvidenceByPRNumber(999999999)
		require.Error(t, err)
		require.Nil(t, pr)
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

	// Hybrid fallback path: V2 returns empty (PRsByCommit not seeded for this
	// commit), so the fake falls back through PRNumbersByCommit + PRsByNumber.
	commitFallback := "fallback-commit"
	prNumber := testHelpers.GithubPRNumber()
	fallbackPR := &types.PREvidence{
		URL:         "https://github.com/kosli-dev/cli/pull/6",
		State:       "MERGED",
		MergeCommit: commitFallback,
	}
	fallbackClient := &FakeGitHubClient{
		PRNumbersByCommit: map[string][]int{
			commitFallback: {prNumber},
		},
		PRsByNumber: map[int]*types.PREvidence{
			prNumber: fallbackPR,
		},
	}

	t.Run("Hybrid returns PRs via fallback path when V2 returns empty", func(t *testing.T) {
		prs, err := fallbackClient.PREvidenceForCommitHybrid(commitFallback)
		require.NoError(t, err)
		require.NotEmpty(t, prs)
		require.Equal(t, fallbackPR.URL, prs[0].URL)
	})

	t.Run("Hybrid returns empty when commit has no PRs in either path", func(t *testing.T) {
		prs, err := fallbackClient.PREvidenceForCommitHybrid("commit-with-no-prs")
		require.NoError(t, err)
		require.Empty(t, prs)
	})

	t.Run("Hybrid returns error when Err is injected", func(t *testing.T) {
		client.Err = errInjected
		defer func() { client.Err = nil }()
		_, err := client.PREvidenceForCommitHybrid(commitWithPR)
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

func TestGitHubContract_Fake_PRByNumber(t *testing.T) {
	knownPRNumber := testHelpers.GithubPRNumber()
	pr := &types.PREvidence{
		URL:         "https://github.com/kosli-dev/cli/pull/6",
		State:       "MERGED",
		MergeCommit: "e21a8afff429e0c87ee523d683f2438113f0a105",
	}
	client := &FakeGitHubClient{
		PRsByNumber: map[int]*types.PREvidence{
			knownPRNumber: pr,
		},
	}
	runPRByNumberContractTests(t, client, knownPRNumber)

	t.Run("returns error when Err is injected", func(t *testing.T) {
		client.Err = errInjected
		defer func() { client.Err = nil }()
		_, err := client.PREvidenceByPRNumber(knownPRNumber)
		require.Error(t, err)
	})
}

func TestGitHubContract_RealGitHub_PRByNumber(t *testing.T) {
	testHelpers.SkipIfEnvVarUnset(t, []string{"KOSLI_GITHUB_TOKEN"})

	config := NewGithubConfig(
		os.Getenv("KOSLI_GITHUB_TOKEN"),
		"",
		"kosli-dev",
		"cli",
	)

	runPRByNumberContractTests(t, config, testHelpers.GithubPRNumber())
}
