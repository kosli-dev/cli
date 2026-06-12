package azure

import (
	"testing"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/stretchr/testify/require"
)

func azStr(s string) *string { return &s }

func azTestCommit() git.GitCommitRef {
	authorDate := azuredevops.Time{Time: time.Unix(1772630000, 0)}
	committerDate := azuredevops.Time{Time: time.Unix(1772635812, 0)}
	return git.GitCommitRef{
		CommitId: azStr("abc1230000000000000000000000000000000000"),
		Comment:  azStr("Apply suggestions from code review"),
		Url:      azStr("https://dev.azure.com/kosli/_git/x/commit/abc123"),
		Author: &git.GitUserDate{
			Name:  azStr("Steve Tooke"),
			Email: azStr("tooky@kosli.com"),
			Date:  &authorDate,
		},
		Committer: &git.GitUserDate{
			Name:  azStr("GitHub"),
			Email: azStr("noreply@github.com"),
			Date:  &committerDate,
		},
	}
}

// TestCommitFromAzureCommit_RecordsAuthorIdentity is a regression test for
// server#5479. Azure recorded the author display name only, and populated
// author_username from the committer. Record the author as "Name <email>",
// and drop author_username (Azure commits carry no login).
func TestCommitFromAzureCommit_RecordsAuthorIdentity(t *testing.T) {
	c := commitFromAzureCommit(azTestCommit(), "my-branch")

	require.Equal(t, "Steve Tooke <tooky@kosli.com>", c.Author,
		"author must be name <email> of the git author")
	require.Empty(t, c.AuthorUsername,
		"Azure commits carry no login; author_username must be omitted, not the committer name")
}
