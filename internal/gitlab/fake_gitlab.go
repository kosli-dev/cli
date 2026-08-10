package gitlab

import (
	"github.com/kosli-dev/cli/internal/types"
)

// FakeGitlabClient is an in-memory implementation of types.PRRetriever for
// testing. Seed MRsByCommit with the commits and merge request evidence you
// want returned. Set Err to simulate a network or API failure.
type FakeGitlabClient struct {
	// MRsByCommit maps a commit SHA to the merge request evidence returned
	// for that commit.
	MRsByCommit map[string][]*types.PREvidence
	// Err, if set, is returned by all calls regardless of commit.
	Err error
}

func (f *FakeGitlabClient) ProviderAndLabel() (string, string) {
	return "gitlab", "merge request"
}

// PREvidenceForCommitV2 mirrors the real client: a commit with no merge
// requests yields an empty result and no error, because GitLab's
// ListMergeRequestsByCommit returns an empty list rather than an error.
//
// The result is always non-nil. The real client builds its slice up front, and
// a nil slice would serialise as null, which the API rejects for pull_requests.
func (f *FakeGitlabClient) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	mrs := f.MRsByCommit[commit]
	if mrs == nil {
		return []*types.PREvidence{}, nil
	}
	return mrs, nil
}

// PREvidenceForCommitV1 mirrors the real client, which serves V1 from the same
// merge request lookup as V2.
func (f *FakeGitlabClient) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	return f.PREvidenceForCommitV2(commit)
}

// PREvidenceForCommitHybrid mirrors the real client, which has no V1 fallback
// for GitLab and always serves V2.
func (f *FakeGitlabClient) PREvidenceForCommitHybrid(commit string) ([]*types.PREvidence, error) {
	return f.PREvidenceForCommitV2(commit)
}
