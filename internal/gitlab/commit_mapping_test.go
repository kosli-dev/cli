package gitlab

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// TestCommitFromGitlabCommit_RecordsAuthorNotCommitter is a regression test for
// server#5479. PR commit attestations were recording the git committer in the
// author field. They must record the git author.
func TestCommitFromGitlabCommit_RecordsAuthorNotCommitter(t *testing.T) {
	created := time.Unix(1772635812, 0)
	commit := &gitlab.Commit{
		ID:             "abc1230000000000000000000000000000000000",
		Message:        "Apply suggestions from code review",
		AuthorName:     "Steve Tooke",
		AuthorEmail:    "tooky@kosli.com",
		CommitterName:  "GitHub",
		CommitterEmail: "noreply@github.com",
		CreatedAt:      &created,
		WebURL:         "https://gitlab.com/kosli/x/-/commit/abc123",
	}

	c := commitFromGitlabCommit(commit, "my-branch")

	require.Equal(t, "Steve Tooke <tooky@kosli.com>", c.Author,
		"author must be the git author, not the committer")
}
