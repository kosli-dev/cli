package github

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// graphQLTestServer serves canned GraphQL responses in order and records every
// request body, so tests can assert which connection each round trip queried.
type graphQLTestServer struct {
	*httptest.Server
	bodies []string
}

// newGraphQLTestServer replies with responses in order. An entry of "500"
// makes that request fail, to exercise the retry path.
func newGraphQLTestServer(t *testing.T, responses ...string) *graphQLTestServer {
	t.Helper()
	s := &graphQLTestServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graphql" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		s.bodies = append(s.bodies, string(body))
		if len(s.bodies) > len(responses) {
			t.Errorf("unexpected request %d: %s", len(s.bodies), body)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		reply := responses[len(s.bodies)-1]
		if reply == "500" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, reply)
	}))
	t.Cleanup(s.Close)
	return s
}

// newPaginationConfig points a config at ts with retries that do not sleep.
func newPaginationConfig(ts *graphQLTestServer) *GithubConfig {
	return &GithubConfig{
		Token:      "fake-token",
		BaseURL:    ts.URL,
		Org:        "test-org",
		Repository: "test-repo",
		Sleep:      func(time.Duration) {},
	}
}

// commitNodeJSON is one node of a GraphQL commits connection.
func commitNodeJSON(sha string) string {
	return fmt.Sprintf(`{"commit":{"oid":%q,"messageHeadline":"msg %s",`+
		`"committedDate":"2026-03-01T12:00:00Z","authoredDate":"2026-03-01T12:00:00Z",`+
		`"url":"https://github.com/o/r/commit/%s",`+
		`"author":{"name":"Ada","email":"ada@example.com","user":{"login":"ada"}},`+
		`"signature":null}}`, sha, sha, sha)
}

// reviewNodeJSON is one node of a GraphQL reviews connection.
func reviewNodeJSON(login string) string {
	return fmt.Sprintf(`{"author":{"login":%q},"state":"APPROVED","submittedAt":"2026-03-01T13:00:00Z"}`, login)
}

// connectionJSON wraps nodes with a pageInfo block. An empty cursor means the
// connection has no further pages.
func connectionJSON(nodes []string, nextCursor string) string {
	hasNext := nextCursor != ""
	return fmt.Sprintf(`{"nodes":[%s],"pageInfo":{"hasNextPage":%t,"endCursor":%q}}`,
		strings.Join(nodes, ","), hasNext, nextCursor)
}

// prJSON is a full pullRequest object with the given commit and review connections.
func prJSON(commits, reviews string) string {
	return fmt.Sprintf(`{"title":"A PR","state":"MERGED","headRefName":"feature","baseRefName":"main",`+
		`"url":"https://github.com/o/r/pull/1","createdAt":"2026-03-01T11:00:00Z",`+
		`"mergedAt":"2026-03-01T14:00:00Z","mergeCommit":{"oid":"merge-sha"},`+
		`"author":{"login":"ada"},"commits":%s,"reviews":%s}`, commits, reviews)
}

func byPRNumberResponse(commits, reviews string) string {
	return fmt.Sprintf(`{"data":{"repository":{"pullRequest":%s}}}`, prJSON(commits, reviews))
}

// commitsPageResponse is a follow-up page reply selecting only commits.
func commitsPageResponse(nodes []string, nextCursor string) string {
	return fmt.Sprintf(`{"data":{"repository":{"pullRequest":{"commits":%s}}}}`,
		connectionJSON(nodes, nextCursor))
}

// reviewsPageResponse is a follow-up page reply selecting only reviews.
func reviewsPageResponse(nodes []string, nextCursor string) string {
	return fmt.Sprintf(`{"data":{"repository":{"pullRequest":{"reviews":%s}}}}`,
		connectionJSON(nodes, nextCursor))
}

func shasOf(t *testing.T, config *GithubConfig, prNumber int) []string {
	t.Helper()
	evidence, err := config.PREvidenceByPRNumber(prNumber)
	require.NoError(t, err)
	require.NotNil(t, evidence)
	shas := []string{}
	for _, c := range evidence.Commits {
		shas = append(shas, c.SHA)
	}
	return shas
}

func TestPREvidenceByPRNumber_FollowsCommitPages(t *testing.T) {
	ts := newGraphQLTestServer(t,
		byPRNumberResponse(
			connectionJSON([]string{commitNodeJSON("sha1"), commitNodeJSON("sha2")}, "c1"),
			connectionJSON([]string{reviewNodeJSON("ada")}, ""),
		),
		commitsPageResponse([]string{commitNodeJSON("sha3")}, ""),
	)

	require.Equal(t, []string{"sha1", "sha2", "sha3"}, shasOf(t, newPaginationConfig(ts), 1))
	require.Len(t, ts.bodies, 2)
	require.Contains(t, ts.bodies[1], "c1", "follow-up must carry the cursor")
	require.NotContains(t, ts.bodies[1], "reviews(", "follow-up must not re-fetch reviews")
}

func TestPREvidenceByPRNumber_FollowsReviewPages(t *testing.T) {
	ts := newGraphQLTestServer(t,
		byPRNumberResponse(
			connectionJSON([]string{commitNodeJSON("sha1")}, ""),
			connectionJSON([]string{reviewNodeJSON("ada")}, "r1"),
		),
		reviewsPageResponse([]string{reviewNodeJSON("grace")}, ""),
	)

	evidence, err := newPaginationConfig(ts).PREvidenceByPRNumber(1)
	require.NoError(t, err)
	require.Len(t, evidence.Approvers, 2)
	require.Len(t, ts.bodies, 2)
	require.Contains(t, ts.bodies[1], "r1")
	require.NotContains(t, ts.bodies[1], "commits(", "follow-up must not re-fetch commits")
}

func TestPREvidenceByPRNumber_SingleRequestWhenNoMorePages(t *testing.T) {
	ts := newGraphQLTestServer(t, byPRNumberResponse(
		connectionJSON([]string{commitNodeJSON("sha1")}, ""),
		connectionJSON([]string{reviewNodeJSON("ada")}, ""),
	))

	require.Equal(t, []string{"sha1"}, shasOf(t, newPaginationConfig(ts), 1))
	require.Len(t, ts.bodies, 1)
}

func TestPREvidenceByPRNumber_RetriesFollowUpPage(t *testing.T) {
	ts := newGraphQLTestServer(t,
		byPRNumberResponse(
			connectionJSON([]string{commitNodeJSON("sha1")}, "c1"),
			connectionJSON(nil, ""),
		),
		"500",
		commitsPageResponse([]string{commitNodeJSON("sha2")}, ""),
	)

	require.Equal(t, []string{"sha1", "sha2"}, shasOf(t, newPaginationConfig(ts), 1))
	require.Len(t, ts.bodies, 3, "the failed follow-up page should have been retried")
}

func TestPREvidenceByPRNumber_ErrorsWhenCursorDoesNotAdvance(t *testing.T) {
	config := newPaginationConfig(newGraphQLTestServer(t,
		byPRNumberResponse(
			connectionJSON([]string{commitNodeJSON("sha1")}, "stuck"),
			connectionJSON(nil, ""),
		),
		commitsPageResponse([]string{commitNodeJSON("sha2")}, "stuck"),
	))

	_, err := config.PREvidenceByPRNumber(1)
	require.ErrorContains(t, err, "did not advance")
}

// v2PRNodeJSON is one associatedPullRequests node. Unlike the by-number query
// it carries the PR number and has no mergeCommit.
func v2PRNodeJSON(number int, commits, reviews string) string {
	return v2PRNodeInRepoJSON("test-org", "test-repo", number, commits, reviews)
}

// v2PRNodeInRepoJSON is an associatedPullRequests node belonging to owner/repo,
// which need not be the configured repository.
func v2PRNodeInRepoJSON(owner, repo string, number int, commits, reviews string) string {
	return fmt.Sprintf(`{"number":%d,"repository":{"name":%q,"owner":{"login":%q}},`+
		`"title":"A PR","state":"MERGED","headRefName":"feature",`+
		`"baseRefName":"main","url":"https://github.com/%s/%s/pull/%d",`+
		`"createdAt":"2026-03-01T11:00:00Z","mergedAt":"2026-03-01T14:00:00Z",`+
		`"author":{"login":"ada"},"commits":%s,"reviews":%s}`,
		number, repo, owner, owner, repo, number, commits, reviews)
}

// forCommitResponse carries no pageInfo for the PR connection: the query does
// not select one, and shurcooL's decoder rejects fields the struct lacks.
func forCommitResponse(prNodes ...string) string {
	return fmt.Sprintf(`{"data":{"repository":{"object":{"associatedPullRequests":{"nodes":[%s]}}}}}`,
		strings.Join(prNodes, ","))
}

func TestPREvidenceForCommitV2_FollowsCommitPagesPerPR(t *testing.T) {
	ts := newGraphQLTestServer(t,
		forCommitResponse(
			v2PRNodeJSON(7,
				connectionJSON([]string{commitNodeJSON("sha1")}, "c1"),
				connectionJSON(nil, "")),
			v2PRNodeJSON(9,
				connectionJSON([]string{commitNodeJSON("sha3")}, "c9"),
				connectionJSON(nil, "")),
		),
		commitsPageResponse([]string{commitNodeJSON("sha2")}, ""),
		commitsPageResponse([]string{commitNodeJSON("sha4")}, ""),
	)

	prs, err := newPaginationConfig(ts).PREvidenceForCommitV2("merge-sha")
	require.NoError(t, err)
	require.Len(t, prs, 2)
	require.Equal(t, []string{"sha1", "sha2"}, shasIn(prs[0]))
	require.Equal(t, []string{"sha3", "sha4"}, shasIn(prs[1]))
	require.Len(t, ts.bodies, 3)
	require.Contains(t, ts.bodies[1], `"prNumber":7`)
	require.Contains(t, ts.bodies[2], `"prNumber":9`)
}

func TestPREvidenceForCommitV2_RetriesFollowUpPage(t *testing.T) {
	ts := newGraphQLTestServer(t,
		forCommitResponse(v2PRNodeJSON(7,
			connectionJSON([]string{commitNodeJSON("sha1")}, "c1"),
			connectionJSON(nil, ""))),
		"500",
		commitsPageResponse([]string{commitNodeJSON("sha2")}, ""),
	)

	prs, err := newPaginationConfig(ts).PREvidenceForCommitV2("merge-sha")
	require.NoError(t, err)
	require.Equal(t, []string{"sha1", "sha2"}, shasIn(prs[0]))
	require.Len(t, ts.bodies, 3)
}

// The initial V2 query keeps its existing fail-fast behaviour; only the
// follow-up pages this change introduces are retried.
func TestPREvidenceForCommitV2_DoesNotRetryInitialQuery(t *testing.T) {
	ts := newGraphQLTestServer(t, "500")

	_, err := newPaginationConfig(ts).PREvidenceForCommitV2("merge-sha")
	require.Error(t, err)
	require.Len(t, ts.bodies, 1)
}

func shasIn(evidence *types.PREvidence) []string {
	shas := []string{}
	for _, c := range evidence.Commits {
		shas = append(shas, c.SHA)
	}
	return shas
}

// A commit's associated PRs are not guaranteed to live in the configured repo.
// The follow-up page query resolves pullRequest(number:) in a separate request,
// so it must target the repo the node came from — otherwise it silently reads a
// different PR's commits, or none.
func TestPREvidenceForCommitV2_FollowUpTargetsTheNodesOwnRepo(t *testing.T) {
	ts := newGraphQLTestServer(t,
		forCommitResponse(v2PRNodeInRepoJSON("upstream-org", "upstream-repo", 7,
			connectionJSON([]string{commitNodeJSON("sha1")}, "c1"),
			connectionJSON(nil, ""))),
		commitsPageResponse([]string{commitNodeJSON("sha2")}, ""),
	)

	prs, err := newPaginationConfig(ts).PREvidenceForCommitV2("merge-sha")
	require.NoError(t, err)
	require.Equal(t, []string{"sha1", "sha2"}, shasIn(prs[0]))
	require.Contains(t, ts.bodies[1], `"owner":"upstream-org"`)
	require.Contains(t, ts.bodies[1], `"repo":"upstream-repo"`)
}

// PREvidenceByPRNumber is given a number for the configured repo, so its
// follow-up pages must keep using it.
func TestPREvidenceByPRNumber_FollowUpUsesConfiguredRepo(t *testing.T) {
	ts := newGraphQLTestServer(t,
		byPRNumberResponse(
			connectionJSON([]string{commitNodeJSON("sha1")}, "c1"),
			connectionJSON(nil, "")),
		commitsPageResponse([]string{commitNodeJSON("sha2")}, ""),
	)

	_, err := newPaginationConfig(ts).PREvidenceByPRNumber(1)
	require.NoError(t, err)
	require.Contains(t, ts.bodies[1], `"owner":"test-org"`)
	require.Contains(t, ts.bodies[1], `"repo":"test-repo"`)
}

// A node's PR can live in a repo the token cannot read. That query fails
// permanently, so retrying it only delays the failure by the full ladder.
func TestPREvidenceForCommitV2_DoesNotRetryCrossRepoFollowUp(t *testing.T) {
	ts := newGraphQLTestServer(t,
		forCommitResponse(v2PRNodeInRepoJSON("upstream-org", "upstream-repo", 7,
			connectionJSON([]string{commitNodeJSON("sha1")}, "c1"),
			connectionJSON(nil, ""))),
		"500",
	)

	_, err := newPaginationConfig(ts).PREvidenceForCommitV2("merge-sha")
	require.Error(t, err)
	require.Len(t, ts.bodies, 2, "a cross-repo follow-up must not be retried")
}

// A stuck cursor is a hard failure now, so the message is the whole
// user-facing surface of the change — it has to say which PR gave up.
func TestPREvidenceForCommitV2_DrainErrorNamesThePullRequest(t *testing.T) {
	ts := newGraphQLTestServer(t,
		forCommitResponse(v2PRNodeInRepoJSON("upstream-org", "upstream-repo", 42,
			connectionJSON([]string{commitNodeJSON("sha1")}, "stuck"),
			connectionJSON(nil, ""))),
		commitsPageResponse([]string{commitNodeJSON("sha2")}, "stuck"),
	)

	_, err := newPaginationConfig(ts).PREvidenceForCommitV2("merge-sha")
	require.ErrorContains(t, err, "upstream-org/upstream-repo#42")
	require.ErrorContains(t, err, "commits")
	require.ErrorContains(t, err, "did not advance", "the cause must survive the wrap")
}

func TestPREvidenceByPRNumber_ApprovalDrainErrorNamesThePullRequest(t *testing.T) {
	ts := newGraphQLTestServer(t,
		byPRNumberResponse(
			connectionJSON([]string{commitNodeJSON("sha1")}, ""),
			connectionJSON([]string{reviewNodeJSON("ada")}, "stuck")),
		reviewsPageResponse([]string{reviewNodeJSON("grace")}, "stuck"),
	)

	_, err := newPaginationConfig(ts).PREvidenceByPRNumber(7)
	require.ErrorContains(t, err, "test-org/test-repo#7")
	require.ErrorContains(t, err, "approvals")
}
