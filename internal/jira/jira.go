package jira

import (
	"fmt"
	"net/http"

	jira "github.com/andygrunwald/go-jira"
)

type JiraConfig struct {
	Username string
	APIToken string // API tokens are used in Jira Cloud
	PAT      string // Personal access tokens are used in self-hosted Jira
	BaseURL  string
}

type JiraIssueInfo struct {
	IssueID     string `json:"issue_id"`
	IssueURL    string `json:"issue_url"`
	IssueExists bool   `json:"issue_exists"`
}

// NewJiraConfig returns a new JiraConfig
func NewJiraConfig(baseURL, username, apiToken, PAT string) *JiraConfig {
	return &JiraConfig{
		Username: username,
		APIToken: apiToken,
		PAT:      PAT,
		BaseURL:  baseURL,
	}
}

func (jc *JiraConfig) NewJiraClient() (*jira.Client, error) {
	var httpClient *http.Client
	if jc.Username != "" && jc.APIToken != "" {
		// Jira docs: https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/
		// Create a new API token: https://id.atlassian.com/manage-profile/security/api-tokens
		tp := jira.BasicAuthTransport{
			Username: jc.Username,
			Password: jc.APIToken,
		}
		httpClient = tp.Client()
	} else if jc.PAT != "" {
		// See "Using Personal Access Tokens"
		// https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html
		tp := jira.BearerAuthTransport{
			Token: jc.PAT,
		}
		httpClient = tp.Client()

	} else {
		return nil, fmt.Errorf("either (username and API token) or personal access token must be provided to create a jira client")
	}

	jiraClient, err := jira.NewClient(httpClient, jc.BaseURL)
	if err != nil {
		return nil, err
	}
	return jiraClient, nil
}

// GetJiraIssueInfo retrieve Jira issue information
// if issue is not found, we still return a JiraIssueInfo object with IssueExists set to false
func (jc *JiraConfig) GetJiraIssueInfo(issueID string) (*JiraIssueInfo, error) {
	result := &JiraIssueInfo{
		IssueID:     issueID,
		IssueExists: false,
		IssueURL:    fmt.Sprintf("%s/browse/%s", jc.BaseURL, issueID),
	}

	jiraClient, err := jc.NewJiraClient()
	if err != nil {
		return result, err
	}
	issue, response, err := jiraClient.Issue.Get(issueID, nil)
	if err != nil && response.StatusCode != http.StatusNotFound {
		return result, err
	}

	if issue != nil {
		result.IssueExists = true
	}
	return result, nil
}
