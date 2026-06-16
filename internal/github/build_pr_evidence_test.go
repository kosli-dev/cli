package github

import (
	"testing"
	"time"

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

	// Author is the real person who wrote the change. The query no longer
	// fetches the committer at all (it would be GitHub's web-flow identity for
	// applied-suggestion / bot commits), so only the author is recorded.
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
		"main",
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

// TestBuildPREvidence_UsesAuthoredDate is a regression test for server#5479.
// Now that the recorded identity is the author, the timestamp should be the
// author date too. For rebased / applied-suggestion commits the authored and
// committed dates differ.
func TestBuildPREvidence_UsesAuthoredDate(t *testing.T) {
	node := graphqlCommitNode{}
	node.Commit.Oid = "0e723254516c841126e81f76100be57258ff1386"
	node.Commit.MessageHeadline = "Apply suggestions from code review"
	node.Commit.AuthoredDate = "2026-03-01T10:00:00Z"
	node.Commit.CommittedDate = "2026-03-01T12:00:00Z"
	node.Commit.Author.Name = "Steve Tooke"
	node.Commit.Author.Email = "tooky@kosli.com"

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"Introduce kosli evaluate", "introduce-kosli-evaluate", "main",
		[]graphqlCommitNode{node}, nil,
	)
	require.NoError(t, err)
	require.Len(t, evidence.Commits, 1)

	wantAuthored, _ := time.Parse(time.RFC3339, "2026-03-01T10:00:00Z")
	require.Equal(t, wantAuthored.Unix(), evidence.Commits[0].Timestamp,
		"timestamp must be the author date, not the committer date")
}

// TestBuildPREvidence_FallsBackToCommittedDate ensures the timestamp falls back
// to the committed date when the GraphQL response omits the authored date.
func TestBuildPREvidence_FallsBackToCommittedDate(t *testing.T) {
	node := graphqlCommitNode{}
	node.Commit.Oid = "0e723254516c841126e81f76100be57258ff1386"
	node.Commit.MessageHeadline = "msg"
	node.Commit.AuthoredDate = ""
	node.Commit.CommittedDate = "2026-03-01T12:00:00Z"
	node.Commit.Author.Name = "Steve Tooke"
	node.Commit.Author.Email = "tooky@kosli.com"

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "branch", "main",
		[]graphqlCommitNode{node}, nil,
	)
	require.NoError(t, err)
	require.Len(t, evidence.Commits, 1)

	wantCommitted, _ := time.Parse(time.RFC3339, "2026-03-01T12:00:00Z")
	require.Equal(t, wantCommitted.Unix(), evidence.Commits[0].Timestamp,
		"timestamp must fall back to the committed date when authored date is absent")
}

// TestBuildPREvidence_RecordsBaseRef verifies the PR's base (target) branch is
// captured, enabling a "merged into main" policy (server#5892).
func TestBuildPREvidence_RecordsBaseRef(t *testing.T) {
	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "feature-branch", "main",
		nil, nil,
	)
	require.NoError(t, err)
	require.Equal(t, "main", evidence.BaseRef,
		"base_ref must record the PR target branch")
}

// TestBuildPREvidence_RecordsCommitSignature verifies a verified commit
// signature is captured (server#5892, control 1.13).
func TestBuildPREvidence_RecordsCommitSignature(t *testing.T) {
	node := graphqlCommitNode{}
	node.Commit.Oid = "0e723254516c841126e81f76100be57258ff1386"
	node.Commit.MessageHeadline = "signed work"
	node.Commit.CommittedDate = "2026-03-01T12:00:00Z"
	node.Commit.Author.Name = "Steve Tooke"
	node.Commit.Author.Email = "tooky@kosli.com"
	node.Commit.Signature = &struct {
		IsValid graphql.Boolean
		State   graphql.String
	}{IsValid: true, State: "VALID"}

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "feature", "main",
		[]graphqlCommitNode{node}, nil,
	)
	require.NoError(t, err)
	require.Len(t, evidence.Commits, 1)
	c := evidence.Commits[0]
	require.NotNil(t, c.Verified, "verified must be populated for a signed commit")
	require.True(t, *c.Verified, "verified must be true for a valid signature")
	require.NotNil(t, c.SignatureState)
	require.Equal(t, "VALID", *c.SignatureState)
}

// TestBuildPREvidence_UnsignedCommitHasNoSignatureFields verifies an unsigned
// commit (no signature node) leaves verified/signature_state nil, so "unsigned"
// stays distinct from "present-but-invalid" (verified=false).
func TestBuildPREvidence_UnsignedCommitHasNoSignatureFields(t *testing.T) {
	node := graphqlCommitNode{}
	node.Commit.Oid = "0e723254516c841126e81f76100be57258ff1386"
	node.Commit.MessageHeadline = "unsigned work"
	node.Commit.CommittedDate = "2026-03-01T12:00:00Z"
	node.Commit.Author.Name = "Steve Tooke"
	node.Commit.Author.Email = "tooky@kosli.com"
	// node.Commit.Signature left nil — unsigned commit

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "feature", "main",
		[]graphqlCommitNode{node}, nil,
	)
	require.NoError(t, err)
	require.Len(t, evidence.Commits, 1)
	require.Nil(t, evidence.Commits[0].Verified, "unsigned commit must leave verified nil")
	require.Nil(t, evidence.Commits[0].SignatureState)
}
