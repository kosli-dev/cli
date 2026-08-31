package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shurcooL/graphql"
)

const (
	// defaultMaxPages bounds every pagination loop. It is a guard against an
	// API that keeps reporting more results, not a tunable limit: at 100 pages
	// of 100 items it sits far beyond any real pull request.
	defaultMaxPages = 100
	// restPageSize is GitHub's REST maximum. The zero-valued list options used
	// before sent no per_page at all, so the API's default of 30 applied (#1082).
	restPageSize = 100
)

// pageInfo is the cursor block shared by every paginated GraphQL connection.
type pageInfo struct {
	HasNextPage graphql.Boolean
	EndCursor   graphql.String
}

// paginate follows cursors from first onwards, appending each page to seed.
// It errors rather than returning early when the cap is hit or a cursor
// repeats: silently stopping is the truncation bug this fixes (#1082).
//
// The counter starts at 1 because seed is already the first page, so maxPages
// means the same "at most N pages of results" here as in the REST, GitLab and
// Azure loops, which count their first request inside the budget.
func paginate[T any](seed []T, first pageInfo, maxPages int,
	fetch func(after graphql.String) ([]T, pageInfo, error)) ([]T, error) {
	all := seed
	page := first
	for pages := 1; page.HasNextPage; pages++ {
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
		var err error
		for _, delay := range []time.Duration{0, 10 * time.Second, 20 * time.Second, 30 * time.Second} {
			if delay > 0 {
				// A wait only happens after a failed attempt, so err holds the
				// reason every attempt failed. Join keeps it alongside the
				// cancellation, so neither side loses its errors.Is.
				if waitErr := c.wait(ctx, delay); waitErr != nil {
					return errors.Join(err, waitErr)
				}
			}
			if err = client.Query(ctx, q, variables); err == nil {
				return nil
			}
		}
		return err
	}
}

// wait pauses between retries. It honours ctx cancellation so a caller with a
// deadline is not held for the full ladder; c.Sleep bypasses that, and exists
// only so tests need not sleep for real.
func (c *GithubConfig) wait(ctx context.Context, d time.Duration) error {
	if c.Sleep != nil {
		c.Sleep(d)
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// prRef identifies the pull request a follow-up page query must resolve. The
// repo is carried explicitly because a follow-up is a separate query: resolving
// pullRequest(number:) against the configured repo would silently return the
// wrong PR, or none, for any node that is not in it.
type prRef struct {
	Owner  string
	Repo   string
	Number int
}

// owns reports whether ref names the configured repository. GitHub owner and
// repo names are case-insensitive, so a differently-cased config value still
// names the same repo.
func (c *GithubConfig) owns(ref prRef) bool {
	return strings.EqualFold(ref.Owner, c.Org) && strings.EqualFold(ref.Repo, c.Repository)
}

// pageVariables are the query variables shared by both follow-up page queries.
// The cursor is a plain graphql.String — the initial queries pass a nil
// *graphql.String, so they declare `String` where these declare `String!`.
func pageVariables(ref prRef, after graphql.String) map[string]any {
	return map[string]any{
		"owner":    graphql.String(ref.Owner),
		"repo":     graphql.String(ref.Repo),
		"prNumber": graphql.Int(ref.Number),
		"cursor":   after,
	}
}

// allPRCommits returns every commit on prNumber, seeded with the nodes the
// caller's first query already returned. Follow-up pages select only commits,
// so draining one connection never re-fetches the other.
func (c *GithubConfig) allPRCommits(ctx context.Context, run graphqlQueryFunc, ref prRef,
	seed []graphqlCommitNode, first pageInfo) ([]graphqlCommitNode, error) {
	return paginate(seed, first, defaultMaxPages, func(after graphql.String) ([]graphqlCommitNode, pageInfo, error) {
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
		if err := run(ctx, &q, pageVariables(ref, after)); err != nil {
			return nil, pageInfo{}, err
		}
		commits := q.Repository.PullRequest.Commits
		return commits.Nodes, commits.PageInfo, nil
	})
}

// allPRReviews returns every approved review on prNumber, seeded as above.
func (c *GithubConfig) allPRReviews(ctx context.Context, run graphqlQueryFunc, ref prRef,
	seed []graphqlReviewNode, first pageInfo) ([]graphqlReviewNode, error) {
	return paginate(seed, first, defaultMaxPages, func(after graphql.String) ([]graphqlReviewNode, pageInfo, error) {
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
		if err := run(ctx, &q, pageVariables(ref, after)); err != nil {
			return nil, pageInfo{}, err
		}
		reviews := q.Repository.PullRequest.Reviews
		return reviews.Nodes, reviews.PageInfo, nil
	})
}
