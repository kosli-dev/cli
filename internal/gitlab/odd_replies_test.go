package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// Baseline replies for a merged MR (iid 101) with 2 commits, 2 reviewers and
// 1 approval, so every field PR evidence reads is present. Shapes follow
// GitLab's documented examples; values are made up.

// https://docs.gitlab.com/api/commits/#list-merge-requests-associated-with-a-commit
const oddMRsForCommit = `[{
  "id": 4501, "iid": 101, "project_id": 35,
  "title": "Add audit log", "description": "Adds an audit log.",
  "state": "merged",
  "created_at": "2026-03-26T17:26:30.916Z", "updated_at": "2026-03-27T09:12:00.000Z",
  "merged_at": "2026-03-27T09:12:00.000Z",
  "merge_user": {"id": 28, "username": "janedoe", "name": "Jane Doe", "state": "active",
    "avatar_url": "https://gitlab.example.com/uploads/user/avatar/28/a.png", "web_url": "https://gitlab.example.com/janedoe"},
  "closed_by": null, "closed_at": null,
  "target_branch": "main", "source_branch": "audit-log",
  "upvotes": 0, "downvotes": 0,
  "author": {"id": 28, "username": "janedoe", "name": "Jane Doe", "state": "active",
    "avatar_url": "https://gitlab.example.com/uploads/user/avatar/28/a.png", "web_url": "https://gitlab.example.com/janedoe"},
  "assignee": null,
  "source_project_id": 35, "target_project_id": 35,
  "labels": ["audit"], "draft": false, "milestone": null,
  "merge_when_pipeline_succeeds": false, "detailed_merge_status": "not_open",
  "sha": "c2a6f2b1e0d94e3f8a7b6c5d4e3f2a1b0c9d8e7f",
  "merge_commit_sha": "9f8e7d6c5b4a39281706f5e4d3c2b1a098765432",
  "squash_commit_sha": null,
  "user_notes_count": 2, "discussion_locked": null,
  "should_remove_source_branch": true, "force_remove_source_branch": false,
  "web_url": "https://gitlab.example.com/acme/app/-/merge_requests/101",
  "time_stats": {"time_estimate": 0, "total_time_spent": 0, "human_time_estimate": null, "human_total_time_spent": null}
}]`

// https://docs.gitlab.com/api/merge_request_approvals/#retrieve-approval-state-for-a-merge-request
const oddApprovals = `{
  "id": 4501, "iid": 101, "project_id": 35,
  "title": "Add audit log", "description": "Adds an audit log.",
  "state": "merged",
  "created_at": "2026-03-26T17:26:30.916Z", "updated_at": "2026-03-27T09:12:00.000Z",
  "merge_status": "can_be_merged",
  "approved": true, "approvals_required": 1, "approvals_left": 0,
  "approved_by": [{"user": {"id": 1, "username": "user1", "name": "John Doe1", "state": "active",
    "avatar_url": "https://gitlab.example.com/uploads/user/avatar/1/a.png", "web_url": "https://gitlab.example.com/user1"}}]
}`

// https://docs.gitlab.com/api/merge_requests/#get-single-merge-request-reviewers
const oddReviewers = `[
  {"user": {"id": 1, "username": "user1", "name": "John Doe1", "state": "active",
    "avatar_url": "https://gitlab.example.com/uploads/user/avatar/1/a.png", "web_url": "https://gitlab.example.com/user1"},
   "state": "approved", "created_at": "2026-03-26T18:03:27.684Z"},
  {"user": {"id": 2, "username": "user2", "name": "John Doe2", "state": "active",
    "avatar_url": "https://gitlab.example.com/uploads/user/avatar/2/a.png", "web_url": "https://gitlab.example.com/user2"},
   "state": "reviewed", "created_at": "2026-03-26T18:03:27.684Z"}
]`

// https://docs.gitlab.com/api/merge_requests/#get-single-merge-request-commits
const oddMRCommits = `[
  {"id": "ed899a2f4b50b4370feeea94676502b42383c746", "short_id": "ed899a2f4b5",
   "title": "Add audit log table", "message": "Add audit log table\n",
   "author_name": "Example User", "author_email": "user@example.com", "authored_date": "2026-03-26T11:50:22+03:00",
   "committer_name": "Example User", "committer_email": "user@example.com", "committed_date": "2026-03-26T11:50:22+03:00",
   "created_at": "2026-03-26T11:50:22+03:00",
   "parent_ids": ["ae1d9fb46aa2b07ee9836d49862ec4e2c46fbbba"],
   "trailers": {}, "extended_trailers": {},
   "web_url": "https://gitlab.example.com/acme/app/-/commit/ed899a2f4b50b4370feeea94676502b42383c746"},
  {"id": "6104942438c14ec7bd21c6cd5bd995272b3faff6", "short_id": "6104942438c",
   "title": "Write audit entries", "message": "Write audit entries\n",
   "author_name": "Example User", "author_email": "user@example.com", "authored_date": "2026-03-26T12:10:01+03:00",
   "committer_name": "Example User", "committer_email": "user@example.com", "committed_date": "2026-03-26T12:10:01+03:00",
   "created_at": "2026-03-26T12:10:01+03:00",
   "parent_ids": ["ed899a2f4b50b4370feeea94676502b42383c746"],
   "trailers": {}, "extended_trailers": {},
   "web_url": "https://gitlab.example.com/acme/app/-/commit/6104942438c14ec7bd21c6cd5bd995272b3faff6"}
]`

// https://docs.gitlab.com/api/commits/#retrieve-commit-signature
// Served for every commit in oddMRCommits; a variant changes both.
const oddSignature = `{
  "signature_type": "PGP", "verification_status": "verified",
  "gpg_key_id": 1, "gpg_key_primary_keyid": "8254AAB3FBD54AC9",
  "gpg_key_user_name": "John Doe", "gpg_key_user_email": "johndoe@example.com",
  "gpg_key_subkey_id": null, "commit_source": "gitaly"
}`

const oddCommit = "c2a6f2b1e0d94e3f8a7b6c5d4e3f2a1b0c9d8e7f"

// oddReplyNames are the replies in the order PREvidenceForCommitV2 asks.
var oddReplyNames = []string{"mrs_for_commit", "approvals", "reviewers", "mr_commits", "signature"}

func oddBaseline() map[string]string {
	return map[string]string{
		"mrs_for_commit": oddMRsForCommit,
		"approvals":      oddApprovals,
		"reviewers":      oddReviewers,
		"mr_commits":     oddMRCommits,
		"signature":      oddSignature,
	}
}

// oddRepliesServer answers like GitLab: only MR 101 and the two baseline
// commits exist, anything else is a 404.
func oddRepliesServer(t *testing.T, replies map[string]string) *httptest.Server {
	t.Helper()
	knownSHA := map[string]bool{
		"ed899a2f4b50b4370feeea94676502b42383c746": true,
		"6104942438c14ec7bd21c6cd5bd995272b3faff6": true,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		reply := ""
		switch {
		case strings.HasSuffix(p, "/repository/commits/"+oddCommit+"/merge_requests"):
			reply = "mrs_for_commit"
		case strings.HasSuffix(p, "/merge_requests/101/approvals"):
			reply = "approvals"
		case strings.HasSuffix(p, "/merge_requests/101/reviewers"):
			reply = "reviewers"
		case strings.HasSuffix(p, "/merge_requests/101/commits"):
			reply = "mr_commits"
		case strings.HasSuffix(p, "/signature"):
			sha := strings.TrimSuffix(p[strings.LastIndex(p, "/commits/")+len("/commits/"):], "/signature")
			if knownSHA[sha] {
				reply = "signature"
			}
		}
		if reply == "" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"message":"404 Not Found"}`)
			return
		}
		_, _ = fmt.Fprint(w, replies[reply])
	}))
	t.Cleanup(ts.Close)
	return ts
}

func oddEvidence(t *testing.T, replies map[string]string) ([]*types.PREvidence, error) {
	ts := oddRepliesServer(t, replies)
	return newPaginationConfig(ts.URL).PREvidenceForCommitV2(oddCommit)
}

func TestOddReplies_BaselineGivesFullEvidence(t *testing.T) {
	evidence, err := oddEvidence(t, oddBaseline())
	require.NoError(t, err)
	require.Len(t, evidence, 1)
	e := evidence[0]
	require.Equal(t, "https://gitlab.example.com/acme/app/-/merge_requests/101", e.URL)
	require.Equal(t, "9f8e7d6c5b4a39281706f5e4d3c2b1a098765432", e.MergeCommit)
	require.Equal(t, "merged", e.State)
	require.Equal(t, "Jane Doe (@janedoe)", e.Author)
	require.NotZero(t, e.CreatedAt)
	require.NotZero(t, e.MergedAt)
	require.Equal(t, "Add audit log", e.Title)
	require.Equal(t, "audit-log", e.HeadRef)
	require.Equal(t, "main", e.BaseRef)
	require.Equal(t, []any{types.PRApprovals{Username: "John Doe1 (@user1)"}}, e.Approvers)
	require.Equal(t, []types.PRApprovals{
		{Username: "John Doe1 (@user1)", State: types.ReviewApproved},
		{Username: "John Doe2 (@user2)", State: types.ReviewCommented},
	}, e.Reviews)
	require.Len(t, e.Commits, 2)
	for _, c := range e.Commits {
		require.NotEmpty(t, c.SHA)
		require.NotZero(t, c.Timestamp)
		require.NotNil(t, c.Verified)
		require.True(t, *c.Verified)
	}
}

// oddRuleExceptions are the replies whose outcome is the opposite of the
// rule's: an error for a null field, or no error for a null list item. An
// error for a null field must name the list item it makes unreadable.
var oddRuleExceptions = map[string]struct{ why, names string }{
	"mrs_for_commit [0].iid":       {why: "error: the iid names the MR in every later request; null asks for MR 0, which 404s"},
	"mrs_for_commit [0].labels[0]": {why: "ok: labels are not read, and the client reads a null string as empty"},
	"mr_commits [0].parent_ids[0]": {why: "ok: parent ids are not read, and the client reads a null string as empty"},
	"mr_commits [1].parent_ids[0]": {why: "ok: parent ids are not read, and the client reads a null string as empty"},
	"mr_commits [0].id":            {why: "error: a commit without a SHA can't be looked up or attested", names: "commits[0]"},
	"mr_commits [1].id":            {why: "error: a commit without a SHA can't be looked up or attested", names: "commits[1]"},
}

var listItemPath = regexp.MustCompile(`([A-Za-z_]*\[\d+\])$`)

// TestOddReplies feeds every entry point that commands use each reply with one
// value made null, or one list emptied. None may panic. A null field or an
// emptied list gives no error; a null list item gives an error naming it.
func TestOddReplies(t *testing.T) {
	// `kosli assert pullrequest gitlab` calls V2 and `attest` calls Hybrid.
	entryPoints := map[string]func(*GitlabConfig) func(string) ([]*types.PREvidence, error){
		"V2":     func(c *GitlabConfig) func(string) ([]*types.PREvidence, error) { return c.PREvidenceForCommitV2 },
		"Hybrid": func(c *GitlabConfig) func(string) ([]*types.PREvidence, error) { return c.PREvidenceForCommitHybrid },
	}
	for entry, fetchFor := range entryPoints {
		for _, name := range oddReplyNames {
			variants, err := testHelpers.NullVariants([]byte(oddBaseline()[name]))
			require.NoError(t, err)
			for _, v := range variants {
				key := name + " " + v.Path
				t.Run(entry+"/"+key, func(t *testing.T) {
					replies := oddBaseline()
					replies[name] = string(v.Body)
					ts := oddRepliesServer(t, replies)
					var fetchErr error
					outcome := testHelpers.Outcome(func() error {
						_, fetchErr = fetchFor(newPaginationConfig(ts.URL))(oddCommit)
						return fetchErr
					})
					require.NotContains(t, outcome, "panic")

					item := listItemPath.FindString(v.Path)
					want := "ok"
					if item != "" {
						want = "error"
					}
					if ex, ok := oddRuleExceptions[key]; ok {
						want = map[string]string{"ok": "error", "error": "ok"}[want]
						item = ex.names
						t.Log(ex.why)
					}
					require.Equal(t, want, outcome)
					if item != "" && want == "error" {
						require.ErrorContains(t, fetchErr, item)
					}
				})
			}
		}
	}
}

// A review whose account GitLab no longer names (a deleted user) is never an
// approver, and is kept in Reviews, without a name, only if it requests changes.
func TestOddReplies_ReviewByNoOne(t *testing.T) {
	approver := types.PRApprovals{Username: "John Doe1 (@user1)"}
	approved := types.PRApprovals{Username: "John Doe1 (@user1)", State: types.ReviewApproved}
	commented := types.PRApprovals{Username: "John Doe2 (@user2)", State: types.ReviewCommented}
	tests := []struct {
		name          string
		nulls         map[string]string
		wantApprovers []any
		wantReviews   []types.PRApprovals
	}{
		{"approval with no user", map[string]string{"approvals": "approved_by[0].user"},
			nil, []types.PRApprovals{commented}},
		{"approval with no username", map[string]string{"approvals": "approved_by[0].user.username"},
			nil, []types.PRApprovals{commented}},
		{"comment with no user", map[string]string{"reviewers": "[1].user"},
			[]any{approver}, []types.PRApprovals{approved}},
		{"comment with no username", map[string]string{"reviewers": "[1].user.username"},
			[]any{approver}, []types.PRApprovals{approved}},
		{"approval and comment by no one", map[string]string{"approvals": "approved_by[0].user", "reviewers": "[1].user"},
			nil, []types.PRApprovals{}},
		{"change request by no one is kept", map[string]string{"reviewers": "[1].user", "requested_changes": ""},
			[]any{approver}, []types.PRApprovals{approved, {State: types.ReviewChangesRequested}}},
		{"change request by no one is kept beside an approval by no one", map[string]string{"approvals": "approved_by[0].user", "reviewers": "[1].user", "requested_changes": ""},
			nil, []types.PRApprovals{{State: types.ReviewChangesRequested}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			replies := oddBaseline()
			if _, ok := tc.nulls["requested_changes"]; ok {
				replies["reviewers"] = strings.Replace(replies["reviewers"], `"state": "reviewed"`, `"state": "requested_changes"`, 1)
			}
			for reply, path := range tc.nulls {
				if reply == "requested_changes" {
					continue
				}
				variants, err := testHelpers.NullVariants([]byte(replies[reply]))
				require.NoError(t, err)
				found := false
				for _, v := range variants {
					if v.Path == path {
						replies[reply], found = string(v.Body), true
					}
				}
				require.True(t, found, path)
			}
			evidence, err := oddEvidence(t, replies)
			require.NoError(t, err)
			require.Len(t, evidence, 1)
			require.Equal(t, tc.wantApprovers, evidence[0].Approvers)
			require.Equal(t, tc.wantReviews, evidence[0].Reviews)
		})
	}
}
