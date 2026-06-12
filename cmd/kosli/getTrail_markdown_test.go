package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPrintTrailAsMarkdown is a server-free unit test for the markdown renderer.
// It feeds a representative trail payload through printTrailAsMarkdown and
// compares the result against the golden file used by the integration test.
func TestPrintTrailAsMarkdown(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true") // makes formattedTimestamp use a fixed time

	raw := `{
		"name": "cli-build-1",
		"description": "test trail",
		"compliance_state": "INCOMPLETE",
		"last_modified_at": 1452902400,
		"events": [
			{"type": "trail_reported", "timestamp": 1452902400}
		]
	}`

	var buf bytes.Buffer
	err := printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	want, err := os.ReadFile(goldenPath("output/get/get-trail-markdown.txt"))
	require.NoError(t, err)

	require.Equal(t, string(want), buf.String())
}

// TestPrintTrailAsMarkdownGitCommitAndEscaping covers the git-commit block and
// mdCell escaping: pipes, LF, CRLF and bare CR must not break the table layout,
// and missing (null) fields must render as empty cells, not "<nil>".
func TestPrintTrailAsMarkdownGitCommitAndEscaping(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")

	raw := `{
		"name": "cli-build-2",
		"description": null,
		"compliance_state": "COMPLIANT",
		"last_modified_at": 1452902400,
		"git_commit_info": {
			"sha1": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b",
			"author": "Jane | Doe",
			"timestamp": 1452902400,
			"url": "https://github.com/kosli-dev/cli/commit/1a2b3c4",
			"message": "fix: a | in a title\r\nsecond line\rthird line\nfourth line"
		},
		"events": []
	}`

	var buf bytes.Buffer
	err := printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	got := buf.String()

	require.NotContains(t, got, "<nil>", "null fields must render as empty cells")
	require.NotContains(t, got, "\r", "carriage returns must not survive into the markdown")
	require.Contains(t, got, "| Description |  |")
	require.Contains(t, got, "| Author | Jane \\| Doe |")
	require.Contains(t, got, "| Message | fix: a \\| in a title<br>second line<br>third line<br>fourth line |")
	require.Contains(t, got, "### Git commit")
	require.Contains(t, got, "| URL | https://github.com/kosli-dev/cli/commit/1a2b3c4 |")
	require.Contains(t, got, "_No events._")
}
