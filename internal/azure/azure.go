package azure

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

type AzureConfig struct {
	Token      string
	OrgURL     string
	Repository string
	Project    string
}

type AzureFlagsTempValueHolder struct {
	Token      string
	OrgUrl     string
	Repository string
	Project    string
}

// NewAzureConfig returns a new AzureConfig
func NewAzureConfig(token, orgURL, project, repository string) *AzureConfig {
	return &AzureConfig{
		Token:      token,
		OrgURL:     orgURL,
		Project:    project,
		Repository: extractRepoName(repository),
	}
}

// extractRepoName returns repository name from 'org/repository_name' string
func extractRepoName(fullRepositoryName string) string {
	repoNameParts := strings.Split(fullRepositoryName, "/")
	repository := repoNameParts[len(repoNameParts)-1]
	return repository
}

// NewAzureClientFromToken returns Azure client with a token and context
func NewAzureClientFromToken(ctx context.Context, azToken, orgURL string) (git.Client, error) {
	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(orgURL, azToken)
	gitClient, err := git.NewClient(ctx, connection)
	if err != nil {
		return nil, err
	}

	return gitClient, nil
}

// This is the old implementation, it will be removed after the PR payload is enhanced for Azure
func (c *AzureConfig) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	prs, err := c.PullRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, pr := range prs {
		evidence, err := c.newPRAzureEvidence(pr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

// This is the new implementation, it will be used for Azure
func (c *AzureConfig) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	prs, err := c.PullRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, pr := range prs {
		evidence, err := c.newPRAzureEvidenceV2(pr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

func (c *AzureConfig) newPRAzureEvidence(pr git.GitPullRequest) (*types.PREvidence, error) {
	prID := strconv.Itoa(*pr.PullRequestId)
	url, err := url.JoinPath(c.OrgURL, c.Project, "_git", c.Repository, "pullrequest", prID)
	if err != nil {
		return nil, err
	}
	evidence := &types.PREvidence{
		URL:         url,
		MergeCommit: *(pr.LastMergeCommit.CommitId),
		State:       string(*pr.Status),
	}
	evidence.Approvers, err = c.GetPullRequestApprovers(*pr.PullRequestId, 1)
	if err != nil {
		return evidence, err
	}
	return evidence, nil
}

// newPRAzureEvidenceV2 creates a new PREvidence for a given pull request in V2 format
func (c *AzureConfig) newPRAzureEvidenceV2(pr git.GitPullRequest) (*types.PREvidence, error) {
	prID := strconv.Itoa(*pr.PullRequestId)
	prURL, err := url.JoinPath(c.OrgURL, c.Project, "_git", c.Repository, "pullrequest", prID)
	if err != nil {
		return nil, err
	}

	evidence := &types.PREvidence{
		URL:         prURL,
		MergeCommit: *(pr.LastMergeCommit.CommitId),
		State:       string(*pr.Status),
		Author:      fmt.Sprintf("%s (%s)", *pr.CreatedBy.DisplayName, *pr.CreatedBy.UniqueName),
		CreatedAt:   pr.CreationDate.Time.Unix(),
		Title:       *pr.Title,
		HeadRef:     *pr.SourceRefName,
	}
	if pr.Status != nil && pr.ClosedDate != nil && *pr.Status == git.PullRequestStatusValues.Completed {
		evidence.MergedAt = pr.ClosedDate.Time.Unix()
	}
	commits, err := c.GetPullRequestCommits(pr)
	if err != nil {
		return evidence, err
	}
	evidence.Commits = commits
	evidence.Approvers, err = c.GetPullRequestApprovers(*pr.PullRequestId, 2)
	if err != nil {
		return evidence, err
	}
	return evidence, nil
}

// GetPullRequestCommits returns a list of commits for a given pull request
func (c *AzureConfig) GetPullRequestCommits(pr git.GitPullRequest) ([]types.Commit, error) {
	commits := []types.Commit{}

	ctx := context.Background()
	client, err := NewAzureClientFromToken(ctx, c.Token, c.OrgURL)
	if err != nil {
		return commits, err
	}

	prCommitsResponse, err := client.GetPullRequestCommits(ctx, git.GetPullRequestCommitsArgs{
		RepositoryId:  &c.Repository,
		PullRequestId: pr.PullRequestId,
		Project:       &c.Project,
	})
	if err != nil {
		return commits, err
	}

	for _, commit := range prCommitsResponse.Value {
		commits = append(commits, types.Commit{
			SHA:               *commit.CommitId,
			Message:           *commit.Comment,
			Committer:         *commit.Author.Name,
			Timestamp:         commit.Committer.Date.Time.Unix(),
			URL:               *commit.Url,
			Branch:            *pr.SourceRefName,
			CommitterUsername: *commit.Committer.Name,
		})
	}

	return commits, nil
}

// PullRequestsForCommit returns a list of pull requests for a specific commit
func (c *AzureConfig) PullRequestsForCommit(commit string) ([]git.GitPullRequest, error) {
	ctx := context.Background()
	client, err := NewAzureClientFromToken(ctx, c.Token, c.OrgURL)
	if err != nil {
		return []git.GitPullRequest{}, err
	}

	prQuery, err := client.GetPullRequestQuery(ctx, git.GetPullRequestQueryArgs{
		Queries: &git.GitPullRequestQuery{
			Queries: &[]git.GitPullRequestQueryInput{
				{
					Items: &[]string{commit},
					Type:  &git.GitPullRequestQueryTypeValues.LastMergeCommit,
				},
			},
		},
		RepositoryId: &c.Repository,
		Project:      &c.Project,
	})
	if err != nil {
		return nil, err
	}
	if prQuery != nil {
		results := prQuery.Results
		if len(*results) > 0 {
			prsForCommit := (*results)[0][commit]
			return prsForCommit, nil
		}
	}

	return []git.GitPullRequest{}, nil
}

// GetPullRequestApprovers returns a list of approvers for a given pull request
func (c *AzureConfig) GetPullRequestApprovers(prNumber, version int) ([]any, error) {
	var approvers []any
	ctx := context.Background()
	client, err := NewAzureClientFromToken(ctx, c.Token, c.OrgURL)
	if err != nil {
		return approvers, err
	}

	reviewers, err := client.GetPullRequestReviewers(ctx, git.GetPullRequestReviewersArgs{
		RepositoryId:  &c.Repository,
		PullRequestId: &prNumber,
		Project:       &c.Project,
	})
	if err != nil {
		return approvers, err
	}

	for _, r := range *reviewers {
		if *r.Vote == 10 {
			approverName := fmt.Sprintf("%s (%s)", *r.DisplayName, *r.UniqueName)
			if version == 1 {
				approvers = append(approvers, approverName)
			} else {
				approvers = append(approvers, types.PRApprovals{
					Username: approverName,
					State:    "APPROVED",
				})
			}
		}
	}
	return approvers, nil
}
