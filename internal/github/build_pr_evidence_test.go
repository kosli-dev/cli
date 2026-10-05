package github

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/types"
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
		"main", "",
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
		"Introduce kosli evaluate", "introduce-kosli-evaluate", "main", "",
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
		"title", "branch", "main", "",
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
		"title", "feature-branch", "main", "",
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
	node.Commit.Signature = &graphqlSignature{IsValid: true, State: "VALID"}

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "feature", "main", "",
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

// TestBuildPREvidence_EmptyAuthorIsPreserved verifies that when the PR
// creator's GitHub account has been deleted (Author.Login = ""), the empty
// string is preserved and serialised as "" rather than omitted, allowing the
// server to accept it directly.
func TestBuildPREvidence_EmptyAuthorIsPreserved(t *testing.T) {
	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED",
		"", // empty author — PR creator account deleted
		"2026-03-01T09:00:00Z",
		"2026-03-01T12:00:00Z",
		"Fix something",
		"fix-branch",
		"main", "",
		nil, nil,
	)
	require.NoError(t, err)
	require.Equal(t, "", evidence.Author,
		"empty PR author login must be preserved so it is serialised as an empty string, not omitted")

	b, err := json.Marshal(evidence)
	require.NoError(t, err)
	require.Contains(t, string(b), `"author":""`,
		"empty author must be serialised, not omitted")
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
		"title", "feature", "main", "",
		[]graphqlCommitNode{node}, nil,
	)
	require.NoError(t, err)
	require.Len(t, evidence.Commits, 1)
	require.Nil(t, evidence.Commits[0].Verified, "unsigned commit must leave verified nil")
	require.Nil(t, evidence.Commits[0].SignatureState)
}

// An approval whose commit GitHub no longer has must leave commit_sha out of
// the payload rather than send an empty string, and so must an unknown head.
func TestBuildPREvidence_OmitsUnknownReviewedAndHeadCommits(t *testing.T) {
	review := reviewNode("User", "grace", "APPROVED")

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "feature-branch", "main", "",
		nil, []graphqlReviewNode{review},
	)
	require.NoError(t, err)

	payload, err := json.Marshal(evidence)
	require.NoError(t, err)
	require.Contains(t, string(payload), `"username":"grace"`)
	require.NotContains(t, string(payload), "commit_sha")
	require.NotContains(t, string(payload), "head_sha")
}

func signedCommitNode(sha string, sig *graphqlSignature) graphqlCommitNode {
	node := graphqlCommitNode{}
	node.Commit.Oid = graphql.String(sha)
	node.Commit.CommittedDate = "2026-03-01T12:00:00Z"
	node.Commit.Signature = sig
	return node
}

// A verified signature names who signed, which the commit's own author
// fields cannot: whoever writes the commit sets those.
func TestBuildPREvidence_RecordsCommitSigner(t *testing.T) {
	bySigner := &graphqlSignature{IsValid: true, State: "VALID"}
	bySigner.Signer = &struct{ Login graphql.String }{Login: "alice"}
	byGitHub := &graphqlSignature{IsValid: true, State: "VALID", WasSignedByGitHub: true}
	byGitHub.Signer = &struct{ Login graphql.String }{Login: "web-flow"}

	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "feature", "main", "",
		[]graphqlCommitNode{
			signedCommitNode("1111111111111111111111111111111111111111", bySigner),
			signedCommitNode("2222222222222222222222222222222222222222", byGitHub),
			signedCommitNode("3333333333333333333333333333333333333333", nil),
		}, nil,
	)
	require.NoError(t, err)
	require.Len(t, evidence.Commits, 3)

	require.Equal(t, "alice", evidence.Commits[0].SignerUsername)
	require.NotNil(t, evidence.Commits[0].SignedByPlatform)
	require.False(t, *evidence.Commits[0].SignedByPlatform)

	require.Equal(t, "", evidence.Commits[1].SignerUsername,
		"GitHub's own signing account is not who made the commit")
	require.NotNil(t, evidence.Commits[1].SignedByPlatform)
	require.True(t, *evidence.Commits[1].SignedByPlatform)

	unsigned, err := json.Marshal(evidence.Commits[2])
	require.NoError(t, err)
	require.NotContains(t, string(unsigned), "signer_username")
	require.NotContains(t, string(unsigned), "signed_by_platform")
}

func reviewNode(typename, login, state string) graphqlReviewNode {
	r := graphqlReviewNode{State: graphql.String(state), SubmittedAt: "2026-03-01T13:00:00Z"}
	r.Author.Typename = graphql.String(typename)
	r.Author.Login = graphql.String(login)
	return r
}

// Only a person's approval counts: a bot's does not, and nor does a later
// request for changes, which GitHub returns as that reviewer's latest review.
func TestBuildPREvidence_KeepsOnlyHumanApprovals(t *testing.T) {
	evidence, err := buildPREvidence(
		"https://github.com/kosli-dev/cli/pull/671",
		"0e723254516c841126e81f76100be57258ff1386",
		"MERGED", "tooky", "2026-03-01T09:00:00Z", "",
		"title", "feature", "main", "",
		nil, []graphqlReviewNode{
			reviewNode("User", "grace", "APPROVED"),
			reviewNode("Bot", "github-actions", "APPROVED"),
			reviewNode("User", "linus", "CHANGES_REQUESTED"),
		},
	)
	require.NoError(t, err)
	require.Len(t, evidence.Approvers, 1)
	require.Equal(t, "grace", evidence.Approvers[0].(types.PRApprovals).Username)
}
