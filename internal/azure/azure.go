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

// value gives what an SDK reply pointer holds, or the zero value when the
// reply left the field out or null.
func value[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// unixTime gives a reply timestamp in Unix seconds, or 0 when it is missing.
func unixTime(t *azuredevops.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Time.Unix()
}

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

func (c *AzureConfig) ProviderAndLabel() (string, string) {
	return "azure", "pull request"
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

func (c *AzureConfig) PREvidenceForCommitHybrid(commit string) ([]*types.PREvidence, error) {
	return c.PREvidenceForCommitV2(commit)
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
	prID := strconv.Itoa(value(pr.PullRequestId))
	url, err := url.JoinPath(c.OrgURL, c.Project, "_git", c.Repository, "pullrequest", prID)
	if err != nil {
		return nil, err
	}
	evidence := &types.PREvidence{
		URL:         url,
		MergeCommit: value(value(pr.LastMergeCommit).CommitId),
		State:       string(value(pr.Status)),
	}
	evidence.Approvers, err = c.GetPullRequestApprovers(value(pr.PullRequestId), 1)
	if err != nil {
		return evidence, err
	}
	return evidence, nil
}

// newPRAzureEvidenceV2 creates a new PREvidence for a given pull request in V2 format
func (c *AzureConfig) newPRAzureEvidenceV2(pr git.GitPullRequest) (*types.PREvidence, error) {
	prID := strconv.Itoa(value(pr.PullRequestId))
	prURL, err := url.JoinPath(c.OrgURL, c.Project, "_git", c.Repository, "pullrequest", prID)
	if err != nil {
		return nil, err
	}

	evidence := &types.PREvidence{
		URL:         prURL,
		MergeCommit: value(value(pr.LastMergeCommit).CommitId),
		State:       string(value(pr.Status)),
		Author:      fmt.Sprintf("%s (%s)", value(value(pr.CreatedBy).DisplayName), value(value(pr.CreatedBy).UniqueName)),
		CreatedAt:   unixTime(pr.CreationDate),
		Title:       value(pr.Title),
	}
	evidence.HeadRef, evidence.BaseRef = azurePRRefs(pr)
	if value(pr.Status) == git.PullRequestStatusValues.Completed {
		evidence.MergedAt = unixTime(pr.ClosedDate)
	}
	commits, err := c.GetPullRequestCommits(pr)
	if err != nil {
		return evidence, err
	}
	evidence.Commits = commits
	evidence.Approvers, evidence.Reviews, err = c.GetPullRequestReviews(value(pr.PullRequestId))
	if err != nil {
		return evidence, err
	}
	return evidence, nil
}

// azurePRRefs returns the head (source) and base (target) ref names of an Azure
// DevOps pull request, kept raw (e.g. "refs/heads/main") to match head_ref.
func azurePRRefs(pr git.GitPullRequest) (head, base string) {
	return value(pr.SourceRefName), value(pr.TargetRefName)
}

// GetPullRequestCommits returns a list of commits for a given pull request
func (c *AzureConfig) GetPullRequestCommits(pr git.GitPullRequest) ([]types.Commit, error) {
	commits := []types.Commit{}
	// A nil id is the one input the SDK rejects outright, so it can only be
	// named here, not read from the reply.
	prID := "with no id"
	if pr.PullRequestId != nil {
		prID = strconv.Itoa(*pr.PullRequestId)
	}

	ctx := context.Background()
	client, err := newPRCommitsClient(ctx, c.Token, c.OrgURL)
	if err != nil {
		return commits, err
	}

	prCommits, err := fetchAzurePRCommits(ctx, client, git.GetPullRequestCommitsArgs{
		RepositoryId:  &c.Repository,
		PullRequestId: pr.PullRequestId,
		Project:       &c.Project,
	}, defaultMaxPages)
	if err != nil {
		return commits, fmt.Errorf("draining commits for pull request %s: %w", prID, err)
	}

	for i, commit := range prCommits {
		// A null commit decodes to an empty one, so no id is the only sign.
		if commit.CommitId == nil {
			return commits, fmt.Errorf("commit %d of pull request %s has no commitId", i, prID)
		}
		commits = append(commits, commitFromAzureCommit(commit, value(pr.SourceRefName)))
	}

	return commits, nil
}

// commitFromAzureCommit maps an Azure DevOps API commit to a types.Commit.
func commitFromAzureCommit(commit git.GitCommitRef, branch string) types.Commit {
	author := value(commit.Author)
	return types.Commit{
		SHA:       value(commit.CommitId),
		Message:   value(commit.Comment),
		Author:    fmt.Sprintf("%s <%s>", value(author.Name), value(author.Email)),
		Timestamp: unixTime(author.Date),
		URL:       value(commit.Url),
		Branch:    branch,
	}
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
	results := value(value(prQuery).Results)
	if len(results) == 0 {
		return []git.GitPullRequest{}, nil
	}
	// A null result indexes as an empty map, so it reads as no pull requests.
	prs := results[0][commit]
	for i, pr := range prs {
		// A null pull request decodes to an empty one, so no id is the only sign.
		if pr.PullRequestId == nil {
			return nil, fmt.Errorf("pull request %d for commit %s has no pullRequestId", i, commit)
		}
	}
	return prs, nil
}

// GetPullRequestApprovers returns a list of approvers for a given pull request
func (c *AzureConfig) GetPullRequestApprovers(prNumber, version int) ([]any, error) {
	reviewers, err := c.pullRequestReviewers(prNumber)
	if err != nil {
		return nil, err
	}
	return azureApprovers(reviewers, version), nil
}

// GetPullRequestReviews returns a PR's approvers and its reviews: each
// reviewer's current vote, since Azure keeps no history of them.
func (c *AzureConfig) GetPullRequestReviews(prNumber int) ([]any, []types.PRApprovals, error) {
	reviewers, err := c.pullRequestReviewers(prNumber)
	if err != nil {
		return nil, nil, err
	}
	return azureApprovers(reviewers, 2), azureReviews(reviewers), nil
}

func (c *AzureConfig) pullRequestReviewers(prNumber int) ([]git.IdentityRefWithVote, error) {
	ctx := context.Background()
	client, err := NewAzureClientFromToken(ctx, c.Token, c.OrgURL)
	if err != nil {
		return nil, err
	}
	reviewers, err := client.GetPullRequestReviewers(ctx, git.GetPullRequestReviewersArgs{
		RepositoryId:  &c.Repository,
		PullRequestId: &prNumber,
		Project:       &c.Project,
	})
	if err != nil {
		return nil, err
	}
	return value(reviewers), nil
}

// azureReviewerName gives "" for a reviewer with no identity, such as a
// deleted account. Its vote is never counted as an approval in approvers or
// reviews.
func azureReviewerName(r git.IdentityRefWithVote) string {
	if value(r.DisplayName) == "" && value(r.UniqueName) == "" {
		return ""
	}
	return fmt.Sprintf("%s (%s)", value(r.DisplayName), value(r.UniqueName))
}

func azureApprovers(reviewers []git.IdentityRefWithVote, version int) []any {
	var approvers []any
	for _, r := range reviewers {
		if approverName := azureReviewerName(r); value(r.Vote) == 10 && approverName != "" {
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
	return approvers
}

// azureReviews leaves out groups: a group's vote is its members' votes, and
// each member who voted is listed in their own right.
func azureReviews(reviewers []git.IdentityRefWithVote) []types.PRApprovals {
	reviews := []types.PRApprovals{}
	for _, r := range reviewers {
		if value(r.IsContainer) {
			continue
		}
		var state string
		switch vote := value(r.Vote); {
		case vote > 0:
			state = types.ReviewApproved
		case vote < 0:
			state = types.ReviewChangesRequested
		default:
			continue
		}
		name := azureReviewerName(r)
		// A change request is kept without a name: dropping it would read as
		// none given.
		if name == "" && state != types.ReviewChangesRequested {
			continue
		}
		reviews = append(reviews, types.PRApprovals{Username: name, State: state})
	}
	return reviews
}
