package azure

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// The baseline replies follow the shapes in the Azure DevOps REST docs (7.1):
//   - PR query:     https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-request-query/get?view=azure-devops-rest-7.1
//   - PR object:    https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/get-pull-request?view=azure-devops-rest-7.1
//   - PR commits:   https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-request-commits/get-pull-request-commits?view=azure-devops-rest-7.1
//   - PR reviewers: https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-request-reviewers/list?view=azure-devops-rest-7.1
// Only the reviewers page carries a sample body; the others are built from the
// documented schemas, in the same fabrikam style.

const oddMergeCommit = "b60280bc6e62e2f880f1b63c1e24987664d3bda3"

const oddPRQueryReply = `{
  "queries": [{"items": ["b60280bc6e62e2f880f1b63c1e24987664d3bda3"], "type": "lastMergeCommit"}],
  "results": [{
    "b60280bc6e62e2f880f1b63c1e24987664d3bda3": [{
      "repository": {
        "id": "3411ebc1-d5aa-464f-9615-0b527bc66719",
        "name": "Fabrikam-Fiber-Git",
        "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719",
        "project": {"id": "a7573007-bbb3-4341-b726-0c4148a07853", "name": "Fabrikam-Fiber-Git", "state": "wellFormed", "visibility": "private"}
      },
      "pullRequestId": 22,
      "codeReviewId": 22,
      "status": "completed",
      "createdBy": {
        "displayName": "Normal Paulk",
        "url": "https://dev.azure.com/fabrikam/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db",
        "id": "d6245f20-2af8-44f4-9451-8107cb2767db",
        "uniqueName": "fabrikamfiber16@hotmail.com",
        "imageUrl": "https://dev.azure.com/fabrikam/_api/_common/identityImage?id=d6245f20-2af8-44f4-9451-8107cb2767db"
      },
      "creationDate": "2026-03-01T10:00:00.0000000Z",
      "closedDate": "2026-03-02T15:30:00.0000000Z",
      "title": "A new feature",
      "description": "Adding a new feature",
      "sourceRefName": "refs/heads/npaulk/my_work",
      "targetRefName": "refs/heads/new_feature",
      "mergeStatus": "succeeded",
      "isDraft": false,
      "mergeId": "f5fc8381-3fb2-49fe-8a0d-27dcc2d6ef82",
      "lastMergeSourceCommit": {"commitId": "b60280bc6e62e2f880f1b63c1e24987664d3bda3", "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"},
      "lastMergeTargetCommit": {"commitId": "f47bbc106853afe3c1b07a81754bce5f4b8dbf62", "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/f47bbc106853afe3c1b07a81754bce5f4b8dbf62"},
      "lastMergeCommit": {"commitId": "b60280bc6e62e2f880f1b63c1e24987664d3bda3", "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"},
      "reviewers": [{
        "reviewerUrl": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/reviewers/d6245f20-2af8-44f4-9451-8107cb2767db",
        "vote": 10,
        "displayName": "Normal Paulk",
        "uniqueName": "fabrikamfiber16@hotmail.com",
        "id": "d6245f20-2af8-44f4-9451-8107cb2767db"
      }],
      "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22",
      "supportsIterations": true
    }]
  }]
}`

const oddPRCommitsReply = `{
  "count": 2,
  "value": [
    {
      "commitId": "52a6ae5ef3a2fde5c3c9e4e0d6b8e7a1d2c3b4a5",
      "author": {"name": "Normal Paulk", "email": "fabrikamfiber16@hotmail.com", "date": "2026-03-01T09:00:00Z"},
      "committer": {"name": "Normal Paulk", "email": "fabrikamfiber16@hotmail.com", "date": "2026-03-01T09:00:00Z"},
      "comment": "Add the feature flag",
      "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/52a6ae5ef3a2fde5c3c9e4e0d6b8e7a1d2c3b4a5"
    },
    {
      "commitId": "7e9b1d2c3a4f5e6d7c8b9a0f1e2d3c4b5a6f7e8d",
      "author": {"name": "Chuck Reinhart", "email": "fabrikamfiber3@hotmail.com", "date": "2026-03-01T09:30:00Z"},
      "committer": {"name": "Chuck Reinhart", "email": "fabrikamfiber3@hotmail.com", "date": "2026-03-01T09:30:00Z"},
      "comment": "Wire the feature into the UI",
      "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/7e9b1d2c3a4f5e6d7c8b9a0f1e2d3c4b5a6f7e8d"
    }
  ]
}`

const oddPRReviewersReply = `{
  "count": 2,
  "value": [
    {
      "reviewerUrl": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/reviewers/d6245f20-2af8-44f4-9451-8107cb2767db",
      "vote": 10,
      "id": "d6245f20-2af8-44f4-9451-8107cb2767db",
      "displayName": "Normal Paulk",
      "uniqueName": "fabrikamfiber16@hotmail.com",
      "url": "https://dev.azure.com/fabrikam/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db",
      "imageUrl": "https://dev.azure.com/fabrikam/_api/_common/identityImage?id=d6245f20-2af8-44f4-9451-8107cb2767db"
    },
    {
      "reviewerUrl": "https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/reviewers/19d9411e-9a34-45bb-b985-d24d9d87c0c9",
      "vote": -10,
      "id": "19d9411e-9a34-45bb-b985-d24d9d87c0c9",
      "displayName": "Johnnie McLeod",
      "uniqueName": "fabrikamfiber2@hotmail.com",
      "url": "https://dev.azure.com/fabrikam/_apis/Identities/19d9411e-9a34-45bb-b985-d24d9d87c0c9",
      "imageUrl": "https://dev.azure.com/fabrikam/_api/_common/identityImage?id=19d9411e-9a34-45bb-b985-d24d9d87c0c9"
    }
  ]
}`

// oddLocations answers the SDK's OPTIONS /_apis discovery call with the
// routes PREvidenceForCommitV2 uses: resource areas (asked first, answered
// empty as on-prem servers do, so the org URL itself is used), PR query, PR
// commits and PR reviewers. Ids and templates are the SDK's own.
const oddLocations = `{"count": 4, "value": [
  {"id": "e81700f7-3be2-46de-8624-2eb35882fcaa", "area": "Location", "resourceName": "ResourceAreas", "routeTemplate": "_apis/{resource}/{areaId}", "resourceVersion": 1, "minVersion": "1.0", "maxVersion": "7.1", "releasedVersion": "7.1"},
  {"id": "b3a6eebe-9cf0-49ea-b6cb-1a4c5f5007b0", "area": "git", "resourceName": "pullRequestQuery", "routeTemplate": "{project}/_apis/{area}/repositories/{repositoryId}/{resource}", "resourceVersion": 1, "minVersion": "1.0", "maxVersion": "7.1", "releasedVersion": "7.1"},
  {"id": "52823034-34a8-4576-922c-8d8b77e9e4c4", "area": "git", "resourceName": "commits", "routeTemplate": "{project}/_apis/{area}/repositories/{repositoryId}/pullRequests/{pullRequestId}/{resource}", "resourceVersion": 1, "minVersion": "1.0", "maxVersion": "7.1", "releasedVersion": "7.1"},
  {"id": "4b6702c7-aa35-4b89-9c96-b9abf6d3e540", "area": "git", "resourceName": "reviewers", "routeTemplate": "{project}/_apis/{area}/repositories/{repositoryId}/pullRequests/{pullRequestId}/{resource}/{reviewerId}", "resourceVersion": 1, "minVersion": "1.0", "maxVersion": "7.1", "releasedVersion": "7.1"}
]}`

// fakeAzureDevOps serves the replies the SDK asks for. One server is shared
// by every variant because the SDK caches route discovery per base URL for
// the life of the process.
type fakeAzureDevOps struct {
	mu      sync.Mutex
	replies map[string]string // keyed by the last path segment
}

func (f *fakeAzureDevOps) set(replies map[string]string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies = replies
}

func (f *fakeAzureDevOps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	path := r.URL.Path
	switch {
	case r.Method == http.MethodOptions && strings.HasSuffix(path, "/_apis"):
		_, _ = io.WriteString(w, oddLocations)
		return
	case strings.HasSuffix(path, "/_apis/ResourceAreas"):
		_, _ = io.WriteString(w, `{"count": 0, "value": []}`)
		return
	}
	f.mu.Lock()
	body, ok := f.replies[path[strings.LastIndex(path, "/")+1:]]
	f.mu.Unlock()
	if !ok {
		http.Error(w, "unexpected "+r.Method+" "+path, http.StatusNotFound)
		return
	}
	_, _ = io.WriteString(w, body)
}

var oddBaselineReplies = map[string]string{
	"pullRequestQuery": oddPRQueryReply,
	"commits":          oddPRCommitsReply,
	"reviewers":        oddPRReviewersReply,
}

// oddAllowedErrors lists the values whose null may fail evidence collection,
// keyed by reply and path with list indexes written [i].
var oddAllowedErrors = map[string]string{
	"pullRequestQuery results[i]." + oddMergeCommit + "[i].pullRequestId": "the commits and reviewers requests need it",
	"commits value[i].commitId": "a commit with no id reads the same as a null one",
}

// oddAllowedSkips lists the list items whose null may be left out instead of
// failing, keyed as oddAllowedErrors is.
var oddAllowedSkips = map[string]string{
	"reviewers value[i]":          "a null reviewer reads as one with no vote, which has never been a review",
	"pullRequestQuery results[i]": "a null result reads as no pull requests, as on main; Azure may send it for a commit with none",
}

// TestPREvidenceOddReplies nulls each value (and empties each list) of each
// Azure DevOps reply in turn, through V1 (`assert pullrequest azure`) and V2,
// and checks the evidence code never panics, fails only on a null list item
// or an allowed value, and otherwise loses only what that value fed.
func TestPREvidenceOddReplies(t *testing.T) {
	fake := &fakeAzureDevOps{}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	config := NewAzureConfig("pat", server.URL+"/fabrikam", "Fabrikam-Fiber-Git", "Fabrikam-Fiber-Git")

	entryPoints := []struct {
		name    string
		fetch   func(string) ([]*types.PREvidence, error)
		replies []string
	}{
		{"V1", config.PREvidenceForCommitV1, []string{"pullRequestQuery", "reviewers"}},
		{"V2", config.PREvidenceForCommitV2, []string{"pullRequestQuery", "commits", "reviewers"}},
	}
	for _, entry := range entryPoints {
		fake.set(oddBaselineReplies)
		baseline, err := entry.fetch(oddMergeCommit)
		require.NoError(t, err)
		require.Len(t, baseline, 1)
		if entry.name == "V2" {
			requireFullAzureEvidence(t, baseline[0])
		} else {
			require.Equal(t, []any{"Normal Paulk (fabrikamfiber16@hotmail.com)"}, baseline[0].Approvers)
		}

		for _, reply := range entry.replies {
			variants, err := testHelpers.NullVariants([]byte(oddBaselineReplies[reply]))
			require.NoError(t, err)
			for _, v := range variants {
				t.Run(entry.name+"/"+reply+"/"+v.Path, func(t *testing.T) {
					replies := map[string]string{}
					for k, body := range oddBaselineReplies {
						replies[k] = body
					}
					replies[reply] = string(v.Body)
					fake.set(replies)

					outcome := testHelpers.EvidenceOutcome(baseline, func() ([]*types.PREvidence, error) {
						return entry.fetch(oddMergeCommit)
					})
					requireOddRule(t, reply, v.Path, outcome)
				})
			}
		}
	}
}

var listIndex = regexp.MustCompile(`\[\d+\]`)

func requireOddRule(t *testing.T, reply, path, outcome string) {
	t.Helper()
	require.False(t, strings.HasPrefix(outcome, "panic"), outcome)
	key := reply + " " + listIndex.ReplaceAllString(path, "[i]")
	feeds := oddFeeds(reply, strings.TrimSuffix(path, "[]"))
	readItem := strings.HasSuffix(path, "]") && !strings.HasSuffix(path, "[]") && len(feeds) > 0
	if outcome == "error" {
		_, allowed := oddAllowedErrors[key]
		require.True(t, allowed || readItem, "unexpected error")
		return
	}
	if readItem {
		_, skip := oddAllowedSkips[key]
		require.True(t, skip, "a null list item that is read must fail, got %q", outcome)
	}
	for _, diff := range strings.Fields(outcome)[1:] {
		field := diff[strings.Index(diff, ":")+1:]
		require.True(t, slices.Contains(feeds, field) || slices.Contains(feeds, "*"),
			"%s: %s is not fed by this value", diff, field)
	}
}

// oddFeeds names the evidence keys a reply value is read into; "*" for a
// value that holds the pull request itself.
func oddFeeds(reply, path string) []string {
	switch reply {
	case "commits":
		return []string{"commits"}
	case "reviewers":
		return []string{"approvers", "reviews"}
	}
	pr := "results[0]." + oddMergeCommit + "[0]"
	if path == "." || strings.HasPrefix(pr, path) {
		return []string{"*"}
	}
	field, ok := strings.CutPrefix(path, pr+".")
	if !ok {
		return nil
	}
	field, _, _ = strings.Cut(field, ".")
	return map[string][]string{
		"pullRequestId":   {"*"},
		"status":          {"state", "merged_at"},
		"createdBy":       {"author"},
		"creationDate":    {"created_at"},
		"closedDate":      {"merged_at"},
		"title":           {"title"},
		"sourceRefName":   {"head_ref", "commits"},
		"targetRefName":   {"base_ref"},
		"lastMergeCommit": {"merge_commit"},
	}[field]
}

// requireFullAzureEvidence checks the baseline fills every field Azure
// evidence can carry. HeadSHA is never set for Azure.
func requireFullAzureEvidence(t *testing.T, e *types.PREvidence) {
	t.Helper()
	v := reflect.ValueOf(*e)
	for i := 0; i < v.NumField(); i++ {
		name := v.Type().Field(i).Name
		if name == "HeadSHA" {
			continue
		}
		require.False(t, v.Field(i).IsZero(), "baseline evidence has no %s", name)
	}
	require.Len(t, e.Commits, 2)
	require.Len(t, e.Approvers, 1)
	require.Equal(t, []types.PRApprovals{
		{Username: "Normal Paulk (fabrikamfiber16@hotmail.com)", State: types.ReviewApproved},
		{Username: "Johnnie McLeod (fabrikamfiber2@hotmail.com)", State: types.ReviewChangesRequested},
	}, e.Reviews)
}

// A reviewer with no displayName or uniqueName (a deleted account) is left
// out of approvers, so it never counts as an approval, but is kept in reviews
// with an empty name. It never fails evidence collection.
func TestPREvidenceReviewerWithNoIdentity(t *testing.T) {
	fake := &fakeAzureDevOps{}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	config := NewAzureConfig("pat", server.URL+"/fabrikam", "Fabrikam-Fiber-Git", "Fabrikam-Fiber-Git")
	known := "Johnnie McLeod (fabrikamfiber2@hotmail.com)"

	for _, tc := range []struct {
		name  string
		vote  int
		state string // the nameless review's state; "" for none
	}{
		{"approved", 10, types.ReviewApproved},
		{"approved with suggestions", 5, types.ReviewApproved},
		{"waiting for author", -5, types.ReviewChangesRequested},
		{"rejected", -10, types.ReviewChangesRequested},
		{"no vote", 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			replies := map[string]string{}
			for k, body := range oddBaselineReplies {
				replies[k] = body
			}
			replies["reviewers"] = fmt.Sprintf(`{"count": 2, "value": [
  {"vote": %d, "id": "d6245f20-2af8-44f4-9451-8107cb2767db"},
  {"vote": 10, "displayName": "Johnnie McLeod", "uniqueName": "fabrikamfiber2@hotmail.com"}
]}`, tc.vote)
			fake.set(replies)

			v1, err := config.PREvidenceForCommitV1(oddMergeCommit)
			require.NoError(t, err)
			require.Equal(t, []any{known}, v1[0].Approvers)

			v2, err := config.PREvidenceForCommitV2(oddMergeCommit)
			require.NoError(t, err)
			require.Equal(t, []any{types.PRApprovals{Username: known, State: "APPROVED"}}, v2[0].Approvers)
			wantReviews := []types.PRApprovals{}
			if tc.state != "" {
				wantReviews = append(wantReviews, types.PRApprovals{Username: "", State: tc.state})
			}
			wantReviews = append(wantReviews, types.PRApprovals{Username: known, State: types.ReviewApproved})
			require.Equal(t, wantReviews, v2[0].Reviews)
		})
	}
}

// One name is enough for an approval to count.
func TestPREvidenceReviewerWithOneName(t *testing.T) {
	fake := &fakeAzureDevOps{}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	config := NewAzureConfig("pat", server.URL+"/fabrikam", "Fabrikam-Fiber-Git", "Fabrikam-Fiber-Git")

	for field, want := range map[string]string{
		"displayName": "Normal Paulk ()",
		"uniqueName":  " (Normal Paulk)",
	} {
		t.Run(field, func(t *testing.T) {
			replies := map[string]string{}
			for k, body := range oddBaselineReplies {
				replies[k] = body
			}
			replies["reviewers"] = fmt.Sprintf(`{"count": 1, "value": [{"vote": 10, %q: "Normal Paulk"}]}`, field)
			fake.set(replies)

			v1, err := config.PREvidenceForCommitV1(oddMergeCommit)
			require.NoError(t, err)
			require.Equal(t, []any{want}, v1[0].Approvers)
		})
	}
}
