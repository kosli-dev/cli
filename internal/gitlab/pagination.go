package gitlab

import (
	"fmt"

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
		if resp == nil || resp.NextPage == 0 {
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
