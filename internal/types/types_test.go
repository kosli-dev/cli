package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// The API validates pull request attestations against FoundPullRequestV2, which
// requires "commits". Dropping the field when a provider returns no commits
// produces a payload that matches neither V1 nor V2 and is rejected (#1081).
func TestPREvidenceAlwaysSerialisesCommits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		evidence PREvidence
		want     string
	}{
		{
			name:     "empty commits serialise as an empty array",
			evidence: PREvidence{Commits: []Commit{}},
			want:     `[]`,
		},
		{
			name:     "nil commits serialise as an empty array",
			evidence: PREvidence{},
			want:     `[]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.evidence)
			require.NoError(t, err)

			var decoded map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(payload, &decoded))

			raw, present := decoded["commits"]
			require.True(t, present, "commits must always be present in the payload")
			require.JSONEq(t, tc.want, string(raw))
		})
	}
}

func TestPREvidenceOmitsReviewsWhenNoneAreRecorded(t *testing.T) {
	for _, reviews := range [][]PRApprovals{nil, {}} {
		payload, err := json.Marshal(PREvidence{Reviews: reviews})
		require.NoError(t, err)
		require.NotContains(t, string(payload), `"reviews"`)
	}
}

func TestApprovalsFromKeepsTheLatestApprovalOfEachCommitPerReviewer(t *testing.T) {
	reviews := []PRApprovals{
		{Username: "ada", State: ReviewApproved, Timestamp: 1, CommitSHA: "c1"},
		{Username: "ada", State: ReviewChangesRequested, Timestamp: 2, CommitSHA: "c1"},
		{Username: "ada", State: ReviewApproved, Timestamp: 3, CommitSHA: "c1"},
		{Username: "bob", State: ReviewCommented, Timestamp: 4, CommitSHA: "c1"},
		{Username: "bob", State: ReviewDismissed, Timestamp: 5, CommitSHA: "c1"},
		{Username: "bob", State: ReviewApproved, Timestamp: 6, CommitSHA: "c2"},
		{Username: "ada", State: ReviewApproved, Timestamp: 7, CommitSHA: "c2"},
	}
	require.Equal(t, []any{
		PRApprovals{Username: "ada", State: ReviewApproved, Timestamp: 3, CommitSHA: "c1"},
		PRApprovals{Username: "bob", State: ReviewApproved, Timestamp: 6, CommitSHA: "c2"},
		PRApprovals{Username: "ada", State: ReviewApproved, Timestamp: 7, CommitSHA: "c2"},
	}, ApprovalsFrom(reviews))
}

func TestApprovalsFromReturnsAnEmptyListWithNoApprovals(t *testing.T) {
	require.Equal(t, []any{}, ApprovalsFrom(nil))
	require.Equal(t, []any{}, ApprovalsFrom([]PRApprovals{{Username: "ada", State: ReviewCommented}}))
}
