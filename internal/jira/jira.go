package jira

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"

	"log"
	"os/exec"
	"regexp"
	"strings"
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

type JiraIssueResult struct {
	IssueID     string `json:"issue_id"`
	IssueURL    string `json:"issue_url"`
	IssueExists bool   `json:"issue_exists"`
}

func GetJiraIssue(jiraBaseURL, issueID string) (*JiraIssueResult, error) {
	result := &JiraIssueResult{
		IssueID:     issueID,
		IssueURL:    fmt.Sprintf("%s/browse/%s", jiraBaseURL, issueID),
		IssueExists: false,
	}
	tp := jira.BasicAuthTransport{
		Username: "username",
		Password: "top-secret",
	}

	jiraClient, err := jira.NewClient(tp.Client(), jiraBaseURL)
	if err != nil {
		return result, err
	}
	issue, _, err := jiraClient.Issue.Get(issueID, nil)
	if err != nil {
		return result, err
	}

	if issue != nil {
		result.IssueExists = true
	}
	return result, nil
}

func getJiraTicketURL(jiraBaseURL string) {

	// Get the current commit hash
	commitHash, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		log.Fatal(err)
	}

	// Get the Jira ticket key from the current branch or commit message
	jiraKey := getJiraKey(string(commitHash))
	if jiraKey == "" {
		log.Fatal("No Jira ticket key found.")
	}

	// Construct the Jira ticket URL
	jiraURL := fmt.Sprintf("%s%s", jiraBaseURL, jiraKey)
	fmt.Println("Jira ticket URL:", jiraURL)
}

func getJiraKey(commitHash string) string {
	// Get the branch name or commit message
	branchOrMessage, err := exec.Command("git", "name-rev", "--name-only", "HEAD").Output()
	if err != nil {
		log.Fatal(err)
	}

	// Regex pattern to match Jira ticket keys
	pattern := regexp.MustCompile(`(ABC-\d+)`)

	// Search for Jira ticket key in the branch name or commit message
	match := pattern.FindStringSubmatch(string(branchOrMessage))
	if match != nil {
		return strings.ToUpper(match[0])
	}

	return ""
}
