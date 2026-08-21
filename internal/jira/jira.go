package jira

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/kosli-dev/cli/internal/logger"
)

// LookupStatus records what a lookup established about an issue. It exists because
// IssueExists cannot say why it is false: Jira answers 404 both for an issue that does
// not exist and for one the caller may not view, and a caller whose credentials Jira
// rejected may not view any issue. Reporting the latter as a missing issue is what makes
// an expired API token look like a compliance failure.
//
// LookupUnverified is the zero value, so a JiraIssueInfo returned from a path that never
// reached Jira does not claim the issue was found.
type LookupStatus int

const (
	// LookupUnverified: existence could not be determined.
	LookupUnverified LookupStatus = iota
	// IssueFound: Jira returned the issue.
	IssueFound
	// IssueMissing: Jira answered 404 and nothing indicates the credentials were rejected.
	IssueMissing
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

	// LookupStatus and LookupReason let a caller report what happened; LookupReason is
	// set only when the status is LookupUnverified. Both carry json:"-" so that the
	// attestation payload stays exactly what the Kosli API accepts today.
	LookupStatus LookupStatus `json:"-"`
	LookupReason string       `json:"-"`
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
//
// An error is returned in exactly the cases it was returned in before LookupStatus
// existed - Jira answered with a status other than 200 or 404 - because callers stop
// their work on it. Everything else is reported through LookupStatus and LookupReason,
// so a lookup that cannot be completed enriches the caller's message without changing
// what the caller does.
func (jc *JiraConfig) GetJiraIssueInfo(issueID string, issueFields string, log *logger.Logger) (*JiraIssueInfo, error) {
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
	logLookup(log, issueID, jc.BaseURL, response)

	switch {
	case err == nil:
		result.LookupStatus = IssueFound
	case response == nil:
		// go-jira returns no response when the request itself never completed, so this
		// is a DNS, proxy, TLS or timeout failure rather than an answer from Jira.
		result.LookupReason = fmt.Sprintf("could not reach Jira at %s: %s", jc.BaseURL, oneLine(err.Error()))
	case response.StatusCode == http.StatusNotFound:
		if reason, rejected := credentialsRejected(response.Header); rejected {
			result.LookupReason = fmt.Sprintf("Jira did not accept the %s for %s (%s), so it answered 404 for every issue; the credentials may have expired or been revoked",
				jc.authDescription(), jc.BaseURL, reason)
		} else {
			result.LookupStatus = IssueMissing
		}
	default:
		message := fmt.Sprintf("looking up Jira issue %s at %s using %s returned %s",
			issueID, jc.BaseURL, jc.authDescription(), response.Status)
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			message += "; the credentials may have expired or been revoked"
		}
		return result, fmt.Errorf("%s: %s", message, errorDetail(err))
	}

	if issue != nil {
		result.IssueExists = true
		if issue.Fields != nil {
			result.IssueFields = issue.Fields
		}
	}
	return result, nil
}

// logLookup records the outcome of a lookup at debug level. The status code alone does not
// say whether Jira accepted the credentials, so the two headers that do are logged with it.
func logLookup(log *logger.Logger, issueID, baseURL string, response *jira.Response) {
	if response == nil {
		log.Debug("looking up Jira issue %s at %s got no response", issueID, baseURL)
		return
	}
	log.Debug("looking up Jira issue %s at %s returned status %d (X-AUSERNAME: %q, X-Seraph-LoginReason: %q)",
		issueID, baseURL, response.StatusCode, response.Header.Get("X-AUSERNAME"), response.Header.Get("X-Seraph-LoginReason"))
}

// credentialsRejected reports whether a response carries positive evidence that Jira did
// not accept the credentials, and a reason to quote in a message.
//
// These headers are the only such evidence on the response itself: the issue endpoint
// answers 404 for an issue the caller may not view, which is indistinguishable by status
// code from an issue that does not exist. They are Seraph-era headers, set by Jira Server
// and Data Center and historically by Jira Cloud, but not part of the documented API
// contract - so their presence is trustworthy and their absence proves nothing, which is
// why a missing header leaves the lookup classified exactly as it was before.
func credentialsRejected(header http.Header) (string, bool) {
	if reason := header.Get("X-Seraph-LoginReason"); reason != "" && !strings.EqualFold(reason, "OK") {
		return "X-Seraph-LoginReason: " + reason, true
	}
	if strings.EqualFold(header.Get("X-AUSERNAME"), "anonymous") {
		return "the request was handled as anonymous", true
	}
	return "", false
}

// authDescription names the credentials in use, so that a message about rejected
// credentials says which ones to renew. The API token and the PAT are secrets and are
// never included.
func (jc *JiraConfig) authDescription() string {
	if jc.Username != "" && jc.APIToken != "" {
		return fmt.Sprintf("username %s and API token", jc.Username)
	}
	return "personal access token"
}

// errorDetail extracts what Jira itself said, for a message that already names the status.
// go-jira parses a JSON error body into jira.Error and renders it together with a generic
// "request failed ... Status code: N" sentence, which only repeats the status we print.
func errorDetail(err error) string {
	var jiraErr *jira.Error
	if errors.As(err, &jiraErr) && len(jiraErr.ErrorMessages) > 0 {
		return oneLine(strings.Join(jiraErr.ErrorMessages, "; "))
	}
	return oneLine(err.Error())
}

// oneLine flattens and truncates text taken from a Jira error before it goes into a
// message. NewJiraError embeds the whole response body in the error it builds, and Jira
// can answer with an HTML login page, so without this a single log line could carry a
// whole page.
func oneLine(text string) string {
	flattened := strings.Join(strings.Fields(text), " ")
	const maxRunes = 200
	runes := []rune(flattened)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "..."
	}
	return flattened
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
