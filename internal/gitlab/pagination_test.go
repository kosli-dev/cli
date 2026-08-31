package gitlab

import (
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
