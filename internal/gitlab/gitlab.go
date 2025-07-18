package gitlab

import (
	"fmt"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/kosli-dev/cli/internal/utils"
	gitlab "gitlab.com/gitlab-org/api/client-go"
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

// This is the old implementation, it will be removed after the PR payload is enhanced for Gitlab
func (c *GitlabConfig) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	mrs, err := c.MergeRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mr := range mrs {
		evidence, err := c.newPRGitlabEvidence(mr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

// This is the new implementation, it will be used for Gitlab
func (c *GitlabConfig) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	return []*types.PREvidence{}, nil
}

func (c *GitlabConfig) newPRGitlabEvidence(mr *gitlab.BasicMergeRequest) (*types.PREvidence, error) {
	evidence := &types.PREvidence{
		URL:         mr.WebURL,
		MergeCommit: mr.MergeCommitSHA,
		State:       mr.State,
	}
	approvers, err := c.GetMergeRequestApprovers(mr.IID)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = utils.ConvertStringListToInterfaceList(approvers)
	return evidence, nil
}

// MergeRequestsForCommit returns a list of MRs for a given commit
func (c *GitlabConfig) MergeRequestsForCommit(commit string) ([]*gitlab.BasicMergeRequest, error) {
	mrs := []*gitlab.BasicMergeRequest{}
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return mrs, err
	}

	mrs, _, err = client.Commits.ListMergeRequestsByCommit(c.ProjectID(), commit)
	if err != nil {
		return mrs, fmt.Errorf("failed to list merge requests for commit %s: %v", commit, err)
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
