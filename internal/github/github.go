package github

import (
	"context"

	gh "github.com/google/go-github/v42/github"

	"golang.org/x/oauth2"
)

// NewGithubClient returns Github client with a token and context
func NewGithubClientFromToken(ctx context.Context, ghToken string) *gh.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := gh.NewClient(tc)
	return client
}

// PullRequestsForCommit returns a list of pull requests for a specific commit
func PullRequestsForCommit(ghToken, ghOwner, repository, commit string) ([]*gh.PullRequest, error) {
	ctx := context.Background()
	client := NewGithubClientFromToken(ctx, ghToken)

	pullrequests, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, ghOwner, repository,
		commit, &gh.PullRequestListOptions{})
	return pullrequests, err
}

// GetPullRequestApprovers returns a list of approvers for a given pull request
func GetPullRequestApprovers(ghToken, ghOwner, repository string, number int) ([]string, error) {
	approvers := []string{}
	ctx := context.Background()
	client := NewGithubClientFromToken(ctx, ghToken)
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
