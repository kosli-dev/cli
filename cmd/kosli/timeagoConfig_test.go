package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xeonx/timeago"
)

func TestTimeagoFormattingLeavesSharedEnglishConfigAlone(t *testing.T) {
	t.Run("list snapshots prints a duration without the ago suffix", func(t *testing.T) {
		suffix, limit := timeago.English.PastSuffix, timeago.English.Max
		raw := `[{"index": 3, "from": 1452902400, "to": 0, "duration": 172800, "compliant": true}]`

		var out bytes.Buffer
		require.NoError(t, printSnapshotsListAsTable(raw, &out, 1))

		require.Contains(t, out.String(), "2 days")
		require.NotContains(t, out.String(), "2 days ago")
		require.Equal(t, suffix, timeago.English.PastSuffix)
		require.Equal(t, limit, timeago.English.Max)
	})

	t.Run("get snapshot prints how long an artifact has been running", func(t *testing.T) {
		suffix, limit := timeago.English.PastSuffix, timeago.English.Max
		started := time.Now().Add(-3 * time.Hour).Unix()
		raw := fmt.Sprintf(`{"index": 1, "type": "K8S", "artifacts": [{"name": "nginx", "fingerprint": "abc",
			"compliant": true, "annotation": {"now": 1}, "creationTimestamp": [%d]}]}`, started)

		var out bytes.Buffer
		require.NoError(t, printSnapshotAsTable(raw, &out, 1))

		require.Contains(t, out.String(), "3 hours ago")
		require.Equal(t, suffix, timeago.English.PastSuffix)
		require.Equal(t, limit, timeago.English.Max)
	})

	t.Run("a full timestamp keeps the ago suffix", func(t *testing.T) {
		t.Setenv("KOSLI_TESTS_FORMATTED_TIMESTAMP", "True")
		suffix, limit := timeago.English.PastSuffix, timeago.English.Max

		got, err := formattedTimestamp(time.Now().Add(-2*time.Hour).Unix(), false)
		require.NoError(t, err)

		require.Contains(t, got, "2 hours ago")
		require.Equal(t, suffix, timeago.English.PastSuffix)
		require.Equal(t, limit, timeago.English.Max)
	})
}
