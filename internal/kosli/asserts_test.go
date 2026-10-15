package kosli

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssertArtifactRequestsTheAssertionAndReturnsTheRawBody(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `{"compliant":true}`)
	client := newTestClient(t, server.URL)

	result, err := client.AssertArtifact(context.Background(), "abc", AssertArtifactFilter{
		Flow:        "flow-1",
		Environment: "prod",
		Policies:    []string{"p1", "p2"},
	})

	require.NoError(t, err)
	require.Equal(t, http.MethodGet, seen.method)
	require.Equal(t, "/api/v2/asserts/test-org/fingerprint/abc", seen.rawPath)
	require.Equal(t, "environment_name=prod&flow_name=flow-1&policy_name=p1&policy_name=p2", seen.rawQuery)
	require.JSONEq(t, `{"compliant":true}`, string(result.Raw))
}

func TestAssertArtifactWithoutAFilterSendsNoQuery(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `{"compliant":true}`)
	client := newTestClient(t, server.URL)

	_, err := client.AssertArtifact(context.Background(), "abc", AssertArtifactFilter{})

	require.NoError(t, err)
	require.Empty(t, seen.rawQuery)
}
