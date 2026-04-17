package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
)

const (
	githubLatestReleaseURL = "https://api.github.com/repos/kosli-dev/cli/releases/latest"
	updateCheckTimeout     = 1 * time.Second // max timeout when checking version
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// OverrideCheckForUpdate may be set in tests to replace the real HTTP check.
var OverrideCheckForUpdate func(currentVersion string) (string, error)

// CheckForUpdate is the public entry point — uses the real GitHub URL
func CheckForUpdate(currentVersion string) (string, error) {
	if OverrideCheckForUpdate != nil {
		return OverrideCheckForUpdate(currentVersion)
	}
	return checkForUpdateWithURL(currentVersion, githubLatestReleaseURL)
}

// checkForUpdateWithURL is the internal, testable implementation.
// It queries the given URL for the latest release and returns a non-empty
// notice string if the current version is outdated.
// It returns silently (empty string, nil) on any network or parse error,
// so it never blocks or fails a command.
// Set KOSLI_NO_UPDATE_CHECK=1 to skip entirely.
func checkForUpdateWithURL(currentVersion string, apiURL string) (string, error) {
	// checks disabled -skip
	if os.Getenv("KOSLI_NO_UPDATE_CHECK") != "" {
		return "", nil
	}
	// dev build — skip
	if currentVersion == "" || strings.HasPrefix(currentVersion, "dev") {
		return "", nil
	}

	// context provides the timeout and not http.Client
	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", nil
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", fmt.Sprintf("kosli-cli/%s", currentVersion))

	client := &http.Client{}
	resp, err := client.Do(req)
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

	latestVer, err := semver.NewVersion(release.TagName)
	if err != nil {
		return "", nil
	}
	currentVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", nil
	}

	if latestVer.GreaterThan(currentVer) {
		return fmt.Sprintf(
			"\nA new version of the Kosli CLI is available: %s (you have %s)\nUpgrade: https://docs.kosli.com/getting_started/install/\n",
			release.TagName, currentVersion,
		), nil
	}
	return "", nil
}
