package github

import (
	"context"

	gh "github.com/google/go-github/v42/github"

	"golang.org/x/oauth2"
)

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

// PullRequestsForCommit returns a list of pull requests for a specific commit
func PullRequestsForCommit(ghToken, ghOwner, repository, commit, baseURL string) ([]*gh.PullRequest, error) {
	ctx := context.Background()
	client, err := NewGithubClientFromToken(ctx, ghToken, baseURL)
	if err != nil {
		return []*gh.PullRequest{}, err
	}

	pullrequests, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, ghOwner, repository,
		commit, &gh.PullRequestListOptions{})
	return pullrequests, err
}

// GetPullRequestApprovers returns a list of approvers for a given pull request
func GetPullRequestApprovers(ghToken, ghOwner, repository string, number int, baseURL string) ([]string, error) {
	approvers := []string{}
	ctx := context.Background()
	client, err := NewGithubClientFromToken(ctx, ghToken, baseURL)
	if err != nil {
		return approvers, err
	}
	reviews, _, err := client.PullRequests.ListReviews(ctx, ghOwner, repository, number, &gh.ListOptions{})
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
