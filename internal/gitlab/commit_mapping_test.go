package gitlab

import (
	"errors"
	"net/http"
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

// TestCommitFromGitlabCommit_UsesAuthoredDate is a regression test for
// server#5479: the timestamp must match the recorded author identity.
func TestCommitFromGitlabCommit_UsesAuthoredDate(t *testing.T) {
	authored := time.Unix(1772630000, 0)
	created := time.Unix(1772635812, 0)
	commit := &gitlab.Commit{
		ID:           "abc1230000000000000000000000000000000000",
		Message:      "msg",
		AuthorName:   "Steve Tooke",
		AuthorEmail:  "tooky@kosli.com",
		AuthoredDate: &authored,
		CreatedAt:    &created,
		WebURL:       "https://gitlab.com/kosli/x/-/commit/abc123",
	}

	c := commitFromGitlabCommit(commit, "my-branch")

	require.Equal(t, int64(1772630000), c.Timestamp,
		"timestamp must be the authored date, not created_at")
}

// TestCommitFromGitlabCommit_FallsBackToCreatedAt ensures the timestamp falls
// back to created_at when the API omits the authored date.
func TestCommitFromGitlabCommit_FallsBackToCreatedAt(t *testing.T) {
	created := time.Unix(1772635812, 0)
	commit := &gitlab.Commit{
		ID:           "abc1230000000000000000000000000000000000",
		Message:      "msg",
		AuthorName:   "Steve Tooke",
		AuthorEmail:  "tooky@kosli.com",
		AuthoredDate: nil,
		CreatedAt:    &created,
		WebURL:       "https://gitlab.com/kosli/x/-/commit/abc123",
	}

	c := commitFromGitlabCommit(commit, "my-branch")

	require.Equal(t, int64(1772635812), c.Timestamp,
		"timestamp must fall back to created_at when authored date is absent")
}

// TestGitlabCommitVerification maps GitLab's verification_status to the neutral
// verified/signature_state fields (server#5892). verified is true only for a
// cryptographically valid signature; a non-"verified" status records
// verified=false (distinct from unsigned), and an empty status leaves both nil.
func TestGitlabCommitVerification(t *testing.T) {
	verified, state := gitlabCommitVerification("verified")
	require.NotNil(t, verified)
	require.True(t, *verified)
	require.NotNil(t, state)
	require.Equal(t, "verified", *state)

	verified, state = gitlabCommitVerification("unverified")
	require.NotNil(t, verified)
	require.False(t, *verified, "a non-'verified' status must record verified=false, not nil")
	require.Equal(t, "unverified", *state)

	verified, state = gitlabCommitVerification("")
	require.Nil(t, verified, "empty status (unsigned) must leave verified nil")
	require.Nil(t, state)
}

// TestResolveGitlabSignature covers the unsigned (404), fatal-error, and
// verified/unverified outcomes of a GetGPGSignature call (server#5892).
func TestResolveGitlabSignature(t *testing.T) {
	notFound := &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}
	verified, state, err := resolveGitlabSignature(nil, notFound, errors.New("404"))
	require.NoError(t, err, "404 (unsigned commit) must not be a fatal error")
	require.Nil(t, verified, "unsigned commit must leave verified nil")
	require.Nil(t, state)

	serverErr := &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}
	_, _, err = resolveGitlabSignature(nil, serverErr, errors.New("boom"))
	require.Error(t, err, "a non-404 error must propagate")

	_, _, err = resolveGitlabSignature(nil, nil, errors.New("transport failure"))
	require.Error(t, err, "an error with no response must propagate")

	verified, state, err = resolveGitlabSignature(
		&gitlab.GPGSignature{VerificationStatus: "verified"}, &gitlab.Response{}, nil,
	)
	require.NoError(t, err)
	require.NotNil(t, verified)
	require.True(t, *verified)
	require.Equal(t, "verified", *state)

	verified, _, err = resolveGitlabSignature(
		&gitlab.GPGSignature{VerificationStatus: "unverified"}, &gitlab.Response{}, nil,
	)
	require.NoError(t, err)
	require.NotNil(t, verified)
	require.False(t, *verified, "a non-'verified' status must record verified=false")
}

// TestGitlabMRRefs verifies the head (source) / base (target) branch mapping.
func TestGitlabMRRefs(t *testing.T) {
	head, base := gitlabMRRefs(&gitlab.BasicMergeRequest{
		SourceBranch: "feature-branch",
		TargetBranch: "main",
	})
	require.Equal(t, "feature-branch", head)
	require.Equal(t, "main", base)
}
