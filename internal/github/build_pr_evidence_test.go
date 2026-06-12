package github

import (
	"testing"

	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/require"
)

// TestBuildPREvidence_RecordsAuthorNotCommitter is a regression test for
// server#5479. PR commit attestations were recording the git committer in the
// "author" field. For GitHub web-flow commits (applied suggestions, bot
// commits) the committer is "GitHub <noreply@github.com>", distinct from the
// real author — so the true author was being lost and the author_username
// dropped entirely (the committer has no associated GitHub user).
func TestBuildPREvidence_RecordsAuthorNotCommitter(t *testing.T) {
	node := graphqlCommitNode{}
	node.Commit.Oid = "0e723254516c841126e81f76100be57258ff1386"
	node.Commit.MessageHeadline = "Apply suggestions from code review"
	node.Commit.CommittedDate = "2026-03-01T12:00:00Z"
	node.Commit.URL = "https://github.com/kosli-dev/cli/commit/0e723254516c841126e81f76100be57258ff1386"

	// Committer is GitHub's web-flow identity, with no associated user.
	node.Commit.Committer.Name = "GitHub"
	node.Commit.Committer.Email = "noreply@github.com"
	node.Commit.Committer.User = nil

	// Author is the real person who wrote the change.
	node.Commit.Author.Name = "Steve Tooke"
	node.Commit.Author.Email = "tooky@kosli.com"
	node.Commit.Author.User = &struct {
		Login graphql.String
	}{Login: "tooky"}

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED",
		"tooky",
		"2026-03-01T11:00:00Z",
		"",
		"Introduce kosli evaluate",
		"introduce-kosli-evaluate",
		[]graphqlCommitNode{node},
		nil,
	)
	require.NoError(t, err)
	require.Len(t, evidence.Commits, 1)

	c := evidence.Commits[0]
	require.Equal(t, "Steve Tooke <tooky@kosli.com>", c.Author,
		"the commit author (wire field author) must be the git author, not the committer")
	require.Equal(t, "tooky", c.AuthorUsername,
		"author_username must be the author's GitHub login, not the committer's (absent) one")
}
