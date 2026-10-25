package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// The odd-reply tests feed the PR evidence path replies that GitHub's schema
// allows but the baseline does not use: a nullable field made null, or a list
// emptied, with and without an errors entry naming it (v2.48.0 broke on a null
// signature.signer). They assert the rule in kosli-dev/cli#1283.

// Reply names: one per kind of query the evidence path sends.
const (
	replyByNumber   = "byNumber"
	replyForCommit  = "forCommit"
	replyCommitPage = "commitsPage"
	replyReviewPage = "reviewsPage"
	replySigners    = "signers"
)

// oddReplyQueryRoutes picks the reply for a request by its query text. Order
// matters: the main queries also contain "commits(first: 100".
var oddReplyQueryRoutes = []struct{ contains, reply string }{
	{"nodes(ids:", replySigners},
	{"associatedPullRequests", replyForCommit},
	{"$commitCursor", replyByNumber},
	{"commits(first: 100, after: $cursor)", replyCommitPage},
	{"reviews(first: 100", replyReviewPage},
}

// newOddReplyServer answers each GraphQL query with the reply its text routes
// to, however many times it is asked, so retries and extra pages never run
// the server out of replies.
func newOddReplyServer(t *testing.T, replies map[string]string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct{ Query string }
		_ = json.Unmarshal(body, &req)
		for _, route := range oddReplyQueryRoutes {
			if strings.Contains(req.Query, route.contains) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, replies[route.reply])
				return
			}
		}
		t.Errorf("no reply routed for query: %s", req.Query)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(ts.Close)
	return ts
}

const (
	// sha1 is signed by a user GitHub resolves; sha2 is unsigned.
	oddSignedCommit = `{"commit":{"id":"C_sha1","oid":"sha1","messageHeadline":"first",` +
		`"committedDate":"2026-03-01T12:00:00Z","authoredDate":"2026-03-01T11:30:00Z",` +
		`"url":"https://github.com/test-org/test-repo/commit/sha1",` +
		`"author":{"name":"Ada","email":"ada@example.com","user":{"login":"ada"}},` +
		`"signature":{"isValid":true,"state":"VALID","wasSignedByGitHub":false}}}`
	oddUnsignedCommit = `{"commit":{"id":"C_sha2","oid":"sha2","messageHeadline":"second",` +
		`"committedDate":"2026-03-01T12:30:00Z","authoredDate":"2026-03-01T12:15:00Z",` +
		`"url":"https://github.com/test-org/test-repo/commit/sha2",` +
		`"author":{"name":"Ada","email":"ada@example.com","user":{"login":"ada"}},` +
		`"signature":null}}`
	oddApproval = `{"author":{"login":"grace"},"state":"APPROVED",` +
		`"submittedAt":"2026-03-01T13:00:00Z","commit":{"oid":"sha2"}}`
	oddChangeRequest = `{"author":{"login":"linus"},"state":"CHANGES_REQUESTED",` +
		`"submittedAt":"2026-03-01T12:45:00Z","commit":{"oid":"sha1"}}`

	// The main queries return one commit and one review with a next page, so
	// both follow-up page queries run.
	oddPRFields = `"title":"A PR","state":"MERGED","headRefName":"feature","baseRefName":"main",` +
		`"url":"https://github.com/test-org/test-repo/pull/7","createdAt":"2026-03-01T11:00:00Z",` +
		`"mergedAt":"2026-03-01T14:00:00Z",`
	oddPRConnections = `"author":{"login":"ada"},"headRefOid":"sha2",` +
		`"commits":{"nodes":[` + oddSignedCommit + `],"pageInfo":{"hasNextPage":true,"endCursor":"c1"}},` +
		`"reviews":{"nodes":[` + oddApproval + `],"pageInfo":{"hasNextPage":true,"endCursor":"r1"}}`

	oddByNumberReply = `{"data":{"repository":{"pullRequest":{` + oddPRFields +
		`"mergeCommit":{"oid":"merge-sha"},` + oddPRConnections + `}}}}`
	oddForCommitReply = `{"data":{"repository":{"object":{"associatedPullRequests":{"nodes":[{"number":7,` +
		`"repository":{"name":"test-repo","owner":{"login":"test-org"}},` + oddPRFields + oddPRConnections + `}]}}}}}`
	oddCommitPageReply = `{"data":{"repository":{"pullRequest":{"commits":{"nodes":[` + oddUnsignedCommit +
		`],"pageInfo":{"hasNextPage":false,"endCursor":"c2"}}}}}}`
	oddReviewPageReply = `{"data":{"repository":{"pullRequest":{"reviews":{"nodes":[` + oddChangeRequest +
		`],"pageInfo":{"hasNextPage":false,"endCursor":"r2"}}}}}}`
	oddSignersReply = `{"data":{"nodes":[{"oid":"sha1","signature":{"signer":{"login":"ada"}}}]}}`
)

func oddBaselineReplies() map[string]string {
	return map[string]string{
		replyByNumber:   oddByNumberReply,
		replyForCommit:  oddForCommitReply,
		replyCommitPage: oddCommitPageReply,
		replyReviewPage: oddReviewPageReply,
		replySigners:    oddSignersReply,
	}
}

// oddNullable lists the reply paths whose GitHub schema type is nullable, for
// every field the evidence queries select. "[]" stands for a list item.
// Source: https://docs.github.com/public/fpt/schema.docs.graphql, checked
// 2026-10-21. GraphQL never returns null for a non-null (!) field, so only
// these are varied; an emptied list is legal for every list and always varied.
var oddNullable = map[string]bool{
	// Query.repository: Repository; Repository.pullRequest: PullRequest
	"data.repository":             true,
	"data.repository.pullRequest": true,
	// PullRequest.mergedAt: DateTime; mergeCommit: Commit; author: Actor
	"data.repository.pullRequest.mergedAt":    true,
	"data.repository.pullRequest.mergeCommit": true,
	"data.repository.pullRequest.author":      true,
	// PullRequestCommitConnection.nodes: [PullRequestCommit]; PageInfo.endCursor: String
	"data.repository.pullRequest.commits.nodes":              true,
	"data.repository.pullRequest.commits.nodes[]":            true,
	"data.repository.pullRequest.commits.pageInfo.endCursor": true,
	// Commit.author: GitActor; GitActor.name/email: String; user: User; Commit.signature: GitSignature
	"data.repository.pullRequest.commits.nodes[].commit.author":       true,
	"data.repository.pullRequest.commits.nodes[].commit.author.name":  true,
	"data.repository.pullRequest.commits.nodes[].commit.author.email": true,
	"data.repository.pullRequest.commits.nodes[].commit.author.user":  true,
	"data.repository.pullRequest.commits.nodes[].commit.signature":    true,
	// PullRequest.reviews: PullRequestReviewConnection; nodes: [PullRequestReview]
	"data.repository.pullRequest.reviews":                    true,
	"data.repository.pullRequest.reviews.nodes":              true,
	"data.repository.pullRequest.reviews.nodes[]":            true,
	"data.repository.pullRequest.reviews.pageInfo.endCursor": true,
	// PullRequestReview.author: Actor; submittedAt: DateTime; commit: Commit
	"data.repository.pullRequest.reviews.nodes[].author":      true,
	"data.repository.pullRequest.reviews.nodes[].submittedAt": true,
	"data.repository.pullRequest.reviews.nodes[].commit":      true,

	// Repository.object: GitObject; Commit.associatedPullRequests: PullRequestConnection;
	// PullRequestConnection.nodes: [PullRequest]
	"data.repository.object":                                                                    true,
	"data.repository.object.associatedPullRequests":                                             true,
	"data.repository.object.associatedPullRequests.nodes":                                       true,
	"data.repository.object.associatedPullRequests.nodes[]":                                     true,
	"data.repository.object.associatedPullRequests.nodes[].mergedAt":                            true,
	"data.repository.object.associatedPullRequests.nodes[].author":                              true,
	"data.repository.object.associatedPullRequests.nodes[].commits.nodes":                       true,
	"data.repository.object.associatedPullRequests.nodes[].commits.nodes[]":                     true,
	"data.repository.object.associatedPullRequests.nodes[].commits.pageInfo.endCursor":          true,
	"data.repository.object.associatedPullRequests.nodes[].commits.nodes[].commit.author":       true,
	"data.repository.object.associatedPullRequests.nodes[].commits.nodes[].commit.author.name":  true,
	"data.repository.object.associatedPullRequests.nodes[].commits.nodes[].commit.author.email": true,
	"data.repository.object.associatedPullRequests.nodes[].commits.nodes[].commit.author.user":  true,
	"data.repository.object.associatedPullRequests.nodes[].commits.nodes[].commit.signature":    true,
	"data.repository.object.associatedPullRequests.nodes[].reviews":                             true,
	"data.repository.object.associatedPullRequests.nodes[].reviews.nodes":                       true,
	"data.repository.object.associatedPullRequests.nodes[].reviews.nodes[]":                     true,
	"data.repository.object.associatedPullRequests.nodes[].reviews.pageInfo.endCursor":          true,
	"data.repository.object.associatedPullRequests.nodes[].reviews.nodes[].author":              true,
	"data.repository.object.associatedPullRequests.nodes[].reviews.nodes[].submittedAt":         true,
	"data.repository.object.associatedPullRequests.nodes[].reviews.nodes[].commit":              true,

	// Query.nodes: [Node]! (items nullable); Commit.signature: GitSignature; GitSignature.signer: User
	"data.nodes[]":                  true,
	"data.nodes[].signature":        true,
	"data.nodes[].signature.signer": true,
}

var listIndex = regexp.MustCompile(`\[\d+\]`)

// oddVariants returns the schema-legal variants of reply: nullable fields made
// null and lists emptied.
func oddVariants(t *testing.T, reply string) []testHelpers.NullVariant {
	t.Helper()
	all, err := testHelpers.NullVariants([]byte(reply))
	require.NoError(t, err)
	legal := []testHelpers.NullVariant{}
	for _, v := range all {
		if strings.HasSuffix(v.Path, "[]") || oddNullable[listIndex.ReplaceAllString(v.Path, "[]")] {
			legal = append(legal, v)
		}
	}
	return legal
}

type oddEntryPoint struct {
	name    string
	replies []string // the replies this entry point's queries read
	run     func(c *GithubConfig) (*types.PREvidence, error)
}

var oddEntryPoints = []oddEntryPoint{
	{
		name:    "PREvidenceByPRNumber",
		replies: []string{replyByNumber, replyCommitPage, replyReviewPage, replySigners},
		run:     func(c *GithubConfig) (*types.PREvidence, error) { return c.PREvidenceByPRNumber(7) },
	},
	{
		name:    "PREvidenceForCommitV2",
		replies: []string{replyForCommit, replyCommitPage, replyReviewPage, replySigners},
		run: func(c *GithubConfig) (*types.PREvidence, error) {
			prs, err := c.PREvidenceForCommitV2("merge-sha")
			if len(prs) == 0 {
				return nil, err
			}
			return prs[0], err
		},
	},
}

// runOddReply runs ep against replies and describes the result: the outcome,
// plus the evidence fields that differ from want when it is ok.
func runOddReply(t *testing.T, ep oddEntryPoint, replies map[string]string, want *types.PREvidence) string {
	t.Helper()
	ts := newOddReplyServer(t, replies)
	config := &GithubConfig{Token: "fake-token", BaseURL: ts.URL, Org: "test-org", Repository: "test-repo",
		Sleep: func(time.Duration) {}}
	return testHelpers.EvidenceOutcome([]*types.PREvidence{want}, func() ([]*types.PREvidence, error) {
		got, err := ep.run(config)
		if got == nil {
			return nil, err
		}
		return []*types.PREvidence{got}, err
	})
}

func oddBaselineEvidence(t *testing.T, ep oddEntryPoint) *types.PREvidence {
	t.Helper()
	ts := newOddReplyServer(t, oddBaselineReplies())
	evidence, err := ep.run(&GithubConfig{Token: "fake-token", BaseURL: ts.URL, Org: "test-org",
		Repository: "test-repo", Sleep: func(time.Duration) {}})
	require.NoError(t, err)
	require.NotNil(t, evidence)
	return evidence
}

func TestOddReplies_BaselineGivesFullEvidence(t *testing.T) {
	for _, ep := range oddEntryPoints {
		t.Run(ep.name, func(t *testing.T) {
			e := oddBaselineEvidence(t, ep)
			require.Equal(t, "merge-sha", e.MergeCommit)
			require.Equal(t, "MERGED", e.State)
			require.Equal(t, "ada", e.Author)
			require.NotZero(t, e.CreatedAt)
			require.NotZero(t, e.MergedAt)
			require.Equal(t, "sha2", e.HeadSHA)
			require.Equal(t, []string{"sha1", "sha2"}, shasIn(e))
			require.Equal(t, "ada", e.Commits[0].SignerUsername)
			require.True(t, *e.Commits[0].Verified)
			require.Nil(t, e.Commits[1].Verified)
			require.Len(t, e.Reviews, 2)
			require.Equal(t, "sha2", e.Reviews[0].CommitSHA)
			require.Equal(t, "CHANGES_REQUESTED", e.Reviews[1].State)
			require.Len(t, e.Approvers, 1)
		})
	}
}

// oddAllowedErrors are the nullable fields whose null fails the evidence
// lookup instead of leaving part of it empty, keyed as in oddNullable.
var oddAllowedErrors = map[string]string{
	"data.repository":             "the repository every query is about; null means it is missing or unreadable",
	"data.repository.pullRequest": "the pull request asked for by number; null means it is missing or unreadable",
}

// oddWholePRLost are the V2 fields whose null means the commit has no pull
// requests, which is a valid answer.
var oddWholePRLost = map[string]bool{
	"data.repository.object":                              true,
	"data.repository.object.associatedPullRequests":       true,
	"data.repository.object.associatedPullRequests.nodes": true,
}

// oddAffected is the evidence a null at path may empty or change.
func oddAffected(reply, path string) []string {
	switch {
	case reply == replySigners || strings.Contains(path, ".commits"):
		return []string{"commits"}
	case strings.Contains(path, ".reviews"):
		return []string{"approvers", "reviews"}
	}
	last := path[strings.LastIndex(path, ".")+1:]
	return map[string][]string{
		"mergedAt":    {"merged_at"},
		"mergeCommit": {"merge_commit"},
		"author":      {"author"},
	}[last]
}

// requireRuleHolds checks one outcome against the rule for a variant of reply
// with the value at path made null, or emptied for a path ending "[]".
func requireRuleHolds(t *testing.T, reply, path, outcome string) {
	t.Helper()
	require.NotContains(t, outcome, "panic")
	key := listIndex.ReplaceAllString(path, "[]")
	switch {
	case strings.HasSuffix(path, "[]"):
		require.NotEqual(t, "error", outcome, "an emptied list is not an error")
	case strings.HasSuffix(path, "]") && reply != replySigners:
		// A null signer node only loses its signer, which is optional.
		require.Equal(t, "error", outcome, "a null list item is missing data, not an optional field")
	case oddAllowedErrors[key] != "":
		require.Equal(t, "error", outcome, oddAllowedErrors[key])
	case isApprovalAuthor(t, reply, path):
		// A deleted account's approval is left out of reviews and approvers.
		require.True(t, strings.HasPrefix(outcome, "ok"), "a review by a deleted account is legal, got %q", outcome)
		require.Contains(t, outcome, ":approvers", "an approval by nobody must not count")
	case oddWholePRLost[key]:
		require.Equal(t, "ok lost:PullRequests(1->0)", outcome)
	default:
		require.True(t, strings.HasPrefix(outcome, "ok"), "a null field is not an error, got %q", outcome)
		allowed := oddAffected(reply, path)
		for _, diff := range strings.Fields(outcome)[1:] {
			_, field, _ := strings.Cut(diff, ":")
			require.Contains(t, allowed, field, "a null at %s changed unrelated evidence", path)
		}
	}
}

// TestOddReplies_NullsFollowTheRule feeds each entry point every legal
// variant of every reply it reads.
func TestOddReplies_NullsFollowTheRule(t *testing.T) {
	for _, ep := range oddEntryPoints {
		want := oddBaselineEvidence(t, ep)
		for _, reply := range ep.replies {
			for _, v := range oddVariants(t, oddBaselineReplies()[reply]) {
				t.Run(ep.name+"/"+reply+"/"+v.Path, func(t *testing.T) {
					replies := oddBaselineReplies()
					replies[reply] = string(v.Body)
					requireRuleHolds(t, reply, v.Path, runOddReply(t, ep, replies, want))
				})
			}
		}
	}
}

// TestOddReplies_FieldErrorMeansTheSameAsNull adds an errors entry naming
// each nullable field made null, as GitHub does when it cannot resolve one.
func TestOddReplies_FieldErrorMeansTheSameAsNull(t *testing.T) {
	for _, ep := range oddEntryPoints {
		want := oddBaselineEvidence(t, ep)
		for _, reply := range ep.replies {
			for _, v := range oddVariants(t, oddBaselineReplies()[reply]) {
				if strings.HasSuffix(v.Path, "[]") {
					continue
				}
				t.Run(ep.name+"/"+reply+"/"+v.Path, func(t *testing.T) {
					plain := oddBaselineReplies()
					plain[reply] = string(v.Body)
					withError := oddBaselineReplies()
					withError[reply] = withErrorAt(t, string(v.Body), v.Path)
					got := runOddReply(t, ep, withError, want)
					if reply != replySigners && isRootField(t, v.Path) {
						require.Equal(t, "error", got, "GitHub's reason for a missing root field must reach the user")
						return
					}
					if reply != replySigners && holdsList(t, oddBaselineReplies()[reply], v.Path) {
						require.Equal(t, "error", got, "a list GitHub could not give would read as empty")
						return
					}
					require.Equal(t, runOddReply(t, ep, plain, want), got)
				})
			}
		}
	}
}

// isRootField reports whether path is at most two segments below data, such
// as repository.pullRequest or nodes[0], whose errors are never dropped.
func isRootField(t *testing.T, path string) bool {
	t.Helper()
	return len(errorPath(t, strings.TrimPrefix(path, "data."))) <= 2
}

// TestOddReplies_ErrorsThatAreNotANullFieldFail covers errors the client must
// not drop: one on a field under a null parent (a non-null field's error nulls
// its nearest nullable parent), one with no path, and one on a list item. The
// signer lookup is the exception: any error that comes with data only loses
// signers.
func TestOddReplies_ErrorsThatAreNotANullFieldFail(t *testing.T) {
	for _, ep := range oddEntryPoints {
		for _, reply := range ep.replies {
			baseline := oddBaselineReplies()[reply]
			all, err := testHelpers.NullVariants([]byte(baseline))
			require.NoError(t, err)
			for _, v := range all {
				if v.Path == "." || strings.HasSuffix(v.Path, "[]") {
					continue
				}
				for _, child := range keysAt(t, baseline, v.Path) {
					t.Run(ep.name+"/"+reply+"/"+v.Path+"."+child, func(t *testing.T) {
						replies := oddBaselineReplies()
						replies[reply] = withErrorAt(t, string(v.Body), v.Path+"."+child)
						requireFailsUnlessSigners(t, reply, v.Path, runOddReply(t, ep, replies, nil))
					})
				}
			}
			t.Run(ep.name+"/"+reply+"/no path", func(t *testing.T) {
				replies := oddBaselineReplies()
				replies[reply] = strings.TrimSuffix(baseline, "}") + `,"errors":[{"message":"Something went wrong"}]}`
				requireFailsUnlessSigners(t, reply, "", runOddReply(t, ep, replies, nil))
			})
		}
		t.Run(ep.name+"/list item", func(t *testing.T) {
			replies := oddBaselineReplies()
			replies[replyCommitPage] = withErrorAt(t, strings.Replace(oddCommitPageReply, oddUnsignedCommit, "null", 1),
				"data.repository.pullRequest.commits.nodes[0]")
			require.Equal(t, "error", runOddReply(t, ep, replies, nil))
		})
	}
}

// requireFailsUnlessSigners expects an error, except from a signer lookup
// reply that still has data, as it does unless nulledPath is "data".
func requireFailsUnlessSigners(t *testing.T, reply, nulledPath, outcome string) {
	t.Helper()
	if reply == replySigners && nulledPath != "data" {
		require.True(t, strings.HasPrefix(outcome, "ok"), "the signer lookup tolerates errors with data, got %q", outcome)
		return
	}
	require.Equal(t, "error", outcome)
}

// TestOddReplies_ReviewWithNoAuthor covers a deleted account: GraphQL sends a
// null author, and "ghost" is GitHub's stand-in login for one. Policies read
// Reviews, so only a change request is kept there without a name.
func TestOddReplies_ReviewWithNoAuthor(t *testing.T) {
	for _, author := range []string{"null", `{"login":"ghost"}`} {
		for _, ep := range oddEntryPoints {
			t.Run(ep.name+"/approval/"+author, func(t *testing.T) {
				replies := oddBaselineReplies()
				noAuthor := strings.Replace(oddApproval, `{"login":"grace"}`, author, 1)
				for _, name := range []string{replyByNumber, replyForCommit} {
					replies[name] = strings.Replace(replies[name], oddApproval, noAuthor, 1)
				}
				e := runReviewReplies(t, ep, replies)
				require.Empty(t, e.Approvers)
				require.Len(t, e.Reviews, 1)
				require.Equal(t, types.ReviewChangesRequested, e.Reviews[0].State)
			})
			t.Run(ep.name+"/change request/"+author, func(t *testing.T) {
				replies := oddBaselineReplies()
				replies[replyReviewPage] = strings.Replace(replies[replyReviewPage], `{"login":"linus"}`, author, 1)
				e := runReviewReplies(t, ep, replies)
				require.Len(t, e.Approvers, 1)
				require.Len(t, e.Reviews, 2)
				require.Equal(t, "", e.Reviews[1].Username)
				require.Equal(t, types.ReviewChangesRequested, e.Reviews[1].State)
			})
		}
	}
}

func runReviewReplies(t *testing.T, ep oddEntryPoint, replies map[string]string) *types.PREvidence {
	t.Helper()
	ts := newOddReplyServer(t, replies)
	e, err := ep.run(&GithubConfig{Token: "fake-token", BaseURL: ts.URL, Org: "test-org",
		Repository: "test-repo", Sleep: func(time.Duration) {}})
	require.NoError(t, err)
	return e
}

// isApprovalAuthor reports whether path is the author of an approving review
// in the named baseline reply.
func isApprovalAuthor(t *testing.T, reply, path string) bool {
	t.Helper()
	parent, field, _ := cutLast(path)
	if field != "author" || !strings.Contains(parent, ".reviews.nodes[") {
		return false
	}
	review, _ := valueAt(t, oddBaselineReplies()[reply], parent).(map[string]any)
	return review["state"] == types.ReviewApproved
}

// holdsList reports whether path in doc is a list or a connection (an object
// holding a nodes list).
func holdsList(t *testing.T, doc, path string) bool {
	t.Helper()
	switch v := valueAt(t, doc, path).(type) {
	case []any:
		return true
	case map[string]any:
		_, ok := v["nodes"]
		return ok
	}
	return false
}

// cutLast splits "a.b.c" into "a.b" and "c".
func cutLast(path string) (string, string, bool) {
	i := strings.LastIndex(path, ".")
	if i < 0 {
		return "", path, false
	}
	return path[:i], path[i+1:], true
}

// valueAt returns the value at path in doc.
func valueAt(t *testing.T, doc, path string) any {
	t.Helper()
	var cur any
	require.NoError(t, json.Unmarshal([]byte(doc), &cur))
	for _, seg := range errorPath(t, path) {
		switch s := seg.(type) {
		case string:
			cur = cur.(map[string]any)[s]
		case int:
			cur = cur.([]any)[s]
		}
	}
	return cur
}

// keysAt names the fields of the object at path in doc, if it is one.
func keysAt(t *testing.T, doc, path string) []string {
	t.Helper()
	obj, _ := valueAt(t, doc, path).(map[string]any)
	keys := []string{}
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// withErrorAt adds to body an errors entry whose path names path.
func withErrorAt(t *testing.T, body, path string) string {
	t.Helper()
	entry, err := json.Marshal(map[string]any{"type": "NOT_FOUND", "message": "Could not resolve",
		"path": errorPath(t, strings.TrimPrefix(path, "data."))})
	require.NoError(t, err)
	return strings.TrimSuffix(body, "}") + `,"errors":[` + string(entry) + `]}`
}

// errorPath turns "a.b[2].c" into GraphQL's ["a","b",2,"c"]. A leading
// "data" segment is kept, so callers strip it when building a reply path.
func errorPath(t *testing.T, path string) []any {
	t.Helper()
	out := []any{}
	for _, part := range strings.Split(path, ".") {
		name, rest, _ := strings.Cut(part, "[")
		out = append(out, name)
		for rest != "" {
			idx, more, _ := strings.Cut(rest, "]")
			i, err := strconv.Atoi(idx)
			require.NoError(t, err)
			out = append(out, i)
			rest = strings.TrimPrefix(more, "[")
		}
	}
	return out
}

// A null list item's error says what the list belongs to: the commit for its
// pull requests, or the pull request for its commits and reviews.
func TestOddReplies_NullItemErrorNamesItsOwner(t *testing.T) {
	for _, ep := range oddEntryPoints {
		for _, reply := range ep.replies {
			if reply == replySigners {
				continue // A null signer node isn't an error.
			}
			variants, err := testHelpers.NullVariants([]byte(oddBaselineReplies()[reply]))
			require.NoError(t, err)
			for _, v := range variants {
				if !strings.HasSuffix(v.Path, "]") || strings.HasSuffix(v.Path, "[]") {
					continue
				}
				t.Run(ep.name+"/"+reply+"/"+v.Path, func(t *testing.T) {
					replies := oddBaselineReplies()
					replies[reply] = string(v.Body)
					ts := newOddReplyServer(t, replies)
					_, err := ep.run(&GithubConfig{Token: "fake-token", BaseURL: ts.URL, Org: "test-org",
						Repository: "test-repo", Sleep: func(time.Duration) {}})
					owner := "for test-org/test-repo#7 "
					if regexp.MustCompile(`associatedPullRequests\.nodes\[\d+\]$`).MatchString(v.Path) {
						owner = "for commit merge-sha "
					}
					require.ErrorContains(t, err, owner)
				})
			}
		}
	}
}

// A commit with no author is recorded with an empty author, not " <>".
func TestOddReplies_MissingCommitAuthorIsEmpty(t *testing.T) {
	author := `"author":{"name":"Ada","email":"ada@example.com","user":{"login":"ada"}},`
	for _, ep := range oddEntryPoints {
		t.Run(ep.name, func(t *testing.T) {
			replies := oddBaselineReplies()
			require.Equal(t, 1, strings.Count(replies[replyCommitPage], author))
			replies[replyCommitPage] = strings.Replace(replies[replyCommitPage], author, `"author":null,`, 1)
			ts := newOddReplyServer(t, replies)
			evidence, err := ep.run(&GithubConfig{Token: "fake-token", BaseURL: ts.URL, Org: "test-org",
				Repository: "test-repo", Sleep: func(time.Duration) {}})
			require.NoError(t, err)
			var authors []string
			for _, c := range evidence.Commits {
				authors = append(authors, c.Author)
			}
			require.Equal(t, []string{"Ada <ada@example.com>", ""}, authors)
		})
		t.Run(ep.name+"/name only", func(t *testing.T) {
			replies := oddBaselineReplies()
			replies[replyCommitPage] = strings.Replace(replies[replyCommitPage], author, `"author":{"name":"Ada","email":null,"user":null},`, 1)
			ts := newOddReplyServer(t, replies)
			evidence, err := ep.run(&GithubConfig{Token: "fake-token", BaseURL: ts.URL, Org: "test-org",
				Repository: "test-repo", Sleep: func(time.Duration) {}})
			require.NoError(t, err)
			require.Equal(t, "Ada <>", evidence.Commits[1].Author, "one part is enough to keep the format")
		})
		t.Run(ep.name+"/email only", func(t *testing.T) {
			replies := oddBaselineReplies()
			replies[replyCommitPage] = strings.Replace(replies[replyCommitPage], author, `"author":{"name":null,"email":"ada@example.com","user":null},`, 1)
			ts := newOddReplyServer(t, replies)
			evidence, err := ep.run(&GithubConfig{Token: "fake-token", BaseURL: ts.URL, Org: "test-org",
				Repository: "test-repo", Sleep: func(time.Duration) {}})
			require.NoError(t, err)
			require.Equal(t, " <ada@example.com>", evidence.Commits[1].Author, "one part is enough to keep the format")
		})
	}
}
