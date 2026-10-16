package azure

import (
	"testing"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/stretchr/testify/require"
)

func azureReviewer(name string, vote int, isGroup bool) git.IdentityRefWithVote {
	unique := name + "@example.com"
	return git.IdentityRefWithVote{DisplayName: &name, UniqueName: &unique, Vote: &vote, IsContainer: &isGroup}
}

// Votes per IdentityRefWithVote: 10 approved, 5 approved with suggestions,
// 0 no vote, -5 waiting for author, -10 rejected.
func TestAzureReviews(t *testing.T) {
	reviewers := []git.IdentityRefWithVote{
		azureReviewer("Ada", 10, false),
		azureReviewer("Bob", 5, false),
		azureReviewer("Cy", 0, false),
		azureReviewer("Di", -5, false),
		azureReviewer("Ed", -10, false),
		azureReviewer("Team", 10, true),
	}
	require.Equal(t, []types.PRApprovals{
		{Username: "Ada (Ada@example.com)", State: types.ReviewApproved},
		{Username: "Bob (Bob@example.com)", State: types.ReviewApproved},
		{Username: "Di (Di@example.com)", State: types.ReviewChangesRequested},
		{Username: "Ed (Ed@example.com)", State: types.ReviewChangesRequested},
	}, azureReviews(reviewers))
}
