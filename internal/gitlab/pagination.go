package gitlab

import (
	"fmt"
	"strconv"

	retryablehttp "github.com/hashicorp/go-retryablehttp"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	// defaultMaxPages bounds every pagination loop. It guards against an API
	// that keeps reporting more results; it is not a tunable limit.
	defaultMaxPages = 100
	// listPageSize is GitLab's maximum. The zero-valued options used before
	// sent no per_page, so GitLab's default of 20 applied (#1082).
	listPageSize = 100
)

// listMergeRequestCommits returns every commit on an MR. GitLab's default page
// size is 20, which silently truncated the commit list on larger MRs (#1082).
// A cap or a non-advancing page errors rather than returning early: quietly
// stopping is the bug being fixed.
func (c *GitlabConfig) listMergeRequestCommits(client *gitlab.Client, mrIID int64, maxPages int) ([]*gitlab.Commit, error) {
	all := []*gitlab.Commit{}
	opts := &gitlab.GetMergeRequestCommitsOptions{
		ListOptions: gitlab.ListOptions{PerPage: listPageSize, Page: 1},
	}
	for pages := 0; pages < maxPages; pages++ {
		glCommits, resp, err := client.MergeRequests.GetMergeRequestCommits(c.ProjectID(), mrIID, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, glCommits...)
		if resp.NextPage == 0 {
			return all, nil
		}
		if resp.NextPage <= opts.Page {
			return nil, fmt.Errorf("next page %d did not advance past page %d for merge request %d",
				resp.NextPage, opts.Page, mrIID)
		}
		opts.Page = resp.NextPage
	}
	return nil, fmt.Errorf("aborting after %d pages of commits for merge request %d", maxPages, mrIID)
}

// withPage sets the offset-pagination query params. ListMergeRequestsByCommit
// takes no list options struct, so they go straight onto the request.
func withPage(page, perPage int64) gitlab.RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		q := req.URL.Query()
		q.Set("page", strconv.FormatInt(page, 10))
		q.Set("per_page", strconv.FormatInt(perPage, 10))
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// listMergeRequestsForCommit returns every MR associated with a commit. Sending
// no pagination params left this at GitLab's default of 20, the same truncation
// as listMergeRequestCommits above and in the same PREvidenceForCommitV2 path
// (#1082).
func (c *GitlabConfig) listMergeRequestsForCommit(client *gitlab.Client, commit string, maxPages int) ([]*gitlab.BasicMergeRequest, error) {
	all := []*gitlab.BasicMergeRequest{}
	page := int64(1)
	for pages := 0; pages < maxPages; pages++ {
		mrs, resp, err := client.Commits.ListMergeRequestsByCommit(c.ProjectID(), commit,
			withPage(page, listPageSize))
		if err != nil {
			return nil, err
		}
		all = append(all, mrs...)
		if resp.NextPage == 0 {
			return all, nil
		}
		if resp.NextPage <= page {
			return nil, fmt.Errorf("next page %d did not advance past page %d for commit %s",
				resp.NextPage, page, commit)
		}
		page = resp.NextPage
	}
	return nil, fmt.Errorf("aborting after %d pages of merge requests for commit %s", maxPages, commit)
}
