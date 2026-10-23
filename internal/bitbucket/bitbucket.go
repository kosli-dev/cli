package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/types"
)

// defaultMaxPages bounds commit paging so next links that cycle end in an
// error. It is set far above the pages a real pull request needs.
const defaultMaxPages = 1000

type Config struct {
	Username    string
	Password    string
	AccessToken string
	Workspace   string
	Repository  string
	Logger      *logger.Logger
	KosliClient *requests.Client
	Assert      bool
}

// parseRFC3339NanoTimestamp parses a timestamp string in RFC3339Nano format and returns its Unix timestamp.
// The fieldName parameter is used for error messages to identify which field failed to parse.
// An empty string gives 0: the field was missing.
func parseRFC3339NanoTimestamp(timestampStr, fieldName string) (int64, error) {
	if timestampStr == "" {
		return 0, nil
	}
	parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s timestamp: %w", fieldName, err)
	}
	return parsedTime.Unix(), nil
}

// str, obj and boolean read a field from a decoded JSON object, giving the
// zero value when it is missing, null or of another type.
func str(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func obj(m map[string]any, key string) map[string]any {
	v, _ := m[key].(map[string]any)
	return v
}

func boolean(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}

// objects reads a list of objects. A missing list is empty, but an item that
// is not an object is an error: it is an entry we can't read.
func objects(m map[string]any, key string) ([]map[string]any, error) {
	items, _ := m[key].([]any)
	out := make([]map[string]any, len(items))
	for i, item := range items {
		o, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s[%d] in the Bitbucket reply is null or not an object", key, i)
		}
		out[i] = o
	}
	return out, nil
}

func (c *Config) ProviderAndLabel() (string, string) {
	return "bitbucket", "pull request"
}

// This is the old implementation, it will be removed after the PR payload is enhanced for Bitbucket
func (c *Config) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	return c.getPullRequestsFromBitbucketApi(commit, 1)
}

func (c *Config) PREvidenceForCommitHybrid(commit string) ([]*types.PREvidence, error) {
	return c.PREvidenceForCommitV2(commit)
}

// This is the new implementation, it will be used for Bitbucket
func (c *Config) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	return c.getPullRequestsFromBitbucketApi(commit, 2)
}

func (c *Config) getPullRequestsFromBitbucketApi(commit string, version int) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}

	url, err := url.JoinPath("https://api.bitbucket.org/2.0/repositories", c.Workspace, c.Repository, "commit", commit, "pullrequests")
	if err != nil {
		return pullRequestsEvidence, err
	}
	c.Logger.Debug("getting pull requests from %s", url)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Username: c.Username,
		Password: c.Password,
		Token:    c.AccessToken,
	}
	response, err := c.KosliClient.Do(reqParams)
	if err != nil {
		return pullRequestsEvidence, err
	}
	switch response.Resp.StatusCode {
	case 200:
		pullRequestsEvidence, err = c.parseBitbucketResponse(commit, response, version)
		if err != nil {
			return pullRequestsEvidence, err
		}
	case 202:
		return pullRequestsEvidence, fmt.Errorf("repository pull requests are still being indexed, please retry")
	case 404:
		return pullRequestsEvidence, fmt.Errorf("repository does not exist or pull requests are not indexed." +
			"Please make sure Pull Request Commit Links app is installed")
	default:
		return pullRequestsEvidence, fmt.Errorf("failed to get pull requests from Bitbucket: %v", response.Body)
	}
	return pullRequestsEvidence, nil
}

// parseBitbucketResponse parses the response from the Bitbucket API and returns the pull requests evidence
func (c *Config) parseBitbucketResponse(commit string, response *requests.HTTPResponse, version int) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	var responseData map[string]any
	err := json.Unmarshal([]byte(response.Body), &responseData)
	if err != nil {
		return pullRequestsEvidence, err
	}
	pullRequests, err := objects(responseData, "values")
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, pr := range pullRequests {
		links := obj(pr, "links")
		apiURL := str(obj(links, "self"), "href")
		if apiURL == "" {
			return pullRequestsEvidence, fmt.Errorf("a pull request in the Bitbucket reply has no links.self.href")
		}
		evidence, err := c.getPullRequestDetailsFromBitbucket(apiURL, str(obj(links, "html"), "href"), commit, version)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}

	return pullRequestsEvidence, nil
}

// getPullRequestDetailsFromBitbucket gets the details of a pull request from the Bitbucket API
func (c *Config) getPullRequestDetailsFromBitbucket(prApiUrl, prHtmlLink, commit string, version int) (*types.PREvidence, error) {
	c.Logger.Debug("getting pull request details for %s", prApiUrl)
	evidence := &types.PREvidence{}

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      prApiUrl,
		Username: c.Username,
		Password: c.Password,
		Token:    c.AccessToken,
	}
	response, err := c.KosliClient.Do(reqParams)
	if err != nil {
		return evidence, err
	}
	if response.Resp.StatusCode == 200 {
		var responseData map[string]any
		err := json.Unmarshal([]byte(response.Body), &responseData)
		if err != nil {
			return evidence, err
		}

		evidence.URL = prHtmlLink
		evidence.MergeCommit = commit
		evidence.State = str(responseData, "state")
		participants, err := objects(responseData, "participants")
		if err != nil {
			return evidence, err
		}
		approvers := []any{}

		if len(participants) > 0 {
			for _, p := range participants {
				if username := approvedBy(p); username != "" {
					if version == 1 {
						approvers = append(approvers, username)
					} else {
						approvalTimestamp, err := parseRFC3339NanoTimestamp(str(p, "participated_on"), "participated_on")
						if err != nil {
							return evidence, err
						}
						approvers = append(approvers, types.PRApprovals{
							Username:  username,
							State:     str(p, "state"),
							Timestamp: approvalTimestamp,
						})
					}
				}
			}
			evidence.Approvers = approvers
		} else {
			c.Logger.Debug("no approvers found")
		}
		if version == 2 {
			evidence.Author = str(obj(responseData, "author"), "display_name")
			createdAt, err := parseRFC3339NanoTimestamp(str(responseData, "created_on"), "created_on")
			if err != nil {
				return evidence, err
			}
			evidence.CreatedAt = createdAt
			mergedAt, err := parseRFC3339NanoTimestamp(str(responseData, "updated_on"), "updated_on")
			if err != nil {
				return evidence, err
			}
			evidence.MergedAt = mergedAt
			evidence.Title = str(responseData, "title")
			evidence.HeadRef = bitbucketBranchName(responseData, "source")
			evidence.BaseRef = bitbucketBranchName(responseData, "destination")

			prID, ok := responseData["id"].(float64)
			if !ok {
				return evidence, fmt.Errorf("the pull request at %s has no id", prApiUrl)
			}
			prCommits, err := c.getPullRequestCommitsFromBitbucket(int(prID))
			if err != nil {
				return evidence, err
			}
			evidence.Commits = prCommits

			evidence.Reviews, err = bitbucketReviews(participants)
			if err != nil {
				return evidence, err
			}
		}
	} else {
		return evidence, fmt.Errorf("failed to get PR details, got HTTP status %d. Please review repository permissions", response.Resp.StatusCode)
	}
	return evidence, nil
}

// approvedBy names who approved in participant p, or gives "" when p is no
// approval. An approval with no identity, as from a deleted account, counts
// as none: approval evidence must name who gave it.
func approvedBy(p map[string]any) string {
	if !boolean(p, "approved") {
		return ""
	}
	return str(obj(p, "user"), "display_name")
}

// bitbucketReviews gives each participant's current decision. A participant
// with no state who has commented is a comment; one who has done neither is
// left out.
func bitbucketReviews(participants []map[string]any) ([]types.PRApprovals, error) {
	reviews := []types.PRApprovals{}
	for _, p := range participants {
		username := str(obj(p, "user"), "display_name")
		participatedOn := str(p, "participated_on")
		var state string
		switch p["state"] {
		case "approved":
			state = types.ReviewApproved
		case "changes_requested":
			state = types.ReviewChangesRequested
		case nil:
			if participatedOn == "" {
				continue
			}
			state = types.ReviewCommented
		default:
			continue
		}
		timestamp, err := parseRFC3339NanoTimestamp(participatedOn, "participated_on")
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, types.PRApprovals{
			Username:  username,
			State:     state,
			Timestamp: timestamp,
		})
	}
	return reviews, nil
}

// bitbucketBranchName extracts a branch name from a Bitbucket PR response for
// the given side: "source" for the head branch, "destination" for the base.
func bitbucketBranchName(prData map[string]any, side string) string {
	return str(obj(obj(prData, side), "branch"), "name")
}

// getPullRequestCommitsFromBitbucket gets the commits of a pull request from the Bitbucket API
func (c *Config) getPullRequestCommitsFromBitbucket(prID int) ([]types.Commit, error) {
	url, err := url.JoinPath("https://api.bitbucket.org/2.0/repositories", c.Workspace, c.Repository, "pullrequests", strconv.Itoa(prID), "commits")
	if err != nil {
		return nil, err
	}
	c.Logger.Debug("getting pull request commits from %s", url)

	allCommits := []types.Commit{}
	currentURL := url

	for pages := 0; currentURL != ""; pages++ {
		if pages == defaultMaxPages {
			return nil, fmt.Errorf("aborting after %d pages of commits for pull request %d", defaultMaxPages, prID)
		}
		reqParams := &requests.RequestParams{
			Method:   http.MethodGet,
			URL:      currentURL,
			Username: c.Username,
			Password: c.Password,
			Token:    c.AccessToken,
		}
		response, err := c.KosliClient.Do(reqParams)
		if err != nil {
			return nil, err
		}
		if response.Resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed to get PR commits, got HTTP status %d", response.Resp.StatusCode)
		}

		var responseData map[string]any
		err = json.Unmarshal([]byte(response.Body), &responseData)
		if err != nil {
			return nil, err
		}

		commits, err := objects(responseData, "values")
		if err != nil {
			return nil, err
		}

		for _, commit := range commits {
			timestamp, err := parseRFC3339NanoTimestamp(str(commit, "date"), "date")
			if err != nil {
				return nil, err
			}
			allCommits = append(allCommits, types.Commit{
				SHA:       str(commit, "hash"),
				Message:   str(commit, "message"),
				Author:    str(obj(commit, "author"), "raw"),
				Timestamp: timestamp,
				URL:       str(obj(obj(commit, "links"), "html"), "href"),
			})
		}

		currentURL = str(responseData, "next")
		if currentURL == "" {
			break
		}
		c.Logger.Debug("fetching next page of commits from %s", currentURL)
	}

	return allCommits, nil
}
