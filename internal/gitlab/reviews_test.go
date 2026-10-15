package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func approvedBy(usernames ...string) []*gitlab.MergeRequestApproverUser {
	users := []*gitlab.MergeRequestApproverUser{}
	for _, u := range usernames {
		users = append(users, &gitlab.MergeRequestApproverUser{User: &gitlab.BasicUser{Name: strings.ToUpper(u), Username: u}})
	}
	return users
}

func reviewer(username, state string) *gitlab.MergeRequestReviewer {
	return &gitlab.MergeRequestReviewer{User: &gitlab.BasicUser{Name: strings.ToUpper(username), Username: username}, State: state}
}

// Reviewer states are GitLab's MergeRequestReviewerState enum: unreviewed,
// reviewed, requested_changes, approved, unapproved, review_started.
func TestGitlabReviews(t *testing.T) {
	for _, tc := range []struct {
		name       string
		approvedBy []*gitlab.MergeRequestApproverUser
		reviewers  []*gitlab.MergeRequestReviewer
		want       []types.PRApprovals
	}{
		{
			name:       "an approval comes from approved_by, reviewer or not",
			approvedBy: approvedBy("ada", "bob"),
			reviewers:  []*gitlab.MergeRequestReviewer{reviewer("ada", "approved")},
			want: []types.PRApprovals{
				{Username: "ADA (@ada)", State: types.ReviewApproved},
				{Username: "BOB (@bob)", State: types.ReviewApproved},
			},
		},
		{
			name:      "a reviewer state of approved is not an approval unless approved_by says so",
			reviewers: []*gitlab.MergeRequestReviewer{reviewer("ada", "approved")},
			want:      []types.PRApprovals{},
		},
		{
			name: "requested_changes and reviewed map to the shared states",
			reviewers: []*gitlab.MergeRequestReviewer{
				reviewer("ada", "requested_changes"), reviewer("bob", "reviewed"),
			},
			want: []types.PRApprovals{
				{Username: "ADA (@ada)", State: types.ReviewChangesRequested},
				{Username: "BOB (@bob)", State: types.ReviewCommented},
			},
		},
		{
			name:       "a change request outweighs an approval by the same person",
			approvedBy: approvedBy("ada"),
			reviewers:  []*gitlab.MergeRequestReviewer{reviewer("ada", "requested_changes")},
			want:       []types.PRApprovals{{Username: "ADA (@ada)", State: types.ReviewChangesRequested}},
		},
		{
			name: "people are matched on username, not on display name",
			approvedBy: []*gitlab.MergeRequestApproverUser{
				{User: &gitlab.BasicUser{Name: "Ada L", Username: "ada"}},
			},
			reviewers: []*gitlab.MergeRequestReviewer{
				{User: &gitlab.BasicUser{Name: "Ada Lovelace", Username: "ada"}, State: "requested_changes"},
			},
			want: []types.PRApprovals{{Username: "Ada Lovelace (@ada)", State: types.ReviewChangesRequested}},
		},
		{
			name:       "a comment does not add a second entry beside an approval",
			approvedBy: approvedBy("ada"),
			reviewers:  []*gitlab.MergeRequestReviewer{reviewer("ada", "reviewed")},
			want:       []types.PRApprovals{{Username: "ADA (@ada)", State: types.ReviewApproved}},
		},
		{
			name: "states with no decision are left out",
			reviewers: []*gitlab.MergeRequestReviewer{
				reviewer("ada", "unreviewed"), reviewer("bob", "review_started"), reviewer("cy", "unapproved"),
			},
			want: []types.PRApprovals{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, gitlabReviews(tc.approvedBy, tc.reviewers))
		})
	}
}

// approvalsServer serves an MR's approvals and reviewers in GitLab's shapes.
func approvalsServer(t *testing.T, approvals, reviewers string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/merge_requests/3/approvals"):
			_, _ = fmt.Fprint(w, approvals)
		case strings.HasSuffix(r.URL.Path, "/merge_requests/3/reviewers"):
			_, _ = fmt.Fprint(w, reviewers)
		case strings.HasSuffix(r.URL.Path, "/merge_requests/3/commits"):
			_, _ = fmt.Fprint(w, "[]")
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ts.Close)
	return ts
}

func TestGetMergeRequestReviews_ReadsApprovalsAndReviewers(t *testing.T) {
	ts := approvalsServer(t,
		`{"approved_by":[{"user":{"name":"Ada","username":"ada"}}]}`,
		`[{"user":{"name":"Bob","username":"bob"},"state":"requested_changes","created_at":"2026-03-01T12:00:00Z"}]`,
	)

	approvers, reviews, err := newPaginationConfig(ts.URL).GetMergeRequestReviews(3)
	require.NoError(t, err)
	require.Equal(t, []any{types.PRApprovals{Username: "Ada (@ada)"}}, approvers)
	require.Equal(t, []types.PRApprovals{
		{Username: "Ada (@ada)", State: types.ReviewApproved},
		{Username: "Bob (@bob)", State: types.ReviewChangesRequested},
	}, reviews)
}

func TestNewPRGitlabEvidenceV2_RecordsReviews(t *testing.T) {
	ts := approvalsServer(t,
		`{"approved_by":[]}`,
		`[{"user":{"name":"Bob","username":"bob"},"state":"requested_changes","created_at":"2026-03-01T12:00:00Z"}]`,
	)
	now := time.Now()
	mr := &gitlab.BasicMergeRequest{IID: 3, Author: &gitlab.BasicUser{}, CreatedAt: &now, MergedAt: &now}

	evidence, err := newPaginationConfig(ts.URL).newPRGitlabEvidenceV2(mr)
	require.NoError(t, err)
	require.Equal(t, []types.PRApprovals{{Username: "Bob (@bob)", State: types.ReviewChangesRequested}}, evidence.Reviews)
}

// GitLab versions without the reviewers endpoint answer 404. Reviews are then
// not reported at all, rather than reported without the change requests.
func TestGetMergeRequestReviews_OmitsReviewsWhenReviewersEndpointIsMissing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/merge_requests/3/approvals") {
			_, _ = fmt.Fprint(w, `{"approved_by":[{"user":{"name":"Ada","username":"ada"}}]}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message":"404 Not Found"}`)
	}))
	t.Cleanup(ts.Close)

	approvers, reviews, err := newPaginationConfig(ts.URL).GetMergeRequestReviews(3)
	require.NoError(t, err)
	require.Equal(t, []any{types.PRApprovals{Username: "Ada (@ada)"}}, approvers)
	require.Nil(t, reviews)
}
