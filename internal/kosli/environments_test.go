package kosli

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnvironmentRequestsTheEnvironmentAndDecodesItsType(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `{"name":"prod","type":"K8S","last_reported_at":null}`)
	client := newTestClient(t, server.URL)

	environment, err := client.GetEnvironment(context.Background(), "prod")

	require.NoError(t, err)
	require.Equal(t, http.MethodGet, seen.method)
	require.Equal(t, "/api/v2/environments/test-org/prod", seen.rawPath)
	value, err := environment.Value()
	require.NoError(t, err)
	require.Equal(t, Environment{Name: "prod", Type: "K8S"}, value)
}

func TestGetEnvironmentOfAMissingEnvironmentIsNotFound(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusNotFound, `{"message":"Environment not found"}`)
	client := newTestClient(t, server.URL)

	_, err := client.GetEnvironment(context.Background(), "missing")

	require.True(t, IsNotFound(err))
}
