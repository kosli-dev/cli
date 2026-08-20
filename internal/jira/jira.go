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

const defaultJiraIssueKeyPattern = `\b[A-Z][A-Z0-9]{1,9}-[0-9]+`

var (
	// compiled once, as the default pattern is a constant
	defaultJiraIssueKeyRegexp = regexp.MustCompile(defaultJiraIssueKeyPattern)
	// dashDigitRegexp is compiled once and shared by all isPartialMultiSegment calls
	dashDigitRegexp = regexp.MustCompile(`^-\d`)
)

func MakeJiraIssueKeyPattern(projectKeys []string) string {
	// Jira issue keys consist of [project-key]-[sequential-number].
	// FindJiraIssueKeys uppercases the text before applying this pattern, so the
	// pattern only needs to handle uppercase. Project keys supplied by the caller
	// are also uppercased here for the same reason.
	// more info: https://support.atlassian.com/jira-software-cloud/docs/what-is-an-issue/#Workingwithissues-Projectandissuekeys
	if len(projectKeys) == 0 {
		return defaultJiraIssueKeyPattern
	}
	upper := make([]string, len(projectKeys))
	for i, k := range projectKeys {
		upper[i] = strings.ToUpper(k)
	}
	return `\b(` + strings.Join(upper, "|") + `)-[0-9]+`
}

// jiraIssueKeyRegexp returns the compiled issue key pattern for the given project keys.
// The default pattern is compiled once; a project-key pattern is compiled per call, which
// is safe from panics because validateJiraProjectKeys has already rejected any key outside
// ^[A-Za-z][A-Za-z0-9_]{1,9}$, so a key cannot carry regex metacharacters.
func jiraIssueKeyRegexp(projectKeys []string) *regexp.Regexp {
	if len(projectKeys) == 0 {
		return defaultJiraIssueKeyRegexp
	}
	return regexp.MustCompile(MakeJiraIssueKeyPattern(projectKeys))
}

// FindJiraIssueKeys finds all Jira issue keys in text, filtering out
// partial matches from multi-segment identifiers like CVE-2026-41284.
// Matching is case-insensitive: the text is uppercased before the regex
// is applied, so all returned keys are in canonical uppercase form.
// A match is discarded if every occurrence in the uppercased text is
// immediately followed by a hyphen and a digit.
func FindJiraIssueKeys(text string, projectKeys []string) []string {
	upperText := strings.ToUpper(text)
	re := jiraIssueKeyRegexp(projectKeys)
	candidates := re.FindAllString(upperText, -1)

	// Deduplicate (all candidates are already uppercase).
	seen := make(map[string]struct{})
	var unique []string
	for _, c := range candidates {
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			unique = append(unique, c)
		}
	}

	// Filter out matches that are always followed by -<digit> in the uppercased text.
	var result []string
	for _, m := range unique {
		if isPartialMultiSegment(upperText, m) {
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
func isPartialMultiSegment(text, match string) bool {
	start := 0
	for {
		idx := strings.Index(text[start:], match)
		if idx < 0 {
			break
		}
		afterIdx := start + idx + len(match)
		if afterIdx >= len(text) || !dashDigitRegexp.MatchString(text[afterIdx:]) {
			return false
		}
		start = start + idx + 1
	}
	return true
}
