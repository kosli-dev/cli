package jira

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	jira "github.com/andygrunwald/go-jira"
)

type JiraConfig struct {
	Username string
	APIToken string // API tokens are used in Jira Cloud
	PAT      string // Personal access tokens are used in self-hosted Jira
	BaseURL  string
}

type JiraIssueInfo struct {
	IssueID     string            `json:"issue_id"`
	IssueURL    string            `json:"issue_url"`
	IssueExists bool              `json:"issue_exists"`
	IssueFields *jira.IssueFields `json:"issue_fields,omitempty"`
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
func (jc *JiraConfig) GetJiraIssueInfo(issueID string, issueFields string) (*JiraIssueInfo, error) {
	issueUrl, err := url.Parse(jc.BaseURL)
	if err != nil {
		return nil, err
	}
	issueUrl = issueUrl.JoinPath("browse", issueID)

	result := &JiraIssueInfo{
		IssueID:     issueID,
		IssueExists: false,
		IssueURL:    issueUrl.String(),
	}

	jiraClient, err := jc.NewJiraClient()
	if err != nil {
		return result, err
	}

	// API will return all fields if the Fields is empty so we default to a non-existing field.
	// The user can use '*all' if they want all
	if issueFields == "" {
		issueFields = "non-existing-key-in-jira-fields"
	}
	queryOptions := jira.GetQueryOptions{
		Fields: issueFields,
	}

	issue, response, err := jiraClient.Issue.Get(issueID, &queryOptions)
	if err != nil && response != nil && response.StatusCode != http.StatusNotFound {
		return result, err
	}

	if issue != nil {
		result.IssueExists = true
		if issue.Fields != nil {
			result.IssueFields = issue.Fields
		}
	}
	return result, nil
}

func MakeJiraIssueKeyPattern(projectKeys []string) string {
	// Jira issue keys consist of [project-key]-[sequential-number]
	// project key must be at least 2 characters long and start with an uppercase letter
	// more info: https://support.atlassian.com/jira-software-cloud/docs/what-is-an-issue/#Workingwithissues-Projectandissuekeys
	if len(projectKeys) == 0 {
		return `[A-Z][A-Z0-9]{1,9}-[0-9]+`
	} else {
		return `(` + strings.Join(projectKeys, "|") + `)-[0-9]+`
	}
}

// FindJiraIssueKeys finds all Jira issue keys in text, filtering out
// partial matches from multi-segment identifiers like CVE-2026-41284.
// A match is discarded if every occurrence in text is immediately
// followed by "-<digit>".
func FindJiraIssueKeys(text string, projectKeys []string) []string {
	pattern := MakeJiraIssueKeyPattern(projectKeys)
	re := regexp.MustCompile(pattern)
	candidates := re.FindAllString(text, -1)

	// Deduplicate
	seen := make(map[string]struct{})
	var unique []string
	for _, c := range candidates {
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			unique = append(unique, c)
		}
	}

	// Filter out matches that are always followed by -<digit> in the text
	dashDigit := regexp.MustCompile(`^-\d`)
	var result []string
	for _, m := range unique {
		if isPartialMultiSegment(text, m, dashDigit) {
			continue
		}
		result = append(result, m)
	}

	sort.Strings(result)
	if len(result) == 0 {
		return nil
	}
	return result
}

// isPartialMultiSegment returns true if every occurrence of match in text
// is immediately followed by a "-<digit>" suffix, indicating it is part
// of a longer multi-segment identifier (e.g. CVE-2026-41284).
// Precondition: match must exist in text (guaranteed when called from FindJiraIssueKeys).
func isPartialMultiSegment(text, match string, dashDigit *regexp.Regexp) bool {
	start := 0
	for {
		idx := strings.Index(text[start:], match)
		if idx < 0 {
			break
		}
		afterIdx := start + idx + len(match)
		if afterIdx >= len(text) || !dashDigit.MatchString(text[afterIdx:]) {
			return false
		}
		start = start + idx + 1
	}
	return true
}
