package github

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/graphql"
)

// defaultMaxPages bounds pagination when a caller does not set MaxPages.
const defaultMaxPages = 100

// pageInfo is the cursor block shared by every paginated GraphQL connection.
type pageInfo struct {
	HasNextPage graphql.Boolean
	EndCursor   graphql.String
}

// maxPages resolves the configured page cap, falling back to the default.
func (c *GithubConfig) maxPages() int {
	if c.MaxPages > 0 {
		return c.MaxPages
	}
	return defaultMaxPages
}

// paginate follows cursors from first onwards, appending each page to seed.
// It errors rather than returning early when the cap is hit or a cursor
// repeats: silently stopping is the truncation bug this fixes (#1082).
func paginate[T any](seed []T, first pageInfo, maxPages int,
	fetch func(after graphql.String) ([]T, pageInfo, error)) ([]T, error) {
	all := seed
	page := first
	for pages := 0; page.HasNextPage; pages++ {
		if pages >= maxPages {
			return nil, fmt.Errorf("aborting after %d pages: the API keeps reporting more results", maxPages)
		}
		cursor := page.EndCursor
		nodes, next, err := fetch(cursor)
		if err != nil {
			return nil, err
		}
		if next.HasNextPage && next.EndCursor == cursor {
			return nil, fmt.Errorf("pagination cursor %q did not advance", string(cursor))
		}
		all = append(all, nodes...)
		page = next
	}
	return all, nil
}

// graphqlQueryFunc runs one GraphQL round trip.
type graphqlQueryFunc func(ctx context.Context, q any, variables map[string]any) error

// queryWithRetry runs a GraphQL query, retrying transient failures. Follow-up
// pages use it too: one blip mid-drain would otherwise discard a PR's whole
// commit list (#1082).
func (c *GithubConfig) queryWithRetry(client *graphql.Client) graphqlQueryFunc {
	return func(ctx context.Context, q any, variables map[string]any) error {
		sleep := c.Sleep
		if sleep == nil {
			sleep = time.Sleep
		}
		var err error
		for _, delay := range []time.Duration{0, 10 * time.Second, 20 * time.Second, 30 * time.Second} {
			if delay > 0 {
				sleep(delay)
			}
			if err = client.Query(ctx, q, variables); err == nil {
				return nil
			}
		}
		return err
	}
}

// pageVariables are the query variables shared by both follow-up page queries.
// The cursor is a plain graphql.String — the initial queries pass a nil
// *graphql.String, so they declare `String` where these declare `String!`.
func (c *GithubConfig) pageVariables(prNumber int, after graphql.String) map[string]any {
	return map[string]any{
		"owner":    graphql.String(c.Org),
		"repo":     graphql.String(c.Repository),
		"prNumber": graphql.Int(prNumber),
		"cursor":   after,
	}
}

// allPRCommits returns every commit on prNumber, seeded with the nodes the
// caller's first query already returned. Follow-up pages select only commits,
// so draining one connection never re-fetches the other.
func (c *GithubConfig) allPRCommits(ctx context.Context, run graphqlQueryFunc, prNumber int,
	seed []graphqlCommitNode, first pageInfo) ([]graphqlCommitNode, error) {
	return paginate(seed, first, c.maxPages(), func(after graphql.String) ([]graphqlCommitNode, pageInfo, error) {
		var q struct {
			Repository struct {
				PullRequest struct {
					Commits struct {
						Nodes    []graphqlCommitNode
						PageInfo pageInfo
					} `graphql:"commits(first: 100, after: $cursor)"`
				} `graphql:"pullRequest(number: $prNumber)"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}
		if err := run(ctx, &q, c.pageVariables(prNumber, after)); err != nil {
			return nil, pageInfo{}, err
		}
		commits := q.Repository.PullRequest.Commits
		return commits.Nodes, commits.PageInfo, nil
	})
}

// allPRReviews returns every approved review on prNumber, seeded as above.
func (c *GithubConfig) allPRReviews(ctx context.Context, run graphqlQueryFunc, prNumber int,
	seed []graphqlReviewNode, first pageInfo) ([]graphqlReviewNode, error) {
	return paginate(seed, first, c.maxPages(), func(after graphql.String) ([]graphqlReviewNode, pageInfo, error) {
		var q struct {
			Repository struct {
				PullRequest struct {
					Reviews struct {
						Nodes    []graphqlReviewNode
						PageInfo pageInfo
					} `graphql:"reviews(first: 100, states: APPROVED, after: $cursor)"`
				} `graphql:"pullRequest(number: $prNumber)"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}
		if err := run(ctx, &q, c.pageVariables(prNumber, after)); err != nil {
			return nil, pageInfo{}, err
		}
		reviews := q.Repository.PullRequest.Reviews
		return reviews.Nodes, reviews.PageInfo, nil
	})
}

// restPageSize is GitHub's REST maximum. The zero-valued list options used
// before sent no per_page at all, so the API's default of 30 applied (#1082).
const restPageSize = 100
