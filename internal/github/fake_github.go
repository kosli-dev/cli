package github

import (
	"errors"
	"fmt"

	"github.com/kosli-dev/cli/internal/types"
)

// errInjected is returned by FakeGitHubClient when Err is set.
var errInjected = errors.New("injected error")

// FakeGitHubClient is an in-memory implementation of types.PRRetriever for
// testing. Seed PRsByCommit with the commits and PR evidence you want returned.
// Seed PRsByNumber with PR numbers and their evidence for PREvidenceByPRNumber.
// Seed PRNumbersByCommit + PRsByNumber to exercise the hybrid fallback path
// (V2 empty → V1 discovery → per-PR GraphQL).
// Set Err to simulate a network or API failure.
type FakeGitHubClient struct {
	// PRsByCommit maps a commit SHA to the PR evidence returned for that commit.
	PRsByCommit map[string][]*types.PREvidence
	// PRsByNumber maps a PR number to the PR evidence returned for that number.
	PRsByNumber map[int]*types.PREvidence
	// PRNumbersByCommit maps a commit SHA to PR numbers for the hybrid fallback path.
	PRNumbersByCommit map[string][]int
	// Err, if set, is returned by all calls regardless of commit.
	Err error
}

func (f *FakeGitHubClient) ProviderAndLabel() (string, string) {
	return "github", "pull request"
}

// PREvidenceForCommitV1 mirrors the REST API: returns an error for commits
// not present in PRsByCommit (matching the real GitHub V1 behaviour of
// returning 422 for unknown commits).
func (f *FakeGitHubClient) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	prs, ok := f.PRsByCommit[commit]
	if !ok {
		return nil, fmt.Errorf("commit not found: %s", commit)
	}
	return prs, nil
}

// PREvidenceForCommitHybrid mirrors the hybrid strategy: tries V2 (PRsByCommit)
// first; if empty, falls back through PRNumbersByCommit + PREvidenceByPRNumber.
// Returns empty (no error) when the commit is not found in either map.
func (f *FakeGitHubClient) PREvidenceForCommitHybrid(commit string) ([]*types.PREvidence, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	prs, err := f.PREvidenceForCommitV2(commit)
	if err != nil {
		return nil, err
	}
	if len(prs) > 0 {
		return prs, nil
	}
	// Fallback: use PRNumbersByCommit for V1-style discovery.
	prNumbers := f.PRNumbersByCommit[commit]
	result := []*types.PREvidence{}
	for _, n := range prNumbers {
		evidence, err := f.PREvidenceByPRNumber(n)
		if err != nil {
			return nil, err
		}
		if evidence != nil {
			result = append(result, evidence)
		}
	}
	return result, nil
}

// PREvidenceByPRNumber mirrors the GraphQL API: returns nil with no error for
// PR numbers not present in PRsByNumber (matching the real GitHub behaviour of
// returning null for non-existent pull requests).
func (f *FakeGitHubClient) PREvidenceByPRNumber(prNumber int) (*types.PREvidence, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	pr, ok := f.PRsByNumber[prNumber]
	if !ok {
		return nil, nil
	}
	return pr, nil
}

// PREvidenceForCommitV2 mirrors the GraphQL API: returns empty with no error
// for commits not present in PRsByCommit (matching the real GitHub V2 behaviour
// of returning null for unknown commits).
func (f *FakeGitHubClient) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	prs := f.PRsByCommit[commit]
	if prs == nil {
		return []*types.PREvidence{}, nil
	}
	return prs, nil
}
