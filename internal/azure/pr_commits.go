package azure

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

const (
	// defaultMaxPages bounds every pagination loop. It guards against an API
	// that keeps reporting more results; it is not a tunable limit.
	defaultMaxPages = 100
	// azurePageSize is the $top sent per request. Nothing was sent before, so
	// only Azure's first page was ever read (#1082).
	azurePageSize = 100
)

// prCommitsClient is the slice of git.Client that PR commit retrieval needs,
// narrowed so tests can drive the continuation-token loop with a fake.
type prCommitsClient interface {
	GetPullRequestCommits(context.Context, git.GetPullRequestCommitsArgs) (*git.GetPullRequestCommitsResponseValue, error)
}

// newPRCommitsClient builds the client used by GetPullRequestCommits.
// Replaced in tests, restored with t.Cleanup.
var newPRCommitsClient = func(ctx context.Context, token, orgURL string) (prCommitsClient, error) {
	return NewAzureClientFromToken(ctx, token, orgURL)
}

// fetchAzurePRCommits drains the PR commits endpoint. Azure returns at most
// $top commits per call plus an x-ms-continuationtoken; ignoring that token
// silently truncated the commit list (#1082). A cap or a repeated token errors
// rather than returning early, since quietly stopping is the bug being fixed.
func fetchAzurePRCommits(ctx context.Context, client prCommitsClient,
	args git.GetPullRequestCommitsArgs, maxPages int) ([]git.GitCommitRef, error) {
	all := []git.GitCommitRef{}
	top := azurePageSize
	args.Top = &top
	for pages := 0; pages < maxPages; pages++ {
		response, err := client.GetPullRequestCommits(ctx, args)
		if err != nil {
			return nil, err
		}
		all = append(all, response.Value...)
		if response.ContinuationToken == "" {
			return all, nil
		}
		if args.ContinuationToken != nil && *args.ContinuationToken == response.ContinuationToken {
			return nil, fmt.Errorf("continuation token %q did not advance", response.ContinuationToken)
		}
		token := response.ContinuationToken
		args.ContinuationToken = &token
	}
	return nil, fmt.Errorf("aborting after %d pages of pull request commits", maxPages)
}
