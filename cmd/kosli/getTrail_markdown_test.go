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
