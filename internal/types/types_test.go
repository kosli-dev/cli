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

// The server rejects an empty merge_commit on a PR that records reviews, so an
// open PR (no merge commit yet) must leave the field out rather than send "".
func TestPREvidenceMergeCommitWhenEmpty(t *testing.T) {
	for _, tc := range []struct {
		name     string
		evidence PREvidence
		want     *string
	}{
		{"recorded reviews, no merge commit: omitted", PREvidence{Reviews: &[]PRApprovals{}}, nil},
		{"recorded reviews, merge commit: kept", PREvidence{Reviews: &[]PRApprovals{}, MergeCommit: "abc"}, ptr("abc")},
		{"no reviews recorded, no merge commit: kept empty", PREvidence{}, ptr("")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.evidence)
			require.NoError(t, err)
			var decoded map[string]any
			require.NoError(t, json.Unmarshal(payload, &decoded))
			got, present := decoded["merge_commit"]
			if tc.want == nil {
				require.False(t, present, "merge_commit must be left out: %s", payload)
				return
			}
			require.Equal(t, *tc.want, got)
			require.Contains(t, decoded, "commits", "the other fields are still serialised")
		})
	}
}

func ptr(s string) *string { return &s }
