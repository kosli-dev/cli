package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

const lookedUpSHA = "4f9d2a7c1e6b8a3d5f0e2c9b7a1d4e6f8c3b5a70"

var _ types.CommitFinder = CommitFinder{}

func TestCommitBySHA(t *testing.T) {
	t.Run("returns the commit's message, author, time and page", func(t *testing.T) {
		var gotAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/api/v3/repos/acme/widget/commits/"+lookedUpSHA, r.URL.Path)
			gotAuth = r.Header.Get("Authorization")
			_, _ = fmt.Fprintf(w, `{
				"sha": %q,
				"html_url": "https://github.com/acme/widget/commit/%s",
				"author": {"login": "jdoe"},
				"commit": {
					"message": "Add the widget\n\nIt does widget things.\n",
					"author": {"name": "Jane Doe", "email": "jane@example.com", "date": "2026-11-01T09:00:00Z"}
				}
			}`, lookedUpSHA, lookedUpSHA)
		}))
		defer server.Close()

		commit, err := CommitFinder{Token: "a-token", BaseURL: server.URL + "/"}.CommitBySHA("acme/widget", lookedUpSHA)
		require.NoError(t, err)
		require.Equal(t, &types.Commit{
			SHA:       lookedUpSHA,
			Message:   "Add the widget\n\nIt does widget things.",
			Author:    "Jane Doe <jane@example.com>",
			Timestamp: 1793523600,
			URL:       "https://github.com/acme/widget/commit/" + lookedUpSHA,
		}, commit)
		require.Equal(t, "Bearer a-token", gotAuth)
	})

	t.Run("names the repository and commit when GitHub refuses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"message": "Not Found"}`)
		}))
		defer server.Close()

		_, err := CommitFinder{Token: "a-token", BaseURL: server.URL + "/"}.CommitBySHA("acme/widget", lookedUpSHA)
		require.ErrorContains(t, err, "failed to get commit "+lookedUpSHA+" from GitHub repository acme/widget")
		require.ErrorContains(t, err, "404 Not Found")
	})

	t.Run("refuses a repository not named owner/name", func(t *testing.T) {
		_, err := CommitFinder{Token: "a-token"}.CommitBySHA("widget", lookedUpSHA)
		require.EqualError(t, err, `GitHub repository "widget" is not in the form owner/name`)
	})
}
