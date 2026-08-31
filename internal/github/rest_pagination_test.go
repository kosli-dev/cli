package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// restTestServer serves canned REST pages and records the query string of each
// request, so tests can assert per_page and page were sent.
type restTestServer struct {
	*httptest.Server
	queries []string
}

// newRestTestServer replies to path with pages in order, setting the Link
// header that go-github reads to derive NextPage.
func newRestTestServer(t *testing.T, path string, pages ...string) *restTestServer {
	t.Helper()
	s := &restTestServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		s.queries = append(s.queries, r.URL.RawQuery)
		i := len(s.queries) - 1
		if i >= len(pages) {
			t.Errorf("unexpected request %d", len(s.queries))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if i < len(pages)-1 {
			w.Header().Set("Link", fmt.Sprintf(`<%s%s?page=%d>; rel="next"`, s.URL, path, i+2))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, pages[i])
	}))
	t.Cleanup(s.Close)
	return s
}

func newRestConfig(url string) *GithubConfig {
	return &GithubConfig{Token: "fake-token", BaseURL: url, Org: "test-org", Repository: "test-repo"}
}

func TestPullRequestsForCommit_FollowsLinkHeader(t *testing.T) {
	ts := newRestTestServer(t, "/api/v3/repos/test-org/test-repo/commits/abc123/pulls",
		`[{"number":1},{"number":2}]`,
		`[{"number":3}]`,
	)

	prs, err := newRestConfig(ts.URL).PullRequestsForCommit("abc123")
	require.NoError(t, err)
	require.Len(t, prs, 3)
	require.Equal(t, 3, prs[2].GetNumber())
	require.Contains(t, ts.queries[0], "per_page=100")
	require.Contains(t, ts.queries[1], "page=2")
}

func TestGetPullRequestApprovers_FindsApproverOnSecondPage(t *testing.T) {
	// The APPROVED filter runs after fetching, so an approval on page 2 was
	// dropped entirely before pagination (#1082).
	ts := newRestTestServer(t, "/api/v3/repos/test-org/test-repo/pulls/5/reviews",
		`[{"state":"COMMENTED","user":{"login":"bob"}}]`,
		`[{"state":"APPROVED","user":{"login":"ada"}}]`,
	)

	approvers, err := newRestConfig(ts.URL).GetPullRequestApprovers(5)
	require.NoError(t, err)
	require.Equal(t, []string{"ada"}, approvers)
	require.Contains(t, ts.queries[0], "per_page=100")
}

func TestPullRequestsForCommit_ReturnsErrorFromSecondPage(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Link", `<http://example.invalid/next?page=2>; rel="next"`)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `[{"number":1}]`)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := newRestConfig(ts.URL).PullRequestsForCommit("abc123")
	require.Error(t, err)
}

func TestGetPullRequestApprovers_DeduplicatesRepeatApprovals(t *testing.T) {
	// A reviewer can approve, request changes, then approve again. These
	// entries carry no timestamp, so the repeat is noise.
	ts := newRestTestServer(t, "/api/v3/repos/test-org/test-repo/pulls/5/reviews",
		`[{"state":"APPROVED","user":{"login":"ada"}},{"state":"CHANGES_REQUESTED","user":{"login":"ada"}}]`,
		`[{"state":"APPROVED","user":{"login":"ada"}},{"state":"APPROVED","user":{"login":"grace"}}]`,
	)

	approvers, err := newRestConfig(ts.URL).GetPullRequestApprovers(5)
	require.NoError(t, err)
	require.Equal(t, []string{"ada", "grace"}, approvers)
}

func TestGetPullRequestApprovers_ErrorsWhenNextPageDoesNotAdvance(t *testing.T) {
	// Without the guard a server pinning NextPage re-reads the same page until
	// the cap fires, accumulating duplicate work rather than failing.
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Link", `<https://example.invalid/x?page=1>; rel="next"`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"state":"APPROVED","user":{"login":"ada"}}]`)
	}))
	defer ts.Close()

	_, err := newRestConfig(ts.URL).GetPullRequestApprovers(5)
	require.ErrorContains(t, err, "did not advance")
	require.Equal(t, 1, calls)
}

// alwaysMorePages advertises a next page forever, so only the cap stops it.
func alwaysMorePages(t *testing.T, body string) (*httptest.Server, *int) {
	t.Helper()
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Link", fmt.Sprintf(`<https://example.invalid/x?page=%d>; rel="next"`, calls+1))
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(ts.Close)
	return ts, &calls
}

func TestListPullRequestsForCommit_ErrorsPastMaxPages(t *testing.T) {
	ts, calls := alwaysMorePages(t, `[{"number":1}]`)
	config := newRestConfig(ts.URL)
	client, err := NewGithubClientFromToken(context.Background(), config.Token, config.BaseURL, false)
	require.NoError(t, err)

	got, err := config.listPullRequestsForCommit(context.Background(), client, "abc123", 2)
	require.ErrorContains(t, err, "2 pages")
	require.Nil(t, got, "no partial result on an error path")
	require.Equal(t, 2, *calls)
}

func TestListApprovers_ErrorsPastMaxPages(t *testing.T) {
	ts, calls := alwaysMorePages(t, `[{"state":"APPROVED","user":{"login":"ada"}}]`)
	config := newRestConfig(ts.URL)
	client, err := NewGithubClientFromToken(context.Background(), config.Token, config.BaseURL, false)
	require.NoError(t, err)

	got, err := config.listApprovers(context.Background(), client, 5, 2)
	require.ErrorContains(t, err, "2 pages")
	require.Nil(t, got, "no partial result on an error path")
	require.Equal(t, 2, *calls)
}

// Mirror of TestGetPullRequestApprovers_ErrorsWhenNextPageDoesNotAdvance for
// the identical guard in the PRs-for-commit loop.
func TestPullRequestsForCommit_ErrorsWhenNextPageDoesNotAdvance(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Link", `<https://example.invalid/x?page=1>; rel="next"`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"number":1}]`)
	}))
	defer ts.Close()

	_, err := newRestConfig(ts.URL).PullRequestsForCommit("abc123")
	require.ErrorContains(t, err, "did not advance")
	require.Equal(t, 1, calls)
}
