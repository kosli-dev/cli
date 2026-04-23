package github

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	gh "github.com/google/go-github/v42/github"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/shurcooL/graphql"

	"golang.org/x/oauth2"
)

type GithubConfig struct {
	Token      string
	BaseURL    string
	Org        string
	Repository string
	// Sleep is called between retries in PREvidenceByPRNumber. Defaults to
	// time.Sleep when nil. Override in tests to avoid real delays.
	Sleep func(time.Duration)
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
	result, _ := url.JoinPath(baseURL, "api/graphql")
	return result
}

func (c *GithubConfig) ProviderAndLabel() (string, string) {
	return "github", "pull request"
}

// NewGithubRetrieverFunc creates a types.PRRetriever from GitHub config
// parameters. It can be replaced in tests to inject a FakeGitHubClient.
var NewGithubRetrieverFunc = defaultNewGithubRetriever

func defaultNewGithubRetriever(token, baseURL, org, repository string) types.PRRetriever {
	return NewGithubConfig(token, baseURL, org, repository)
}

// ResetGithubRetrieverFunc restores NewGithubRetrieverFunc to its default.
func ResetGithubRetrieverFunc() {
	NewGithubRetrieverFunc = defaultNewGithubRetriever
}

// PREvidenceForCommitHybrid tries PREvidenceForCommitV2 first. If it returns
// no results it falls back to V1 REST discovery (immediately consistent) +
// PREvidenceByPRNumber for each PR found, preserving all rich V2 fields.
func (c *GithubConfig) PREvidenceForCommitHybrid(commit string) ([]*types.PREvidence, error) {
	prs, err := c.PREvidenceForCommitV2(commit)
	if err != nil {
		return nil, err
	}
	if len(prs) > 0 {
		return prs, nil
	}

	// V2 returned nothing — fall back to REST discovery.
	restPRs, err := c.PullRequestsForCommit(commit)
	if err != nil {
		return nil, err
	}

	result := []*types.PREvidence{}
	for _, pr := range restPRs {
		evidence, err := c.PREvidenceByPRNumber(pr.GetNumber())
		if err != nil {
			return nil, err
		}
		if evidence != nil {
			result = append(result, evidence)
		}
	}
	return result, nil
}

// graphqlCommitNode is the shared GraphQL node type for commits on a PR,
// used in both PREvidenceByPRNumber and PREvidenceForCommitV2.
type graphqlCommitNode struct {
	Commit struct {
		Oid             graphql.String
		MessageHeadline graphql.String
		CommittedDate   graphql.String
		URL             graphql.String
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

// graphqlReviewNode is the shared GraphQL node type for approved reviews on a PR.
type graphqlReviewNode struct {
	Author struct {
		Login graphql.String
	}
	State       graphql.String
	SubmittedAt graphql.String
}

// buildPREvidence constructs a PREvidence from pre-resolved fields and the
// raw GraphQL commit/review nodes. mergeCommit must be resolved by the caller
// (it differs between commit-SHA queries and PR-number queries).
func buildPREvidence(
	url, mergeCommit, state, author, createdAtStr, mergedAtStr, title, headRef string,
	commitNodes []graphqlCommitNode,
	reviewNodes []graphqlReviewNode,
) (*types.PREvidence, error) {
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	mergedAt := int64(0)
	if mergedAtStr != "" {
		mergedAtTime, err := time.Parse(time.RFC3339, mergedAtStr)
		if err != nil {
			return nil, err
		}
		mergedAt = mergedAtTime.Unix()
	}

	evidence := &types.PREvidence{
		URL:         url,
		MergeCommit: mergeCommit,
		State:       state,
		Author:      author,
		CreatedAt:   createdAt.Unix(),
		MergedAt:    mergedAt,
		Title:       title,
		HeadRef:     headRef,
		Approvers:   []any{},
		Commits:     []types.Commit{},
	}

	for _, n := range commitNodes {
		timestamp, err := time.Parse(time.RFC3339, string(n.Commit.CommittedDate))
		if err != nil {
			return nil, err
		}
		committerUsername := ""
		if n.Commit.Committer.User != nil {
			committerUsername = string(n.Commit.Committer.User.Login)
		}
		evidence.Commits = append(evidence.Commits, types.Commit{
			SHA:               string(n.Commit.Oid),
			Message:           string(n.Commit.MessageHeadline),
			Committer:         fmt.Sprintf("%s <%s>", string(n.Commit.Committer.Name), string(n.Commit.Committer.Email)),
			CommitterUsername: committerUsername,
			Timestamp:         timestamp.Unix(),
			Branch:            headRef,
			URL:               string(n.Commit.URL),
		})
	}

	for _, r := range reviewNodes {
		submittedAt, err := time.Parse(time.RFC3339, string(r.SubmittedAt))
		if err != nil {
			return nil, err
		}
		evidence.Approvers = append(evidence.Approvers, types.PRApprovals{
			Username:  string(r.Author.Login),
			State:     string(r.State),
			Timestamp: submittedAt.Unix(),
		})
	}

	return evidence, nil
}

// PREvidenceByPRNumber fetches full PR evidence for a single PR number via
// GraphQL. Returns an error when the PR does not exist.
func (c *GithubConfig) PREvidenceByPRNumber(prNumber int) (*types.PREvidence, error) {
	ctx := context.Background()

	ghClient, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL)
	if err != nil {
		return nil, err
	}
	httpClient := ghClient.Client()
	client := graphql.NewClient(graphqlEndpoint(c.BaseURL), httpClient)

	var query struct {
		Repository struct {
			PullRequest *struct {
				Title       graphql.String
				State       graphql.String
				HeadRefName graphql.String
				URL         graphql.String
				CreatedAt   graphql.String
				MergedAt    graphql.String
				MergeCommit *struct {
					Oid graphql.String
				}
				Author struct {
					Login graphql.String
				}
				Commits struct {
					Nodes    []graphqlCommitNode
					PageInfo struct {
						HasNextPage graphql.Boolean
						EndCursor   graphql.String
					}
				} `graphql:"commits(first: 100, after: $commitCursor)"`
				Reviews struct {
					Nodes    []graphqlReviewNode
					PageInfo struct {
						HasNextPage graphql.Boolean
						EndCursor   graphql.String
					}
				} `graphql:"reviews(first: 100, states: APPROVED, after: $reviewCursor)"`
			} `graphql:"pullRequest(number: $prNumber)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":        graphql.String(c.Org),
		"repo":         graphql.String(c.Repository),
		"prNumber":     graphql.Int(prNumber),
		"commitCursor": (*graphql.String)(nil),
		"reviewCursor": (*graphql.String)(nil),
	}

	sleep := c.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	delays := []time.Duration{0, 10 * time.Second, 20 * time.Second, 30 * time.Second}
	for _, delay := range delays {
		if delay > 0 {
			sleep(delay)
		}
		err = client.Query(ctx, &query, variables)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	pr := query.Repository.PullRequest
	if pr == nil {
		return nil, nil
	}

	mergeCommit := ""
	if pr.MergeCommit != nil {
		mergeCommit = string(pr.MergeCommit.Oid)
	}

	return buildPREvidence(
		string(pr.URL), mergeCommit, string(pr.State), string(pr.Author.Login),
		string(pr.CreatedAt), string(pr.MergedAt), string(pr.Title), string(pr.HeadRefName),
		pr.Commits.Nodes, pr.Reviews.Nodes,
	)
}

func (c *GithubConfig) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
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
								Nodes    []graphqlCommitNode
								PageInfo struct {
									HasNextPage graphql.Boolean
									EndCursor   graphql.String
								}
							} `graphql:"commits(first: 100, after: $commitCursor)"`

							Reviews struct {
								Nodes    []graphqlReviewNode
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

	for _, pr := range query.Repository.Object.Commit.AssociatedPullRequests.Nodes {
		// MergeCommit is set to the queried commit SHA — V2 queries by commit SHA
		// so the commit is by definition the merge commit.
		evidence, err := buildPREvidence(
			string(pr.URL), commit, string(pr.State), string(pr.Author.Login),
			string(pr.CreatedAt), string(pr.MergedAt), string(pr.Title), string(pr.HeadRefName),
			pr.Commits.Nodes, pr.Reviews.Nodes,
		)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

type GitObjectID string

func (v GitObjectID) MarshalGQL(w io.Writer) {
	if _, err := fmt.Fprintf(w, `"%s"`, string(v)); err != nil {
		// Log warning for output error
		fmt.Printf("warning: failed to write GitObjectID: %v\n", err)
	}
}

func (c *GithubConfig) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
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
	evidence.Approvers = utils.ConvertStringListToInterfaceList(approvers)
	return evidence, nil
}
