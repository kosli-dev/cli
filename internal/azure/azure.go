package azure

import (
	"context"
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

func (c *AzureConfig) PREvidenceForCommit(commit string) ([]*types.PREvidence, error) {
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
	approvers, err := c.GetPullRequestApprovers(*pr.PullRequestId)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	return evidence, nil
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
func (c *AzureConfig) GetPullRequestApprovers(number int) ([]string, error) {
	approvers := []string{}
	ctx := context.Background()
	client, err := NewAzureClientFromToken(ctx, c.Token, c.OrgURL)
	if err != nil {
		return approvers, err
	}

	reviewers, err := client.GetPullRequestReviewers(ctx, git.GetPullRequestReviewersArgs{
		RepositoryId:  &c.Repository,
		PullRequestId: &number,
		Project:       &c.Project,
	})
	if err != nil {
		return approvers, err
	}

	for _, r := range *reviewers {
		if *r.Vote == 10 {
			approvers = append(approvers, *r.DisplayName)
		}
	}
	return approvers, nil
}
