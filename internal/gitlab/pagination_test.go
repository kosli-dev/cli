package gitlab

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// mrCommitsServer serves paged MR commits and 404s every signature lookup,
// which resolveGitlabSignature treats as an unsigned commit.
type mrCommitsServer struct {
	*httptest.Server
	queries []string
}

// newMRCommitsServer replies with one page of commit SHAs per entry in pages,
// setting X-Next-Page the way GitLab does. nextPageOverride, when non-empty,
// is sent as X-Next-Page on every page to simulate a misbehaving server.
func newMRCommitsServer(t *testing.T, nextPageOverride string, pages ...[]string) *mrCommitsServer {
	t.Helper()
	s := &mrCommitsServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/signature") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"message":"404 Not Found"}`)
			return
		}
		s.queries = append(s.queries, r.URL.RawQuery)
		i := len(s.queries) - 1
		if nextPageOverride != "" {
			w.Header().Set("X-Next-Page", nextPageOverride)
			_, _ = fmt.Fprint(w, commitsJSON(pages[0]))
			return
		}
		if i >= len(pages) {
			t.Errorf("unexpected request %d", len(s.queries))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if i < len(pages)-1 {
			w.Header().Set("X-Next-Page", strconv.Itoa(i+2))
		}
		_, _ = fmt.Fprint(w, commitsJSON(pages[i]))
	}))
	t.Cleanup(s.Close)
	return s
}

func commitsJSON(shas []string) string {
	items := []string{}
	for _, sha := range shas {
		items = append(items, fmt.Sprintf(`{"id":%q,"message":"msg","author_name":"Ada",`+
			`"author_email":"ada@example.com","created_at":"2026-03-01T12:00:00Z",`+
			`"authored_date":"2026-03-01T12:00:00Z","web_url":"https://gitlab.com/o/r/-/commit/%s"}`, sha, sha))
	}
	return "[" + strings.Join(items, ",") + "]"
}

func newPaginationConfig(url string) *GitlabConfig {
	return &GitlabConfig{Token: "fake-token", BaseURL: url, Org: "test-org", Repository: "test-repo"}
}

func mrWithIID(iid int64) *gitlab.BasicMergeRequest {
	return &gitlab.BasicMergeRequest{IID: iid, SourceBranch: "feature"}
}

func TestGetMergeRequestCommits_FollowsNextPage(t *testing.T) {
	ts := newMRCommitsServer(t, "", []string{"sha1", "sha2"}, []string{"sha3"})
	commits, err := newPaginationConfig(ts.URL).GetMergeRequestCommits(mrWithIID(1))
	require.NoError(t, err)

	shas := []string{}
	for _, c := range commits {
		shas = append(shas, c.SHA)
	}
	require.Equal(t, []string{"sha1", "sha2", "sha3"}, shas)
	require.Len(t, ts.queries, 2)
	require.Contains(t, ts.queries[0], "per_page=100")
	require.Contains(t, ts.queries[1], "page=2")
	require.Nil(t, commits[0].Verified, "a 404 signature means unsigned, not an error")
}

func TestGetMergeRequestCommits_SingleRequestWhenNoNextPage(t *testing.T) {
	ts := newMRCommitsServer(t, "", []string{"sha1"})

	commits, err := newPaginationConfig(ts.URL).GetMergeRequestCommits(mrWithIID(1))
	require.NoError(t, err)
	require.Len(t, commits, 1)
	require.Len(t, ts.queries, 1)
}

func TestGetMergeRequestCommits_ErrorsWhenNextPageDoesNotAdvance(t *testing.T) {
	ts := newMRCommitsServer(t, "1", []string{"sha1"})

	_, err := newPaginationConfig(ts.URL).GetMergeRequestCommits(mrWithIID(1))
	require.ErrorContains(t, err, "did not advance")
}

func TestListMergeRequestCommits_ErrorsPastMaxPages(t *testing.T) {
	// A server that always advances would page forever without the cap.
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Next-Page", strconv.Itoa(calls+1))
		_, _ = fmt.Fprint(w, commitsJSON([]string{"sha"}))
	}))
	defer ts.Close()

	config := newPaginationConfig(ts.URL)
	client, err := config.NewGitlabClientFromToken()
	require.NoError(t, err)

	_, err = config.listMergeRequestCommits(client, 1, 2)
	require.ErrorContains(t, err, "2 pages")
	require.Equal(t, 2, calls)
}

// mrListServer serves paged MRs-for-commit, setting X-Next-Page as GitLab does.
// alwaysAdvance keeps advertising a next page, so only the cap stops it.
func mrListServer(t *testing.T, alwaysAdvance bool, pages ...[]int64) (*httptest.Server, *int) {
	t.Helper()
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if alwaysAdvance {
			w.Header().Set("X-Next-Page", strconv.Itoa(calls+1))
			_, _ = fmt.Fprint(w, mrsJSON(pages[0]))
			return
		}
		i := calls - 1
		if i >= len(pages) {
			t.Errorf("unexpected request %d", calls)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if i < len(pages)-1 {
			w.Header().Set("X-Next-Page", strconv.Itoa(i+2))
		}
		_, _ = fmt.Fprint(w, mrsJSON(pages[i]))
	}))
	t.Cleanup(ts.Close)
	return ts, &calls
}

func mrsJSON(iids []int64) string {
	items := []string{}
	for _, iid := range iids {
		items = append(items, fmt.Sprintf(`{"iid":%d,"source_branch":"feature","web_url":"https://gitlab.com/o/r/-/merge_requests/%d"}`, iid, iid))
	}
	return "[" + strings.Join(items, ",") + "]"
}

func iidsOf(mrs []*gitlab.BasicMergeRequest) []int64 {
	iids := []int64{}
	for _, mr := range mrs {
		iids = append(iids, mr.IID)
	}
	return iids
}

func TestMergeRequestsForCommit_FollowsNextPage(t *testing.T) {
	ts, calls := mrListServer(t, false, []int64{1, 2}, []int64{3})

	mrs, err := newPaginationConfig(ts.URL).MergeRequestsForCommit("abc123")
	require.NoError(t, err)
	require.Equal(t, []int64{1, 2, 3}, iidsOf(mrs))
	require.Equal(t, 2, *calls)
}

func TestMergeRequestsForCommit_SingleRequestWhenNoNextPage(t *testing.T) {
	ts, calls := mrListServer(t, false, []int64{1})

	mrs, err := newPaginationConfig(ts.URL).MergeRequestsForCommit("abc123")
	require.NoError(t, err)
	require.Len(t, mrs, 1)
	require.Equal(t, 1, *calls)
}

func TestListMergeRequestsForCommit_ErrorsPastMaxPages(t *testing.T) {
	ts, calls := mrListServer(t, true, []int64{1})
	config := newPaginationConfig(ts.URL)
	client, err := config.NewGitlabClientFromToken()
	require.NoError(t, err)

	got, err := config.listMergeRequestsForCommit(client, "abc123", 2)
	require.ErrorContains(t, err, "2 pages")
	require.Nil(t, got)
	require.Equal(t, 2, *calls)
}

func TestListMergeRequestsForCommit_SendsPageParams(t *testing.T) {
	queries := []string{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, mrsJSON([]int64{1}))
	}))
	defer ts.Close()

	_, err := newPaginationConfig(ts.URL).MergeRequestsForCommit("abc123")
	require.NoError(t, err)
	require.Len(t, queries, 1)
	require.Contains(t, queries[0], "per_page=100")
	require.Contains(t, queries[0], "page=1")
}

func TestListMergeRequestsForCommit_ErrorsWhenNextPageDoesNotAdvance(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Next-Page", "1")
		_, _ = fmt.Fprint(w, mrsJSON([]int64{1}))
	}))
	defer ts.Close()

	_, err := newPaginationConfig(ts.URL).MergeRequestsForCommit("abc123")
	require.ErrorContains(t, err, "did not advance")
	// Unwrap, not ErrorContains: the message survives %v too, so only the
	// chain distinguishes a wrapped cause from an interpolated one.
	require.ErrorContains(t, errors.Unwrap(err), "did not advance", "the cause must stay unwrappable")
	// page starts at 1, so "next is 1" is already non-advancing: it fails on
	// the first response rather than burning maxPages round trips.
	require.Equal(t, 1, calls)
}
