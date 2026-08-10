package gitlab

import (
	"testing"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// The real client builds its result slice up front, so a commit with no merge
// requests yields an empty list rather than nil. The fake must match: a nil
// slice serialises as null, and the API rejects null for pull_requests.
func TestFakeGitlabClientReturnsEmptyNotNilForUnknownCommit(t *testing.T) {
	client := &FakeGitlabClient{
		MRsByCommit: map[string][]*types.PREvidence{
			"known": {{URL: "https://gitlab.com/org/repo/-/merge_requests/1"}},
			// an explicitly seeded nil must be normalised too
			"seeded-nil": nil,
		},
	}

	for _, retrieve := range []struct {
		name string
		call func(string) ([]*types.PREvidence, error)
	}{
		{"V2", client.PREvidenceForCommitV2},
		{"V1", client.PREvidenceForCommitV1},
		{"Hybrid", client.PREvidenceForCommitHybrid},
	} {
		for _, commit := range []string{"unknown", "seeded-nil"} {
			t.Run(retrieve.name+"/"+commit, func(t *testing.T) {
				mrs, err := retrieve.call(commit)
				require.NoError(t, err)
				require.NotNil(t, mrs, "must be an empty slice, not nil")
				require.Empty(t, mrs)
			})
		}
	}
}
