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

// omitempty must not swallow a value the user did give, so a payload carrying
// both fields sends both.
func TestTrailPayloadKeepsSetDescriptionAndUserData(t *testing.T) {
	payload := TrailPayload{
		Name:        "test-123",
		Description: "the release trail",
		UserData:    map[string]interface{}{"release": "2.11.21"},
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &got))
	require.Equal(t, "the release trail", got["description"])
	require.Equal(t, map[string]interface{}{"release": "2.11.21"}, got["user_data"])
}
