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
// Set Err to simulate a network or API failure.
type FakeGitHubClient struct {
	// PRsByCommit maps a commit SHA to the PR evidence returned for that commit.
	PRsByCommit map[string][]*types.PREvidence
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
