package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPrintTrailAsTableCommitInfo checks that the table renderer tolerates
// commit info with only some of its fields, which the API accepts for trails.
func TestPrintTrailAsTableCommitInfo(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true") // makes formattedTimestamp use a fixed time

	for _, tc := range []struct {
		name       string
		commitInfo string
		want       []string
	}{
		{
			name:       "only sha1",
			commitInfo: `{"sha1": "abc1234def"}`,
			want:       []string{"Sha1:            abc1234def", "Author:          N/A", "Timestamp:       N/A", "message:         N/A"},
		},
		{
			name:       "sha1 and branch",
			commitInfo: `{"sha1": "abc1234def", "branch": "main"}`,
			want:       []string{"abc1234def", "Author:          N/A", "Timestamp:       N/A", "message:         N/A"},
		},
		{
			name:       "full commit info",
			commitInfo: `{"sha1": "abc1234def", "author": "Jane <jane@example.com>", "timestamp": 1452902400, "url": "https://example.com/c/abc1234def", "message": "first line\nsecond line", "branch": "main"}`,
			want:       []string{"abc1234def", "Jane <jane@example.com>", "Timestamp:       Sat, 16 Jan 2016", "url:             https://example.com/c/abc1234def", "message:         first line\n                   second line"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := `{"name": "t1", "description": "d", "compliance_state": "INCOMPLETE",
				"last_modified_at": 1452902400, "git_commit_info": ` + tc.commitInfo + `, "events": []}`

			var buf bytes.Buffer
			require.NoError(t, printTrailAsTable(raw, &buf, 0))
			for _, w := range tc.want {
				require.Contains(t, buf.String(), w)
			}
		})
	}
}
