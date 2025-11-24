package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/types"
)

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
func parseRFC3339NanoTimestamp(timestampStr, fieldName string) (int64, error) {
	parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s timestamp: %w", fieldName, err)
	}
	return parsedTime.Unix(), nil
}

// This is the old implementation, it will be removed after the PR payload is enhanced for Bitbucket
func (c *Config) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	return c.getPullRequestsFromBitbucketApi(commit, 1)
}

// This is the new implementation, it will be used for Bitbucket
func (c *Config) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	return c.getPullRequestsFromBitbucketApi(commit, 2)
}

func (c *Config) getPullRequestsFromBitbucketApi(commit string, version int) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/commit/%s/pullrequests", c.Workspace, c.Repository, commit)
	c.Logger.Debug("getting pull requests from " + url)

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
	pullRequests, ok := responseData["values"].([]any)
	if !ok {
		return pullRequestsEvidence, nil
	}
	for _, prInterface := range pullRequests {
		pr := prInterface.(map[string]any)
		linksInterface := pr["links"].(map[string]any)
		apiLinkMap := linksInterface["self"].(map[string]any)
		htmlLinkMap := linksInterface["html"].(map[string]any)
		evidence, err := c.getPullRequestDetailsFromBitbucket(apiLinkMap["href"].(string), htmlLinkMap["href"].(string), commit, version)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}

	return pullRequestsEvidence, nil
}

// getPullRequestDetailsFromBitbucket gets the details of a pull request from the Bitbucket API
func (c *Config) getPullRequestDetailsFromBitbucket(prApiUrl, prHtmlLink, commit string, version int) (*types.PREvidence, error) {
	c.Logger.Debug("getting pull request details for " + prApiUrl)
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
		evidence.State = responseData["state"].(string)
		participants := responseData["participants"].([]any)
		approvers := []any{}

		if len(participants) > 0 {
			for _, participantInterface := range participants {
				p := participantInterface.(map[string]any)
				if p["approved"].(bool) {
					user := p["user"].(map[string]any)
					if version == 1 {
						approvers = append(approvers, user["display_name"].(string))
					} else {
						approvalTimestamp, err := parseRFC3339NanoTimestamp(p["participated_on"].(string), "participated_on")
						if err != nil {
							return evidence, err
						}
						approvers = append(approvers, types.PRApprovals{
							Username:  user["display_name"].(string),
							State:     p["state"].(string),
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
			evidence.Author = responseData["author"].(map[string]any)["display_name"].(string)
			createdAt, err := parseRFC3339NanoTimestamp(responseData["created_on"].(string), "created_on")
			if err != nil {
				return evidence, err
			}
			evidence.CreatedAt = createdAt
			mergedAt, err := parseRFC3339NanoTimestamp(responseData["updated_on"].(string), "updated_on")
			if err != nil {
				return evidence, err
			}
			evidence.MergedAt = mergedAt
			evidence.Title = responseData["title"].(string)
			evidence.HeadRef = responseData["source"].(map[string]any)["branch"].(map[string]any)["name"].(string)

			prCommits, err := c.getPullRequestCommitsFromBitbucket(int(responseData["id"].(float64)))
			if err != nil {
				return evidence, err
			}
			evidence.Commits = prCommits
		}
	} else {
		return evidence, fmt.Errorf("failed to get PR details, got HTTP status %d. Please review repository permissions", response.Resp.StatusCode)
	}
	return evidence, nil
}

// getPullRequestCommitsFromBitbucket gets the commits of a pull request from the Bitbucket API
func (c *Config) getPullRequestCommitsFromBitbucket(prID int) ([]types.Commit, error) {
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/commits", c.Workspace, c.Repository, prID)
	c.Logger.Debug("getting pull request commits from " + url)

	allCommits := []types.Commit{}
	currentURL := url

	for currentURL != "" {
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

		commits, ok := responseData["values"].([]any)
		if !ok {
			break
		}

		for _, commitInterface := range commits {
			commit := commitInterface.(map[string]any)
			timestamp, err := parseRFC3339NanoTimestamp(commit["date"].(string), "date")
			if err != nil {
				return nil, err
			}
			allCommits = append(allCommits, types.Commit{
				SHA:       commit["hash"].(string),
				Message:   commit["message"].(string),
				Committer: commit["author"].(map[string]any)["raw"].(string),
				Timestamp: timestamp,
				URL:       commit["links"].(map[string]any)["html"].(map[string]any)["href"].(string),
			})
		}

		// Check for next page
		nextInterface, hasNext := responseData["next"]
		if !hasNext {
			break
		}
		nextURL, ok := nextInterface.(string)
		if !ok || nextURL == "" {
			break
		}
		currentURL = nextURL
		c.Logger.Debug("fetching next page of commits from " + currentURL)
	}

	return allCommits, nil
}
