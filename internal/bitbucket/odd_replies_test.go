package bitbucket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// Baseline replies for every call PREvidenceForCommitV2 makes. Shapes follow
// https://api.bitbucket.org/swagger.json: paginated_pullrequests (GET
// /repositories/{workspace}/{repo_slug}/commit/{commit}/pullrequests),
// pullrequest and participant (GET .../pullrequests/{pull_request_id}) and
// paginated_changeset of base_commit (GET .../pullrequests/{pull_request_id}/commits),
// as documented at
// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/.
// {{base}} stands for https://api.bitbucket.org and is replaced with the test
// server's address when served, so links the code follows come back to it.
const (
	oddCommit = "4f1c2a9e8d7b6c5a4f3e2d1c0b9a8f7e6d5c4b3a"

	oddCommitPullRequests = `{
	"pagelen": 50,
	"page": 1,
	"size": 1,
	"values": [{
		"type": "pullrequest",
		"id": 7,
		"title": "Add retry to uploads",
		"state": "MERGED",
		"author": {"type": "user", "display_name": "Ada Lovelace", "uuid": "{a1b2c3d4-0000-4000-8000-000000000001}", "account_id": "557058:ada", "nickname": "ada"},
		"source": {"branch": {"name": "feature/retry"}, "commit": {"type": "commit", "hash": "9a8b7c6d5e4f"}, "repository": {"type": "repository", "full_name": "acme/widgets", "name": "widgets", "uuid": "{a1b2c3d4-0000-4000-8000-0000000000aa}"}},
		"destination": {"branch": {"name": "main"}, "commit": {"type": "commit", "hash": "1a2b3c4d5e6f"}, "repository": {"type": "repository", "full_name": "acme/widgets", "name": "widgets", "uuid": "{a1b2c3d4-0000-4000-8000-0000000000aa}"}},
		"links": {
			"self": {"href": "{{base}}/2.0/repositories/acme/widgets/pullrequests/7"},
			"html": {"href": "https://bitbucket.org/acme/widgets/pull-requests/7"}
		}
	}]
}`

	oddPullRequest = `{
	"type": "pullrequest",
	"id": 7,
	"title": "Add retry to uploads",
	"description": "Retries uploads on a 503.",
	"state": "MERGED",
	"draft": false,
	"queued": false,
	"reason": "",
	"close_source_branch": true,
	"comment_count": 2,
	"task_count": 0,
	"created_on": "2026-03-01T12:00:00.000000+00:00",
	"updated_on": "2026-03-01T16:30:00.654321+00:00",
	"author": {"type": "user", "display_name": "Ada Lovelace", "uuid": "{a1b2c3d4-0000-4000-8000-000000000001}", "account_id": "557058:ada", "nickname": "ada"},
	"closed_by": {"type": "user", "display_name": "Ada Lovelace", "uuid": "{a1b2c3d4-0000-4000-8000-000000000001}", "account_id": "557058:ada", "nickname": "ada"},
	"merge_commit": {"type": "commit", "hash": "4f1c2a9e8d7b"},
	"source": {"branch": {"name": "feature/retry"}, "commit": {"type": "commit", "hash": "9a8b7c6d5e4f"}, "repository": {"type": "repository", "full_name": "acme/widgets", "name": "widgets", "uuid": "{a1b2c3d4-0000-4000-8000-0000000000aa}"}},
	"destination": {"branch": {"name": "main"}, "commit": {"type": "commit", "hash": "1a2b3c4d5e6f"}, "repository": {"type": "repository", "full_name": "acme/widgets", "name": "widgets", "uuid": "{a1b2c3d4-0000-4000-8000-0000000000aa}"}},
	"reviewers": [
		{"type": "user", "display_name": "Grace Hopper", "uuid": "{a1b2c3d4-0000-4000-8000-000000000002}", "account_id": "557058:grace", "nickname": "grace"},
		{"type": "user", "display_name": "Alan Turing", "uuid": "{a1b2c3d4-0000-4000-8000-000000000003}", "account_id": "557058:alan", "nickname": "alan"}
	],
	"participants": [
		{"type": "participant", "user": {"type": "user", "display_name": "Grace Hopper", "uuid": "{a1b2c3d4-0000-4000-8000-000000000002}", "account_id": "557058:grace", "nickname": "grace"}, "role": "REVIEWER", "approved": true, "state": "approved", "participated_on": "2026-03-01T13:00:00.123456+00:00"},
		{"type": "participant", "user": {"type": "user", "display_name": "Alan Turing", "uuid": "{a1b2c3d4-0000-4000-8000-000000000003}", "account_id": "557058:alan", "nickname": "alan"}, "role": "REVIEWER", "approved": false, "state": "changes_requested", "participated_on": "2026-03-01T14:00:00.000000+00:00"}
	],
	"links": {
		"self": {"href": "{{base}}/2.0/repositories/acme/widgets/pullrequests/7"},
		"html": {"href": "https://bitbucket.org/acme/widgets/pull-requests/7"},
		"commits": {"href": "{{base}}/2.0/repositories/acme/widgets/pullrequests/7/commits"}
	}
}`

	// Commits come one per page so the code's paging is exercised.
	oddPullRequestCommitsPage1 = `{
	"pagelen": 1,
	"page": 1,
	"values": [{
		"type": "commit",
		"hash": "9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b",
		"date": "2026-03-01T11:00:00+00:00",
		"author": {"type": "author", "raw": "Ada Lovelace <ada@example.com>", "user": {"type": "user", "display_name": "Ada Lovelace", "uuid": "{a1b2c3d4-0000-4000-8000-000000000001}"}},
		"message": "Add retry to uploads\n",
		"parents": [{"type": "commit", "hash": "8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c"}],
		"links": {
			"self": {"href": "{{base}}/2.0/repositories/acme/widgets/commit/9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b"},
			"html": {"href": "https://bitbucket.org/acme/widgets/commits/9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b"}
		}
	}],
	"next": "{{base}}/2.0/repositories/acme/widgets/pullrequests/7/commits?page=2"
}`

	oddPullRequestCommitsPage2 = `{
	"pagelen": 1,
	"page": 2,
	"values": [{
		"type": "commit",
		"hash": "8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c",
		"date": "2026-03-01T10:00:00+00:00",
		"author": {"type": "author", "raw": "Ada Lovelace <ada@example.com>", "user": {"type": "user", "display_name": "Ada Lovelace", "uuid": "{a1b2c3d4-0000-4000-8000-000000000001}"}},
		"message": "Add upload test\n",
		"parents": [{"type": "commit", "hash": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"}],
		"links": {
			"self": {"href": "{{base}}/2.0/repositories/acme/widgets/commit/8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c"},
			"html": {"href": "https://bitbucket.org/acme/widgets/commits/8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c"}
		}
	}],
	"previous": "{{base}}/2.0/repositories/acme/widgets/pullrequests/7/commits?page=1"
}`
)

// oddRoutes maps each reply name to the request that gets it.
var oddRoutes = map[string]string{
	"commit_pullrequests":   "/2.0/repositories/acme/widgets/commit/" + oddCommit + "/pullrequests",
	"pullrequest":           "/2.0/repositories/acme/widgets/pullrequests/7",
	"pullrequest_commits_1": "/2.0/repositories/acme/widgets/pullrequests/7/commits",
	"pullrequest_commits_2": "/2.0/repositories/acme/widgets/pullrequests/7/commits?page=2",
}

func oddBaselineReplies() map[string]string {
	return map[string]string{
		"commit_pullrequests":   oddCommitPullRequests,
		"pullrequest":           oddPullRequest,
		"pullrequest_commits_1": oddPullRequestCommitsPage1,
		"pullrequest_commits_2": oddPullRequestCommitsPage2,
	}
}

// redirectTransport sends requests for api.bitbucket.org to the test server
// and refuses any other host, so the test never reaches the network.
type redirectTransport struct {
	server *url.URL
}

func (rt redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Host {
	case "api.bitbucket.org":
		req = req.Clone(req.Context())
		req.URL.Scheme = rt.server.Scheme
		req.URL.Host = rt.server.Host
		req.Host = rt.server.Host
	case rt.server.Host:
	default:
		return nil, fmt.Errorf("test refuses request to %s", req.URL)
	}
	return http.DefaultTransport.RoundTrip(req)
}

// runOddReplies serves replies and collects evidence for oddCommit through
// the given entry point: 1 for PREvidenceForCommitV1, 2 for V2.
func runOddReplies(t *testing.T, version int, replies map[string]string) ([]*types.PREvidence, error) {
	t.Helper()
	var base string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for name, route := range oddRoutes {
			if r.URL.RequestURI() == route {
				_, _ = w.Write([]byte(strings.ReplaceAll(replies[name], "{{base}}", base)))
				return
			}
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	base = server.URL
	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	c := &Config{
		Workspace:   "acme",
		Repository:  "widgets",
		AccessToken: "test-token",
		Logger:      logger.NewStandardLogger(),
		KosliClient: &requests.Client{
			Logger:     logger.NewStandardLogger(),
			HttpClient: &http.Client{Transport: redirectTransport{server: serverURL}},
		},
	}
	if version == 1 {
		return c.PREvidenceForCommitV1(oddCommit)
	}
	return c.PREvidenceForCommitV2(oddCommit)
}

func TestOddReplies_BaselineGivesFullEvidenceV1(t *testing.T) {
	evidence, err := runOddReplies(t, 1, oddBaselineReplies())
	require.NoError(t, err)
	require.Equal(t, []*types.PREvidence{{
		MergeCommit: oddCommit,
		URL:         "https://bitbucket.org/acme/widgets/pull-requests/7",
		State:       "MERGED",
		Approvers:   []any{"Grace Hopper"},
	}}, evidence)
}

func TestOddReplies_BaselineGivesFullEvidenceV2(t *testing.T) {
	evidence, err := runOddReplies(t, 2, oddBaselineReplies())
	require.NoError(t, err)
	require.Equal(t, []*types.PREvidence{{
		MergeCommit: oddCommit,
		URL:         "https://bitbucket.org/acme/widgets/pull-requests/7",
		State:       "MERGED",
		Approvers: []any{types.PRApprovals{
			Username: "Grace Hopper", State: "approved", Timestamp: 1772370000,
		}},
		Author:    "Ada Lovelace",
		CreatedAt: 1772366400,
		MergedAt:  1772382600,
		Title:     "Add retry to uploads",
		HeadRef:   "feature/retry",
		BaseRef:   "main",
		Commits: []types.Commit{
			{
				SHA:       "9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b",
				Message:   "Add retry to uploads\n",
				Author:    "Ada Lovelace <ada@example.com>",
				Timestamp: 1772362800,
				URL:       "https://bitbucket.org/acme/widgets/commits/9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b",
			},
			{
				SHA:       "8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c",
				Message:   "Add upload test\n",
				Author:    "Ada Lovelace <ada@example.com>",
				Timestamp: 1772359200,
				URL:       "https://bitbucket.org/acme/widgets/commits/8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c",
			},
		},
		Reviews: []types.PRApprovals{
			{Username: "Grace Hopper", State: types.ReviewApproved, Timestamp: 1772370000},
			{Username: "Alan Turing", State: types.ReviewChangesRequested, Timestamp: 1772373600},
		},
	}}, evidence)
}

// oddAllowedErrors are the values the code needs for its next request, so
// making them (or anything holding them) null may fail. Keyed by entry point,
// reply and path.
var oddAllowedErrors = map[string]string{
	"V1 commit_pullrequests values[0].links.self.href": "the pull request's API URL, fetched next",
	"V2 commit_pullrequests values[0].links.self.href": "the pull request's API URL, fetched next",
	"V2 pullrequest id": "the pull request number, used to fetch its commits",
}

// oddReadLists are the lists the code reads, per reply. A null item in one of
// them is an entry that can't be read, so it must fail.
var oddReadLists = map[string]string{
	"commit_pullrequests":   "values",
	"pullrequest":           "participants",
	"pullrequest_commits_1": "values",
	"pullrequest_commits_2": "values",
}

// oddNamelessApprovals make the approval in participants[0] one with no
// identity, as for a deleted account. It must not count as an approval, but
// stays in the reviews.
var oddNamelessApprovals = map[string]bool{
	"pullrequest participants[0].user":              true,
	"pullrequest participants[0].user.display_name": true,
}

// oddMayError says whether path in reply holds, or is, a value in
// oddAllowedErrors for that entry point.
func oddMayError(entry, reply, path string) bool {
	for key := range oddAllowedErrors {
		allowed, ok := strings.CutPrefix(key, entry+" "+reply+" ")
		if !ok {
			continue
		}
		if path == "." || allowed == path || strings.HasPrefix(allowed, path+".") || strings.HasPrefix(allowed, path+"[") {
			return true
		}
	}
	return false
}

// TestOddReplies makes each value in each reply null in turn, and empties each
// list, and checks the rule of kosli-dev/cli#1283 through both entry points:
// never a panic; a null field is no error unless oddAllowedErrors lists it; a
// null item in a list the code reads is an error; an emptied list is no error.
func TestOddReplies(t *testing.T) {
	for _, version := range []int{1, 2} {
		entry := fmt.Sprintf("V%d", version)
		baseline, err := runOddReplies(t, version, oddBaselineReplies())
		require.NoError(t, err)

		for name := range oddRoutes {
			if version == 1 && strings.HasPrefix(name, "pullrequest_commits") {
				continue // V1 doesn't fetch commits.
			}
			variants, err := testHelpers.NullVariants([]byte(oddBaselineReplies()[name]))
			require.NoError(t, err, name)
			for _, v := range variants {
				t.Run(entry+" "+name+" "+v.Path, func(t *testing.T) {
					replies := oddBaselineReplies()
					replies[name] = string(v.Body)
					outcome := testHelpers.EvidenceOutcome(baseline, func() ([]*types.PREvidence, error) {
						return runOddReplies(t, version, replies)
					})
					require.False(t, strings.HasPrefix(outcome, "panic"), outcome)
					switch {
					case strings.HasSuffix(v.Path, "[]"):
						require.True(t, strings.HasPrefix(outcome, "ok"), "an emptied list must not fail: %s", outcome)
					case regexp.MustCompile(`^` + oddReadLists[name] + `\[\d+\]$`).MatchString(v.Path):
						require.Equal(t, "error", outcome, "a null list item must fail")
					case oddNamelessApprovals[name+" "+v.Path]:
						evidence, err := runOddReplies(t, version, replies)
						require.NoError(t, err)
						require.Empty(t, evidence[0].Approvers)
						if version == 2 {
							require.Contains(t, evidence[0].Reviews, types.PRApprovals{State: types.ReviewApproved, Timestamp: 1772370000})
						}
					case oddMayError(entry, name, v.Path):
					default:
						require.True(t, strings.HasPrefix(outcome, "ok"), "a null field must not fail: %s", outcome)
					}
				})
			}
		}
	}
}

// Commit pages whose next links cycle would be fetched forever: a page
// pointing at itself, or two pages pointing at each other.
func TestOddReplies_CommitPagesThatCycleFail(t *testing.T) {
	for name, next := range map[string]string{
		"page 2 to itself": "{{base}}/2.0/repositories/acme/widgets/pullrequests/7/commits?page=2",
		"page 2 to page 1": "{{base}}/2.0/repositories/acme/widgets/pullrequests/7/commits",
	} {
		t.Run(name, func(t *testing.T) {
			replies := oddBaselineReplies()
			replies["pullrequest_commits_2"] = `{"pagelen": 1, "page": 2, "values": [], "next": "` + next + `"}`
			done := make(chan error, 1)
			go func() {
				_, err := runOddReplies(t, 2, replies)
				done <- err
			}()
			select {
			case err := <-done:
				require.ErrorContains(t, err, fmt.Sprintf("aborting after %d pages", defaultMaxPages))
			case <-time.After(20 * time.Second):
				t.Fatal("paging did not stop")
			}
		})
	}
}
