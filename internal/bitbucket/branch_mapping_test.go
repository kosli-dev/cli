package bitbucket

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBitbucketBranchName verifies head (source) and base (destination) branch
// extraction from a Bitbucket PR response (server#5892).
func TestBitbucketBranchName(t *testing.T) {
	prData := map[string]any{
		"source": map[string]any{
			"branch": map[string]any{"name": "feature-branch"},
		},
		"destination": map[string]any{
			"branch": map[string]any{"name": "main"},
		},
	}
	require.Equal(t, "feature-branch", bitbucketBranchName(prData, "source"))
	require.Equal(t, "main", bitbucketBranchName(prData, "destination"))
}
