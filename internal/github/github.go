package github

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	gh "github.com/google/go-github/v42/github"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/shurcooL/graphql"

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

func graphqlEndpoint(baseURL string) string {
	if baseURL == "" || baseURL == "https://api.github.com" {
		return "https://api.github.com/graphql"
	}
	return strings.TrimSuffix(baseURL, "/") + "/api/graphql"
}

func (c *GithubConfig) PREvidenceForCommit(commit string) ([]*types.PREvidence, error) {
	ctx := context.Background()
	pullRequestsEvidence := []*types.PREvidence{}

	ghClient, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL)
	if err != nil {
		return pullRequestsEvidence, err
	}
	httpClient := ghClient.Client()

	client := graphql.NewClient(graphqlEndpoint(c.BaseURL), httpClient)

	var query struct {
		Repository struct {
			Object struct {
				Commit struct {
					AssociatedPullRequests struct {
						Nodes []struct {
							Number      graphql.Int
							Title       graphql.String
							State       graphql.String
							HeadRefName graphql.String
							URL         graphql.String
							CreatedAt   graphql.String
							MergedAt    graphql.String

							Author struct {
								Login graphql.String
							}

							Commits struct {
								Nodes []struct {
									Commit struct {
										Oid             graphql.String
										MessageHeadline graphql.String
										CommittedDate   graphql.String
										Committer       struct {
											Name  graphql.String
											Email graphql.String
											Date  graphql.String
											User  *struct {
												Login graphql.String
											}
										}
									}
								}
								PageInfo struct {
									HasNextPage graphql.Boolean
									EndCursor   graphql.String
								}
							} `graphql:"commits(first: 100, after: $commitCursor)"`

							Reviews struct {
								Nodes []struct {
									Author struct {
										Login graphql.String
									}
									State       graphql.String
									SubmittedAt graphql.String
								}
								PageInfo struct {
									HasNextPage graphql.Boolean
									EndCursor   graphql.String
								}
							} `graphql:"reviews(first: 100, states: APPROVED, after: $reviewCursor)"`
						}
						PageInfo struct {
							HasNextPage graphql.Boolean
							EndCursor   graphql.String
						}
					} `graphql:"associatedPullRequests(first: 100, after: $prCursor)"`
				} `graphql:"... on Commit"`
			} `graphql:"object(oid: $commitSHA)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":        graphql.String(c.Org),
		"repo":         graphql.String(c.Repository),
		"commitSHA":    GitObjectID(commit),
		"prCursor":     (*graphql.String)(nil),
		"commitCursor": (*graphql.String)(nil),
		"reviewCursor": (*graphql.String)(nil),
	}

	err = client.Query(context.Background(), &query, variables)
	if err != nil {
		return pullRequestsEvidence, err
	}

	// Print results for demonstration
	for _, pr := range query.Repository.Object.Commit.AssociatedPullRequests.Nodes {
		createdAt, err := time.Parse(time.RFC3339, string(pr.CreatedAt))
		if err != nil {
			return pullRequestsEvidence, err
		}
		mergedAt := int64(0)
		if pr.MergedAt != "" {
			mergedAtTime, err := time.Parse(time.RFC3339, string(pr.MergedAt))
			if err != nil {
				return pullRequestsEvidence, err
			}
			mergedAt = mergedAtTime.Unix()
		}

		evidence := &types.PREvidence{
			URL:         string(pr.URL),
			MergeCommit: commit,
			State:       string(pr.State),
			Author:      string(pr.Author.Login),
			CreatedAt:   createdAt.Unix(),
			MergedAt:    mergedAt,
			Title:       string(pr.Title),
			HeadRef:     string(pr.HeadRefName),
			Approvers:   []interface{}{},
			Commits:     []types.Commit{},
		}

		for _, c := range pr.Commits.Nodes {
			timestamp, err := time.Parse(time.RFC3339, string(c.Commit.CommittedDate))
			if err != nil {
				return pullRequestsEvidence, err
			}

			evidence.Commits = append(evidence.Commits, types.Commit{
				SHA:       string(c.Commit.Oid),
				Message:   string(c.Commit.MessageHeadline),
				Committer: string(c.Commit.Committer.User.Login),
				Timestamp: timestamp.Unix(),
			})
		}

		for _, r := range pr.Reviews.Nodes {
			submittedAt, err := time.Parse(time.RFC3339, string(r.SubmittedAt))
			if err != nil {
				return pullRequestsEvidence, err
			}

			evidence.Approvers = append(evidence.Approvers, types.PRApprovals{
				Author:    string(r.Author.Login),
				State:     string(r.State),
				Timestamp: submittedAt.Unix(),
			})
		}

		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

type GitObjectID string

func (v GitObjectID) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, `"%s"`, string(v))
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
