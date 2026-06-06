package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

const serviceAccountApiKeysDesc = `Manage API keys for a Kosli service account.`

func newServiceAccountApiKeysCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"ak"},
		Short:   serviceAccountApiKeysDesc,
		Long:    serviceAccountApiKeysDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newCreateApiKeyCmd(out),
		newRevokeApiKeyCmd(out),
		newRotateApiKeyCmd(out),
		newListApiKeysCmd(out),
	)

	return cmd
}

// apiKeyResponse models the JSON returned by the create and rotate endpoints.
// The key value is only ever returned once, at creation/rotation time.
type apiKeyResponse struct {
	Id                   string  `json:"id"`
	Key                  string  `json:"key"`
	Description          string  `json:"description"`
	CreatedAt            float64 `json:"created_at"`
	ExpiresAt            float64 `json:"expires_at"`
	GracePeriodExpiresAt float64 `json:"grace_period_expires_at,omitempty"`
}

// parseExpiresAt converts a user-supplied --expires-at value into a Unix
// (epoch-second) timestamp. It accepts a bare epoch integer, or one of the
// date/time layouts below (interpreted as UTC). An empty string returns 0.
func parseExpiresAt(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	if epoch, err := strconv.ParseInt(value, 10, 64); err == nil {
		return epoch, nil
	}

	formats := []string{
		"2006-1-2",
		"2006-1-2 15:04:05",
		time.RFC3339,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t.UTC().Unix(), nil
		}
	}

	return 0, fmt.Errorf("invalid --expires-at value %q: expected an epoch timestamp or a date like '2006-01-02', '2006-01-02 15:04:05', or an RFC3339 timestamp", value)
}

// printApiKeyAsTable renders a single api key (the create response) as a table.
func printApiKeyAsTable(raw string, out io.Writer, page int) error {
	var key apiKeyResponse
	if err := json.Unmarshal([]byte(raw), &key); err != nil {
		return err
	}

	rows, err := apiKeyTableRows(key)
	if err != nil {
		return err
	}
	tabFormattedPrint(out, []string{}, rows)
	return nil
}

// printApiKeysAsTable renders one or more api keys (the rotate response) as
// table blocks separated by a blank line.
func printApiKeysAsTable(raw string, out io.Writer, page int) error {
	var keys []apiKeyResponse
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return err
	}

	for i, key := range keys {
		if i > 0 {
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
		rows, err := apiKeyTableRows(key)
		if err != nil {
			return err
		}
		tabFormattedPrint(out, []string{}, rows)
	}
	return nil
}

// optionalTimestamp formats an epoch timestamp, returning "N/A" when it is
// unset (nil, or a zero value meaning "never"/"not set").
func optionalTimestamp(epoch interface{}) (string, error) {
	switch v := epoch.(type) {
	case nil:
		return "N/A", nil
	case float64:
		if v == 0 {
			return "N/A", nil
		}
	case int64:
		if v == 0 {
			return "N/A", nil
		}
	}
	return formattedTimestamp(epoch, false)
}

// apiKeyTableRows builds the key:value rows describing a single api key.
func apiKeyTableRows(key apiKeyResponse) ([]string, error) {
	createdAt, err := formattedTimestamp(key.CreatedAt, false)
	if err != nil {
		return nil, err
	}
	expiresAt, err := optionalTimestamp(key.ExpiresAt)
	if err != nil {
		return nil, err
	}

	rows := []string{}
	rows = append(rows, fmt.Sprintf("ID:\t%s", key.Id))
	rows = append(rows, fmt.Sprintf("Key:\t%s", key.Key))
	rows = append(rows, fmt.Sprintf("Description:\t%s", key.Description))
	rows = append(rows, fmt.Sprintf("Created At:\t%s", createdAt))
	rows = append(rows, fmt.Sprintf("Expires At:\t%s", expiresAt))
	if key.GracePeriodExpiresAt != 0 {
		gracePeriodExpiresAt, err := formattedTimestamp(key.GracePeriodExpiresAt, false)
		if err != nil {
			return nil, err
		}
		rows = append(rows, fmt.Sprintf("Old Key Valid Until:\t%s", gracePeriodExpiresAt))
	}
	return rows, nil
}
