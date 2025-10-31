package gitlab

import (
	"fmt"

	"github.com/kosli-dev/cli/internal/types"
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

// This is the old implementation, it will be removed after the PR payload is enhanced for all VCS providers
func (c *GitlabConfig) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	mrs, err := c.MergeRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mr := range mrs {
		evidence, err := c.newPRGitlabEvidenceV1(mr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

// This is the new implementation, it will be used for all VCS providers
func (c *GitlabConfig) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	mrs, err := c.MergeRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mr := range mrs {
		evidence, err := c.newPRGitlabEvidenceV2(mr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

func (c *GitlabConfig) newPRGitlabEvidenceV1(mr *gitlab.BasicMergeRequest) (*types.PREvidence, error) {
	evidence := &types.PREvidence{
		URL:         mr.WebURL,
		MergeCommit: mr.MergeCommitSHA,
		State:       mr.State,
	}
	approvers, err := c.GetMergeRequestApprovers(mr.IID, 1)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	return evidence, nil
}

func (c *GitlabConfig) newPRGitlabEvidenceV2(mr *gitlab.BasicMergeRequest) (*types.PREvidence, error) {
	evidence := &types.PREvidence{
		URL:         mr.WebURL,
		MergeCommit: mr.MergeCommitSHA,
		State:       mr.State,
		Author:      fmt.Sprintf("%s (@%s)", mr.Author.Name, mr.Author.Username),
		CreatedAt:   mr.CreatedAt.Unix(),
		MergedAt:    mr.MergedAt.Unix(),
		Title:       mr.Title,
		HeadRef:     mr.SourceBranch,
	}
	approvers, err := c.GetMergeRequestApprovers(mr.IID, 2)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	commits, err := c.GetMergeRequestCommits(mr)
	if err != nil {
		return evidence, err
	}
	evidence.Commits = commits
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
func (c *GitlabConfig) GetMergeRequestApprovers(mrIID, version int) ([]any, error) {
	var approvers []any
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return approvers, err
	}
	approvals, _, err := client.MergeRequestApprovals.GetConfiguration(c.ProjectID(), mrIID)
	if err != nil {
		return approvers, err
	}

	for _, approver := range approvals.ApprovedBy {
		approverName := fmt.Sprintf("%s (@%s)", approver.User.Name, approver.User.Username)
		if version == 1 {
			approvers = append(approvers, approverName)
		} else {
			approvers = append(approvers, types.PRApprovals{
				Username: approverName,
			})
		}
	}
	return approvers, nil
}

// GetMergeRequestCommits returns a list of commits for a given MR
func (c *GitlabConfig) GetMergeRequestCommits(mr *gitlab.BasicMergeRequest) ([]types.Commit, error) {
	commits := []types.Commit{}
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return commits, err
	}
	glCommits, _, err := client.MergeRequests.GetMergeRequestCommits(c.ProjectID(), mr.IID,
		&gitlab.GetMergeRequestCommitsOptions{})
	if err != nil {
		return commits, err
	}
	for _, commit := range glCommits {
		commits = append(commits, types.Commit{
			SHA:       commit.ID,
			Message:   commit.Message,
			Committer: fmt.Sprintf("%s <%s>", commit.CommitterName, commit.CommitterEmail),
			Timestamp: commit.CreatedAt.Unix(),
			Branch:    mr.SourceBranch,
			URL:       commit.WebURL,
		})
	}
	return commits, nil
}
