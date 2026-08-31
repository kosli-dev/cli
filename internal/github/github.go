package github

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	Debug      bool
	// Sleep replaces the wait between GraphQL retries — every retry, including
	// the follow-up page queries of both PR-evidence entry points. Nil means an
	// interruptible wait that returns as soon as the context is cancelled;
	// setting it bypasses that, so leave it nil to exercise cancellation.
	Sleep func(time.Duration)
}

type GithubFlagsTempValueHolder struct {
	Token      string
	BaseURL    string
	Org        string
	Repository string
}

// NewGithubConfig returns a new GithubConfig
func NewGithubConfig(token, baseURL, org, repository string, debug bool) *GithubConfig {
	return &GithubConfig{
		Token:   token,
		BaseURL: baseURL,
		Org:     org,
		// repository name must be extracted if a user is using default value from ${GITHUB_REPOSITORY}
		// because the value is in the format of "org/repository"
		Repository: extractRepoName(repository),
		Debug:      debug,
	}
}

// extractRepoName returns repository name from 'org/repository_name' string
func extractRepoName(fullRepositoryName string) string {
	repoNameParts := strings.Split(fullRepositoryName, "/")
	repository := repoNameParts[len(repoNameParts)-1]
	return repository
}

// NewGithubClientFromToken returns Github client with a token and context.
// When debug is true the underlying transport is wrapped so every HTTP
// request and response (REST and GraphQL) is dumped to stderr.
//
// debugTransport is installed as oauth2.Transport.Base (not as a wrapper
// around it) so the Authorization header that oauth2 attaches is visible
// in the dump — otherwise debug logs the request before oauth2 adds the
// header and we lose the ability to verify it was actually sent.
func NewGithubClientFromToken(ctx context.Context, ghToken string, baseURL string, debug bool) (*gh.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	if debug {
		t := tc.Transport.(*oauth2.Transport)
		base := t.Base
		if base == nil {
			base = http.DefaultTransport
		}
		t.Base = &debugTransport{base: base, out: os.Stderr}
	}
	if baseURL != "" {
		client, err := gh.NewEnterpriseClient(baseURL, baseURL, tc)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return gh.NewClient(tc), nil
}

// debugTransport is an http.RoundTripper that logs each request and
// response to its writer. Used when --debug is enabled to make GitHub
// auth/permission failures debuggable.
type debugTransport struct {
	base http.RoundTripper
	out  io.Writer
	// proxyFunc resolves the proxy URL for a request. Override in tests;
	// nil means use http.ProxyFromEnvironment, whose cache makes it
	// effectively impossible to set via t.Setenv after process start.
	proxyFunc func(*http.Request) (*url.URL, error)
}

// logf writes debug output and intentionally ignores write errors —
// failures writing to stderr are not actionable from a debug transport.
func (d *debugTransport) logf(format string, args ...any) {
	_, _ = fmt.Fprintf(d.out, format, args...)
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	d.logf("[debug-github] --> %s %s\n", req.Method, req.URL)
	proxyFunc := d.proxyFunc
	if proxyFunc == nil {
		proxyFunc = http.ProxyFromEnvironment
	}
	if proxyURL, proxyErr := proxyFunc(req); proxyErr != nil {
		d.logf("[debug-github]     <proxy lookup error: %v>\n", proxyErr)
	} else if proxyURL != nil {
		d.logf("[debug-github]     <via proxy %s>\n", proxyURL.Redacted())
	}
	for k, vs := range req.Header {
		for _, v := range vs {
			d.logf("[debug-github]     %s: %s\n", k, redactSensitiveHeader(k, v))
		}
	}
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			d.logf("[debug-github]     <failed to read request body: %v>\n", err)
		} else {
			if len(bodyBytes) > 0 {
				d.logf("[debug-github]     body: %s\n", string(bodyBytes))
			}
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	resp, err := d.base.RoundTrip(req)
	if err != nil {
		d.logf("[debug-github] <-- transport error: %v\n", err)
	}
	if resp != nil {
		d.logResponse(resp, req)
	}
	return resp, err
}

func (d *debugTransport) logResponse(resp *http.Response, req *http.Request) {
	d.logf("[debug-github] <-- %d %s (%s %s)\n", resp.StatusCode, resp.Status, req.Method, req.URL)
	for k, vs := range resp.Header {
		for _, v := range vs {
			d.logf("[debug-github]     %s: %s\n", k, redactSensitiveHeader(k, v))
		}
	}
	if resp.Body != nil {
		respBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			d.logf("[debug-github]     <failed to read response body: %v>\n", err)
			resp.Body = io.NopCloser(bytes.NewReader(nil))
		} else {
			if len(respBytes) > 0 {
				d.logf("[debug-github]     body: %s\n", string(respBytes))
			}
			resp.Body = io.NopCloser(bytes.NewReader(respBytes))
		}
	}
}

// redactSensitiveHeader hides credentials in headers that commonly carry
// them so debug output is safe to paste into bug reports. Authorization
// keeps the last 4 chars of the token so the user can spot truncation
// or whitespace issues; cookie/proxy-auth values are fully replaced.
func redactSensitiveHeader(name, value string) string {
	switch strings.ToLower(name) {
	case "authorization":
		return redactAuthHeader(value)
	case "cookie", "set-cookie", "proxy-authorization":
		return "***"
	default:
		return value
	}
}

// redactAuthHeader hides all but the last 4 chars of the credential so
// debug output is safe to paste into bug reports while still letting the
// user spot truncation/whitespace issues.
func redactAuthHeader(v string) string {
	parts := strings.SplitN(v, " ", 2)
	if len(parts) != 2 {
		if len(v) <= 4 {
			return "***"
		}
		return "***" + v[len(v)-4:]
	}
	tok := parts[1]
	if len(tok) <= 4 {
		return parts[0] + " ***"
	}
	return parts[0] + " ***" + tok[len(tok)-4:]
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

func defaultNewGithubRetriever(token, baseURL, org, repository string, debug bool) types.PRRetriever {
	return NewGithubConfig(token, baseURL, org, repository, debug)
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
		AuthoredDate    graphql.String
		URL             graphql.String
		Author          struct {
			Name  graphql.String
			Email graphql.String
			User  *struct {
				Login graphql.String
			}
		}
		Signature *struct {
			IsValid graphql.Boolean
			State   graphql.String
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
	url, mergeCommit, state, author, createdAtStr, mergedAtStr, title, headRef, baseRef string,
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
		BaseRef:     baseRef,
		Approvers:   []any{},
		Commits:     []types.Commit{},
	}

	for _, n := range commitNodes {
		// Use the author date to match the recorded author identity; fall back to
		// the committed date if the API omits it (server#5479).
		dateStr := string(n.Commit.AuthoredDate)
		if dateStr == "" {
			dateStr = string(n.Commit.CommittedDate)
		}
		timestamp, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return nil, err
		}
		authorUsername := ""
		if n.Commit.Author.User != nil {
			authorUsername = string(n.Commit.Author.User.Login)
		}
		// Capture the commit signature when present. A nil signature node means
		// the commit is unsigned, which must stay distinct from a present but
		// invalid signature (verified=false) — so leave the fields nil (server#5892).
		var verified *bool
		var signatureState *string
		if n.Commit.Signature != nil {
			v := bool(n.Commit.Signature.IsValid)
			s := string(n.Commit.Signature.State)
			verified = &v
			signatureState = &s
		}
		evidence.Commits = append(evidence.Commits, types.Commit{
			SHA:            string(n.Commit.Oid),
			Message:        string(n.Commit.MessageHeadline),
			Author:         fmt.Sprintf("%s <%s>", string(n.Commit.Author.Name), string(n.Commit.Author.Email)),
			AuthorUsername: authorUsername,
			Timestamp:      timestamp.Unix(),
			Branch:         headRef,
			URL:            string(n.Commit.URL),
			Verified:       verified,
			SignatureState: signatureState,
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

	ghClient, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL, c.Debug)
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
				BaseRefName graphql.String
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
					PageInfo pageInfo
				} `graphql:"commits(first: 100, after: $commitCursor)"`
				Reviews struct {
					Nodes    []graphqlReviewNode
					PageInfo pageInfo
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

	run := c.queryWithRetry(client)
	if err := run(ctx, &query, variables); err != nil {
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

	ref := prRef{Owner: c.Org, Repo: c.Repository, Number: prNumber}
	commits, err := c.allPRCommits(ctx, run, ref, pr.Commits.Nodes, pr.Commits.PageInfo)
	if err != nil {
		return nil, err
	}
	reviews, err := c.allPRReviews(ctx, run, ref, pr.Reviews.Nodes, pr.Reviews.PageInfo)
	if err != nil {
		return nil, err
	}

	return buildPREvidence(
		string(pr.URL), mergeCommit, string(pr.State), string(pr.Author.Login),
		string(pr.CreatedAt), string(pr.MergedAt), string(pr.Title), string(pr.HeadRefName), string(pr.BaseRefName),
		commits, reviews,
	)
}

func (c *GithubConfig) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	ctx := context.Background()
	pullRequestsEvidence := []*types.PREvidence{}

	ghClient, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL, c.Debug)
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
							Number graphql.Int
							// Follow-up page queries resolve the PR by number, so they
							// need the repo it actually lives in, not the configured one.
							Repository struct {
								Name  graphql.String
								Owner struct {
									Login graphql.String
								}
							}
							Title       graphql.String
							State       graphql.String
							HeadRefName graphql.String
							BaseRefName graphql.String
							URL         graphql.String
							CreatedAt   graphql.String
							MergedAt    graphql.String

							Author struct {
								Login graphql.String
							}

							Commits struct {
								Nodes    []graphqlCommitNode
								PageInfo pageInfo
							} `graphql:"commits(first: 100, after: $commitCursor)"`

							Reviews struct {
								Nodes    []graphqlReviewNode
								PageInfo pageInfo
							} `graphql:"reviews(first: 100, states: APPROVED, after: $reviewCursor)"`
						}
						// Intentionally not paginated, so no cursor is selected: a
						// commit with more than 100 associated PRs is not a realistic
						// case, and draining it would mean nesting a page walk per
						// PR page (#1082).
					} `graphql:"associatedPullRequests(first: 100)"`
				} `graphql:"... on Commit"`
			} `graphql:"object(oid: $commitSHA)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":        graphql.String(c.Org),
		"repo":         graphql.String(c.Repository),
		"commitSHA":    GitObjectID(commit),
		"commitCursor": (*graphql.String)(nil),
		"reviewCursor": (*graphql.String)(nil),
	}

	if err := client.Query(ctx, &query, variables); err != nil {
		return pullRequestsEvidence, err
	}

	// Only the follow-up pages retry; the initial query keeps its fail-fast
	// behaviour, which callers already depend on.
	retrying := c.queryWithRetry(client)
	for _, pr := range query.Repository.Object.Commit.AssociatedPullRequests.Nodes {
		ref := prRef{
			Owner:  string(pr.Repository.Owner.Login),
			Repo:   string(pr.Repository.Name),
			Number: int(pr.Number),
		}
		// A node can live in a repo this token cannot read. That failure is
		// permanent, and it is the likely one here, so cross-repo pages trade
		// the retry ladder away rather than spend 60s reaching it — at the
		// cost of not retrying a genuine blip on those pages either.
		run := retrying
		if !c.owns(ref) {
			run = client.Query
		}
		commits, err := c.allPRCommits(ctx, run, ref, pr.Commits.Nodes, pr.Commits.PageInfo)
		if err != nil {
			return pullRequestsEvidence, err
		}
		reviews, err := c.allPRReviews(ctx, run, ref, pr.Reviews.Nodes, pr.Reviews.PageInfo)
		if err != nil {
			return pullRequestsEvidence, err
		}
		// MergeCommit is set to the queried commit SHA — V2 queries by commit SHA
		// so the commit is by definition the merge commit.
		evidence, err := buildPREvidence(
			string(pr.URL), commit, string(pr.State), string(pr.Author.Login),
			string(pr.CreatedAt), string(pr.MergedAt), string(pr.Title), string(pr.HeadRefName), string(pr.BaseRefName),
			commits, reviews,
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
	client, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL, c.Debug)
	if err != nil {
		return []*gh.PullRequest{}, err
	}

	return c.listPullRequestsForCommit(ctx, client, commit, defaultMaxPages)
}

// listPullRequestsForCommit drains the PRs-for-commit endpoint. maxPages is a
// parameter so the cap is testable without 100 round trips.
func (c *GithubConfig) listPullRequestsForCommit(ctx context.Context, client *gh.Client,
	commit string, maxPages int) ([]*gh.PullRequest, error) {
	// Page starts at 1, not the zero value: at 0 the non-advancing guard below
	// cannot fire on the first response. page=1 is the server default anyway.
	opts := &gh.PullRequestListOptions{ListOptions: gh.ListOptions{PerPage: restPageSize, Page: 1}}
	all := []*gh.PullRequest{}
	for pages := 0; pages < maxPages; pages++ {
		pullrequests, resp, err := client.PullRequests.ListPullRequestsWithCommit(ctx, c.Org, c.Repository,
			commit, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, pullrequests...)
		if resp.NextPage == 0 {
			return all, nil
		}
		if resp.NextPage <= opts.Page {
			return nil, fmt.Errorf("next page %d did not advance past page %d for commit %s",
				resp.NextPage, opts.Page, commit)
		}
		opts.Page = resp.NextPage
	}
	return nil, fmt.Errorf("aborting after %d pages of pull requests for commit %s", maxPages, commit)
}

// GetPullRequestApprovers returns a list of approvers for a given pull request
func (c *GithubConfig) GetPullRequestApprovers(number int) ([]string, error) {
	approvers := []string{}
	ctx := context.Background()
	client, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL, c.Debug)
	if err != nil {
		return approvers, err
	}
	return c.listApprovers(ctx, client, number, defaultMaxPages)
}

// listApprovers drains the reviews endpoint, keeping the logins that approved.
// maxPages is a parameter so the cap is testable without 100 round trips.
//
// ListReviews returns every review event, not just approvals, so the APPROVED
// filter has to run over all pages: an approval past page one was previously
// dropped outright (#1082).
func (c *GithubConfig) listApprovers(ctx context.Context, client *gh.Client,
	number, maxPages int) ([]string, error) {
	approvers := []string{}
	// Page starts at 1 for the same reason as above: the guard needs a floor it
	// can compare against on the first response.
	opts := &gh.ListOptions{PerPage: restPageSize, Page: 1}
	// A reviewer can approve more than once. These entries carry no timestamp,
	// so a repeated login adds nothing and is dropped.
	seen := map[string]bool{}
	for pages := 0; pages < maxPages; pages++ {
		reviews, resp, err := client.PullRequests.ListReviews(ctx, c.Org, c.Repository, number, opts)
		if err != nil {
			return nil, err
		}
		for _, r := range reviews {
			login := r.GetUser().GetLogin()
			if r.GetState() == "APPROVED" && !seen[login] {
				seen[login] = true
				approvers = append(approvers, login)
			}
		}
		if resp.NextPage == 0 {
			return approvers, nil
		}
		if resp.NextPage <= opts.Page {
			return nil, fmt.Errorf("next page %d did not advance past page %d for pull request %d",
				resp.NextPage, opts.Page, number)
		}
		opts.Page = resp.NextPage
	}
	return nil, fmt.Errorf("aborting after %d pages of reviews for pull request %d", maxPages, number)
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
