package jira

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/types"

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

func getJiraTicketURL() {
	// Define the Jira base URL
	jiraBaseURL := "https://your-jira-instance/browse/"

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
