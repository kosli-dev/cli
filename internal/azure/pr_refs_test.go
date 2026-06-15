package azure

import (
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/stretchr/testify/require"
)

// TestAzurePRRefs verifies head (source) and base (target) ref extraction. The
// refs are kept raw with the refs/heads/ prefix to match head_ref (server#5892).
func TestAzurePRRefs(t *testing.T) {
	head, base := azurePRRefs(git.GitPullRequest{
		SourceRefName: azStr("refs/heads/feature-branch"),
		TargetRefName: azStr("refs/heads/main"),
	})
	require.Equal(t, "refs/heads/feature-branch", head)
	require.Equal(t, "refs/heads/main", base)
}
