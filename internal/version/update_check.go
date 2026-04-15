package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	githubLatestReleaseURL = "https://api.github.com/repos/kosli-dev/cli/releases/latest"
	updateCheckTimeout     = 3 * time.Second
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckForUpdate is the public entry point — uses the real GitHub URL
func CheckForUpdate(currentVersion string) (string, error) {
	return checkForUpdateWithURL(currentVersion, githubLatestReleaseURL)
}

// CheckForUpdate queries the GitHub API for the latest release and returns
// a non-empty notice string if the current version is outdated.
// It returns silently (empty string, nil) on any network or parse error,
// so it never blocks or fails a command.
// Set KOSLI_NO_UPDATE_CHECK=1 to skip entirely.
// checkForUpdateWithURL is the testable implementation
func checkForUpdateWithURL(currentVersion string, apiURL string) (string, error) {
	if os.Getenv("KOSLI_NO_UPDATE_CHECK") != "" {
		return "", nil
	}
	if currentVersion == "" || strings.HasPrefix(currentVersion, "main") || strings.Contains(currentVersion, "+unreleased") {
		return "", nil // dev build — skip
	}

	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", fmt.Sprintf("kosli-cli/%s", currentVersion))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", nil
	}

	latestVersion := release.TagName
	// Strip leading 'v' for semver comparison
	if latestVersion != "" && latestVersion[0] == 'v' {
		latestVersion = latestVersion[1:]
	}
	current := currentVersion
	if len(current) > 0 && current[0] == 'v' {
		current = current[1:]
	}

	if latestVersion != "" && current != latestVersion {
		return fmt.Sprintf(
			"\nA new version of the Kosli CLI is available: v%s (you have v%s)\nUpgrade: https://docs.kosli.com/getting_started/install/\n",
			latestVersion, current,
		), nil
	}
	return "", nil
}
