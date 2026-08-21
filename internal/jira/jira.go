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

	// LookupFailure is set when Jira never answered the lookup at all, so IssueExists is
	// false because nothing was learned rather than because the issue is absent. It is
	// reported rather than returned as an error, so that an unreachable Jira cannot fail a
	// run that would otherwise have passed - the exit code stays what it was before this
	// distinction existed. Excluded from the payload: it describes this attempt, not the
	// issue, and the cause stays wrapped so a caller can still tell a timeout from a
	// refused connection.
	LookupFailure error `json:"-"`
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

// credentialDescription names the credential the client is using, so that an
// authentication or permission failure says which one to go and check.
func (jc *JiraConfig) credentialDescription() string {
	if jc.Username != "" && jc.APIToken != "" {
		return fmt.Sprintf("API token of user %s", jc.Username)
	}
	return "personal access token"
}

// VerifyCredentials checks that Jira accepts the configured credential, by asking it
// who we are. It returns nil if Jira accepted it, and otherwise an error saying what to
// go and look at.
//
// It exists because a 404 from the issue endpoint is ambiguous: Jira will not confirm
// that an issue exists to a caller who may not see it, so it answers "not found" both
// for an issue that is absent and for a credential that has expired or lost access. A
// caller that has just been told an issue is missing can use this to explain which of the
// two it is probably looking at.
//
// The result is diagnostic, not a verdict: every outcome here is something the caller
// reports, never something it fails on, so the statuses differ only in wording. 401 and
// 403 name the credential, because Jira saw it and said no. A 404 names the base URL as
// well, because everything that realistically answers /myself that way is shaped like a
// misconfiguration rather than a credential - a base URL that is not a Jira, one missing a
// Data Center context path, a proxy that 404s paths it does not recognise.
func (jc *JiraConfig) VerifyCredentials() error {
	jiraClient, err := jc.NewJiraClient()
	if err != nil {
		return err
	}

	_, response, err := jiraClient.User.GetSelf()
	if err == nil {
		return nil
	}
	if response == nil {
		return fmt.Errorf("failed to reach Jira at %s: %w", jc.BaseURL, err)
	}
	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("the %s was not accepted by Jira at %s (HTTP %d)", jc.credentialDescription(), jc.BaseURL, response.StatusCode)
	case http.StatusNotFound:
		return fmt.Errorf("could not check the %s at %s: verify the credential, and that --jira-base-url points at your Jira (HTTP 404)", jc.credentialDescription(), jc.BaseURL)
	}
	return fmt.Errorf("could not check the Jira credential at %s: %w", jc.BaseURL, err)
}

// GetJiraIssueInfo retrieve Jira issue information
// if issue is not found, we still return a JiraIssueInfo object with IssueExists set to false.
//
// A Jira that answers with something other than 404 - a rejected credential, a server error
// - is returned as an error rather than as a missing issue, so that a caller asserting on
// Jira references fails on what actually went wrong. A Jira that does not answer at all is
// reported through JiraIssueInfo.LookupFailure instead, which keeps the exit code of such a
// run what it has always been.
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
	if err != nil {
		if response == nil {
			// No response at all: DNS, TLS, timeout, connection refused. Nothing was learned
			// about the issue, so IssueExists stays false but the reason travels with the
			// result rather than becoming an error - a Jira outage must not fail a run that
			// would have passed before this distinction was drawn.
			result.LookupFailure = fmt.Errorf("failed to reach Jira at %s while looking up issue %s: %w", jc.BaseURL, issueID, err)
			return result, nil
		}
		switch response.StatusCode {
		case http.StatusNotFound:
			// The only status that answers the question. Jira returns 404 both for an
			// issue that does not exist and for one the credential may not browse - it
			// will not confirm the existence of an issue you cannot see - so a 404
			// alone cannot tell the two apart.
			return result, nil
		case http.StatusUnauthorized:
			return result, fmt.Errorf("failed to authenticate with Jira at %s: check the %s (Jira answered HTTP 401)", jc.BaseURL, jc.credentialDescription())
		case http.StatusForbidden:
			return result, fmt.Errorf("not permitted to view Jira issue %s at %s: check that the %s has the Browse Projects permission (Jira answered HTTP 403)", issueID, jc.BaseURL, jc.credentialDescription())
		default:
			// go-jira's message carries the status and the response body but names neither
			// the Jira nor the issue, so on a 5xx it reads as an unattributable "request
			// failed" in CI output. Say which call it was.
			return result, fmt.Errorf("failed to look up Jira issue %s at %s: %w", issueID, jc.BaseURL, err)
		}
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

// makeJiraIssueKeyPattern builds the regex matching Jira issue keys of the given projects.
// Jira issue keys consist of [project-key]-[sequential-number]; see
// https://support.atlassian.com/jira-software-cloud/docs/what-is-an-issue/#Workingwithissues-Projectandissuekeys
//
// The return value carries three distinct meanings:
//
//   - no project keys at all: the default pattern, matching keys of every project.
//   - at least one usable key: a pattern matching keys of those projects only.
//   - keys given, none of them usable: "", meaning no issue key can match.
//
// The "" case must NOT be compiled directly - the empty pattern matches at every position,
// so compiling it yields the exact opposite of what it means. Use jiraIssueKeyRegexp, which
// maps it to a nil regexp; this function is unexported so that nothing else can get it
// wrong. It is separate from the no-keys case on purpose: answering "these projects all
// turned out to be unusable" with the default pattern would widen a caller who named
// projects to every project, and on an attestation path the keys that widening invents are
// then looked up and attested.
//
// FindJiraIssueKeys uppercases the text before applying the pattern, so the pattern only
// needs to handle uppercase, and the project keys are uppercased here for the same reason.
// Each key is also quoted, so the pattern always compiles whatever the caller passes; a key
// a Jira project could actually have carries no regex metacharacters, so quoting leaves it
// unchanged. Keys are trimmed, and blank ones dropped rather than interpolated: an empty
// alternative reduces the group to nothing and leaves a pattern matching any -[0-9]+, which
// reports a phantom key such as -41284 out of CVE-2026-41284, and a whitespace-only key does
// the same thing one column over, since a space is not a metacharacter for QuoteMeta to
// escape. Trimming also stops " PROJ" from reporting " PROJ-123", which is not a key any
// Jira project has.
func makeJiraIssueKeyPattern(projectKeys []string) string {
	if len(projectKeys) == 0 {
		return defaultJiraIssueKeyPattern
	}
	upper := make([]string, 0, len(projectKeys))
	for _, k := range projectKeys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		upper = append(upper, regexp.QuoteMeta(strings.ToUpper(k)))
	}
	if len(upper) == 0 {
		return ""
	}
	return `\b(` + strings.Join(upper, "|") + `)-[0-9]+`
}

// jiraIssueKeyRegexp returns the compiled issue key pattern for the given project keys, or
// nil if no issue key can match, which makeJiraIssueKeyPattern reports as "".
//
// A project-key pattern is compiled per call, and cannot panic because
// makeJiraIssueKeyPattern quotes the keys it interpolates. When the pattern is the default
// one, the copy compiled at package level is returned instead.
func jiraIssueKeyRegexp(projectKeys []string) *regexp.Regexp {
	switch pattern := makeJiraIssueKeyPattern(projectKeys); pattern {
	case "":
		return nil
	case defaultJiraIssueKeyPattern:
		return defaultJiraIssueKeyRegexp
	default:
		return regexp.MustCompile(pattern)
	}
}

// FindJiraIssueKeys finds all Jira issue keys in text, filtering out
// partial matches from multi-segment identifiers like CVE-2026-41284.
// Matching is case-insensitive: the text is uppercased before the regex
// is applied, so all returned keys are in canonical uppercase form.
// A match is discarded if every occurrence in the uppercased text is
// immediately followed by a hyphen and a digit.
func FindJiraIssueKeys(text string, projectKeys []string) []string {
	re := jiraIssueKeyRegexp(projectKeys)
	if re == nil {
		// project keys were given but none of them is usable, so no key can belong to them
		return nil
	}
	upperText := strings.ToUpper(text)
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
