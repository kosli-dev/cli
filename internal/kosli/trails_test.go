package kosli

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTrailRequestsTheTrailInTheOrgAndReturnsTheRawBody(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `{"name":"trail 1","compliance_status":{}}`)
	client := newTestClient(t, server.URL)

	result, err := client.GetTrail(context.Background(), "flow-1", "trail 1")

	require.NoError(t, err)
	require.Equal(t, http.MethodGet, seen.method)
	require.Equal(t, "/api/v2/trails/test-org/flow-1/trail%201", seen.rawPath)
	require.JSONEq(t, `{"name":"trail 1","compliance_status":{}}`, string(result.Raw))
}

func TestGetTrailOfAMissingTrailIsNotFound(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusNotFound, `{"message":"Trail not found"}`)
	client := newTestClient(t, server.URL)

	_, err := client.GetTrail(context.Background(), "flow-1", "missing")

	require.True(t, IsNotFound(err))
}
