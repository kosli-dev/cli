package gitlab

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

type GitlabConfig struct {
	Token      string
	BaseURL    string
	Org        string
	Repository string
}

// GetClientOptFns creates a list of ClientOptionFunc
func (c *GitlabConfig) GetClientOptFns() []gitlab.ClientOptionFunc {
	clientOptFns := []gitlab.ClientOptionFunc{}
	if c.BaseURL != "" {
		clientOptFns = append(clientOptFns, gitlab.WithBaseURL(c.BaseURL))
	}
	return clientOptFns
}

// NewGitlabClientFromToken returns an API client from GitlabConfig
func (c *GitlabConfig) NewGitlabClientFromToken() (*gitlab.Client, error) {
	client, err := gitlab.NewClient(c.Token, c.GetClientOptFns()...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// projectID returns a project ID that can be
// used when making calls to Gitlab API
func (c *GitlabConfig) ProjectID() string {
	return fmt.Sprintf("%s/%s", c.Org, c.Repository)
}

// MergeRequestsForCommit returns a list of MRs for a given commit
func (c *GitlabConfig) MergeRequestsForCommit(commit string) ([]*gitlab.MergeRequest, error) {
	mrs := []*gitlab.MergeRequest{}
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return mrs, err
	}

	mrs, _, err = client.Commits.ListMergeRequestsByCommit(c.ProjectID(), commit)
	if err != nil {
		return mrs, err
	}
	return mrs, nil
}

// GetMergeRequestApprovers returns a list of users (name and username) who approved an MR
func (c *GitlabConfig) GetMergeRequestApprovers(mrIID int) ([]string, error) {
	approvers := []string{}
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return approvers, err
	}
	approvals, _, err := client.MergeRequestApprovals.GetConfiguration(c.ProjectID(), mrIID)
	if err != nil {
		return approvers, err
	}
	for _, approver := range approvals.ApprovedBy {
		approvers = append(approvers, fmt.Sprintf("%s (@%s)", approver.User.Name, approver.User.Username))
	}
	return approvers, nil
}
