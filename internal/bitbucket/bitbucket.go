package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/types"
)

type Config struct {
	Username    string
	Password    string
	Workspace   string
	Repository  string
	Logger      *logger.Logger
	KosliClient *requests.Client
	Assert      bool
}

func (c *Config) PREvidenceForCommit(commit string) ([]*types.PREvidence, error) {
	return c.getPullRequestsFromBitbucketApi(commit, c.Assert)
}

func (c *Config) getPullRequestsFromBitbucketApi(commit string, assert bool) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/commit/%s/pullrequests", c.Workspace, c.Repository, commit)
	c.Logger.Debug("getting pull requests from " + url)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Username: c.Username,
		Password: c.Password,
	}
	response, err := c.KosliClient.Do(reqParams)
	if err != nil {
		return pullRequestsEvidence, err
	}
	if response.Resp.StatusCode == 200 {
		pullRequestsEvidence, err = c.parseBitbucketResponse(commit, response, assert)
		if err != nil {
			return pullRequestsEvidence, err
		}
	} else if response.Resp.StatusCode == 202 {
		return pullRequestsEvidence, fmt.Errorf("repository pull requests are still being indexed, please retry")
	} else if response.Resp.StatusCode == 404 {
		return pullRequestsEvidence, fmt.Errorf("repository does not exist or pull requests are not indexed." +
			"Please make sure Pull Request Commit Links app is installed")
	} else {
		return pullRequestsEvidence, fmt.Errorf("failed to get pull requests from Bitbucket: %v", response.Body)
	}
	return pullRequestsEvidence, nil
}

func (c *Config) parseBitbucketResponse(commit string, response *requests.HTTPResponse, assert bool) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	var responseData map[string]interface{}
	err := json.Unmarshal([]byte(response.Body), &responseData)
	if err != nil {
		return pullRequestsEvidence, err
	}
	pullRequests, ok := responseData["values"].([]interface{})
	if !ok {
		return pullRequestsEvidence, nil
	}
	for _, prInterface := range pullRequests {
		pr := prInterface.(map[string]interface{})
		linksInterface := pr["links"].(map[string]interface{})
		apiLinkMap := linksInterface["self"].(map[string]interface{})
		htmlLinkMap := linksInterface["html"].(map[string]interface{})
		evidence, err := c.getPullRequestDetailsFromBitbucket(apiLinkMap["href"].(string), htmlLinkMap["href"].(string), commit)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	if len(pullRequestsEvidence) == 0 {
		if assert {
			return pullRequestsEvidence, fmt.Errorf("no pull requests found for the given commit: %s", commit)
		}
	}
	return pullRequestsEvidence, nil
}

func (c *Config) getPullRequestDetailsFromBitbucket(prApiUrl, prHtmlLink, commit string) (*types.PREvidence, error) {
	c.Logger.Debug("getting pull request details for " + prApiUrl)
	evidence := &types.PREvidence{}

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      prApiUrl,
		Username: c.Username,
		Password: c.Password,
	}
	response, err := c.KosliClient.Do(reqParams)
	if err != nil {
		return evidence, err
	}
	if response.Resp.StatusCode == 200 {
		var responseData map[string]interface{}
		err := json.Unmarshal([]byte(response.Body), &responseData)
		if err != nil {
			return evidence, err
		}

		evidence.URL = prHtmlLink
		evidence.MergeCommit = commit
		evidence.State = responseData["state"].(string)
		participants := responseData["participants"].([]interface{})
		approvers := []string{}

		if len(participants) > 0 {
			for _, participantInterface := range participants {
				p := participantInterface.(map[string]interface{})
				if p["approved"].(bool) {
					user := p["user"].(map[string]interface{})
					approvers = append(approvers, user["display_name"].(string))
				}
			}
		} else {
			c.Logger.Debug("no approvers found")
		}
		evidence.Approvers = approvers
		// prID := int(responseData["id"].(float64))
		// evidence.LastCommit, evidence.LastCommitter, err = getBitbucketPRLastCommit(workspace, repository, username, password, prID)
		// if err != nil {
		// 	return evidence, err
		// }
		// if utils.Contains(approvers, evidence.LastCommitter) {
		// 	evidence.SelfApproved = true
		// }
	} else {
		return evidence, fmt.Errorf("failed to get PR details, got HTTP status %d. Please review repository permissions", response.Resp.StatusCode)
	}
	return evidence, nil
}

// func getBitbucketPRLastCommit(workspace, repository, username, password string, prID int) (string, string, error) {
// 	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/commits", workspace, repository, prID)
// 	log.Debug("Getting pull requests commits from" + url)
// 	response, err := requests.SendPayload([]byte{}, url, username, password,
// 		global.MaxAPIRetries, false, http.MethodGet, log)
// 	if err != nil {
// 		return "", "", err
// 	}

// 	if response.Resp.StatusCode == 200 {
// 		var responseData map[string]interface{}
// 		err := json.Unmarshal([]byte(response.Body), &responseData)
// 		if err != nil {
// 			return "", "", err
// 		}
// 		prCommits := responseData["values"].([]interface{})

// 		// the first commit is the merge commit
// 		// TODO: is it safe to always to get the second commit?
// 		lastCommit := prCommits[1].(map[string]interface{})
// 		lastAuthor := lastCommit["author"].(map[string]interface{})
// 		return lastCommit["hash"].(string), lastAuthor["user"].(map[string]interface{})["display_name"].(string), nil

// 	} else {
// 		return "", "", fmt.Errorf("failed to get PR commits, got HTTP status %d", response.Resp.StatusCode)
// 	}
// }
