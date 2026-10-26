package bitbucket

import (
	"encoding/json"
	"testing"

	"github.com/kosli-dev/cli/internal/types"
	"github.com/stretchr/testify/require"
)

// Participant shapes follow the participant definition in
// https://api.bitbucket.org/swagger.json: state is approved, changes_requested
// or null; participated_on is the approval time, else the last comment, else null.
func TestBitbucketReviews(t *testing.T) {
	var participants []map[string]any
	require.NoError(t, json.Unmarshal([]byte(`[
		{"user":{"display_name":"Ada"},"role":"REVIEWER","approved":true,"state":"approved","participated_on":"2026-03-01T13:00:00.123456+00:00"},
		{"user":{"display_name":"Bob"},"role":"REVIEWER","approved":false,"state":"changes_requested","participated_on":"2026-03-01T14:00:00+00:00"},
		{"user":{"display_name":"Cy"},"role":"REVIEWER","approved":false,"state":"changes_requested","participated_on":null},
		{"user":{"display_name":"Di"},"role":"PARTICIPANT","approved":false,"state":null,"participated_on":"2026-03-01T15:00:00+00:00"},
		{"user":{"display_name":"Ed"},"role":"REVIEWER","approved":false,"state":null,"participated_on":null},
		{"user":null,"role":"REVIEWER","approved":false,"state":"changes_requested","participated_on":null},
		{"user":null,"role":"PARTICIPANT","approved":false,"state":null,"participated_on":"2026-03-01T16:00:00+00:00"},
		{"user":{"display_name":null},"role":"REVIEWER","approved":true,"state":"approved","participated_on":"2026-03-01T17:00:00+00:00"}
	]`), &participants))

	reviews, err := bitbucketReviews(participants)
	require.NoError(t, err)
	require.Equal(t, []types.PRApprovals{
		{Username: "Ada", State: types.ReviewApproved, Timestamp: 1772370000},
		{Username: "Bob", State: types.ReviewChangesRequested, Timestamp: 1772373600},
		{Username: "Cy", State: types.ReviewChangesRequested},
		{Username: "Di", State: types.ReviewCommented, Timestamp: 1772377200},
		{State: types.ReviewChangesRequested},
	}, reviews)
}

func TestBitbucketReviews_RejectsABadTimestamp(t *testing.T) {
	var participants []map[string]any
	require.NoError(t, json.Unmarshal([]byte(`[
		{"user":{"display_name":"Ada"},"approved":true,"state":"approved","participated_on":"yesterday"}
	]`), &participants))

	_, err := bitbucketReviews(participants)
	require.ErrorContains(t, err, "participated_on")
}
