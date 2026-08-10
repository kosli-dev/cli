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
