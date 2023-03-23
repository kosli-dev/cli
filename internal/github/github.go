package github

import (
	"context"
	"strings"

	gh "github.com/google/go-github/v42/github"
	"github.com/kosli-dev/cli/internal/types"

	"golang.org/x/oauth2"
)

type GithubConfig struct {
	Token      string
	BaseURL    string
	Org        string
	Repository string
}

type GithubFlagsTempValueHolder struct {
	Token      string
	BaseURL    string
	Org        string
	Repository string
}

// NewGithubConfig returns a new GithubConfig
func NewGithubConfig(token, baseURL, org, repository string) *GithubConfig {
	return &GithubConfig{
		Token:   token,
		BaseURL: baseURL,
		Org:     org,
		// repository name must be extracted if a user is using default value from ${GITHUB_REPOSITORY}
		// because the value is in the format of "org/repository"
		Repository: extractRepoName(repository),
	}
}

// extractRepoName returns repository name from 'org/repository_name' string
func extractRepoName(fullRepositoryName string) string {
	repoNameParts := strings.Split(fullRepositoryName, "/")
	repository := repoNameParts[len(repoNameParts)-1]
	return repository
}

// NewGithubClientFromToken returns Github client with a token and context
func NewGithubClientFromToken(ctx context.Context, ghToken string, baseURL string) (*gh.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	if baseURL != "" {
		client, err := gh.NewEnterpriseClient(baseURL, baseURL, tc)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return gh.NewClient(tc), nil
}

func (c *GithubConfig) PREvidenceForCommit(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	prs, err := c.PullRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, pr := range prs {
		evidence, err := c.newPRGithubEvidence(pr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

func (c *GithubConfig) newPRGithubEvidence(pr *gh.PullRequest) (*types.PREvidence, error) {
	evidence := &types.PREvidence{
		URL:         pr.GetHTMLURL(),
		MergeCommit: pr.GetMergeCommitSHA(),
		State:       pr.GetState(),
	}
	approvers, err := c.GetPullRequestApprovers(pr.GetNumber())
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	return evidence, nil
}

// PullRequestsForCommit returns a list of pull requests for a specific commit
func (c *GithubConfig) PullRequestsForCommit(commit string) ([]*gh.PullRequest, error) {
	ctx := context.Background()
	client, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL)
	if err != nil {
		return []*gh.PullRequest{}, err
	}

	pullrequests, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, c.Org, c.Repository,
		commit, &gh.PullRequestListOptions{})
	return pullrequests, err
}

// GetPullRequestApprovers returns a list of approvers for a given pull request
func (c *GithubConfig) GetPullRequestApprovers(number int) ([]string, error) {
	approvers := []string{}
	ctx := context.Background()
	client, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL)
	if err != nil {
		return approvers, err
	}
	reviews, _, err := client.PullRequests.ListReviews(ctx, c.Org, c.Repository, number, &gh.ListOptions{})
	if err != nil {
		return approvers, err
	}
	for _, r := range reviews {
		if r.GetState() == "APPROVED" {
			approvers = append(approvers, r.GetUser().GetLogin())
		}
	}
	return approvers, nil
}
