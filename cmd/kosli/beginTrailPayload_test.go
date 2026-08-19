package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// An API client that omits description leaves the trail's stored description
// alone. Keeping the key out of the payload is what gives the CLI the same
// behaviour, because the server only writes the field when it is present.
func TestTrailPayloadOmitsUnsetDescription(t *testing.T) {
	body, err := json.Marshal(TrailPayload{Name: "test-123"})
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &got))
	require.NotContains(t, got, "description")
}

// The same for user_data.
func TestTrailPayloadOmitsUnsetUserData(t *testing.T) {
	body, err := json.Marshal(TrailPayload{Name: "test-123"})
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &got))
	require.NotContains(t, got, "user_data")
}
