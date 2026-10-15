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

func TestCreateEnvironmentPutsTheRequestToTheOrgsEnvironments(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, ``)
	client := newTestClient(t, server.URL)
	include := true

	result, err := client.CreateEnvironment(context.Background(), CreateEnvironmentRequest{
		Name: "prod", Type: "K8S", Description: "production", IncludeScaling: &include,
	})

	require.NoError(t, err)
	require.True(t, result.Created)
	require.Equal(t, http.MethodPut, seen.method)
	require.Equal(t, "/api/v2/environments/test-org", seen.rawPath)
	require.JSONEq(t, `{"name":"prod","type":"K8S","description":"production","include_scaling":true}`, seen.body)
}

func TestReportSnapshotPutsThePayloadToTheEnvironmentTypesReport(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, ``)
	client := newTestClient(t, server.URL)

	_, err := client.ReportSnapshot(context.Background(), "prod eu", "azure-apps", map[string]any{"artifacts": []any{}})

	require.NoError(t, err)
	require.Equal(t, http.MethodPut, seen.method)
	require.Equal(t, "/api/v2/environments/test-org/prod%20eu/report/azure-apps", seen.rawPath)
	require.JSONEq(t, `{"artifacts":[]}`, seen.body)
}

func TestADryRunSendsNeitherEnvironmentWrite(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, ``)
	client := New(DryRun(newHTTPClient(t)), server.URL, "test-org", "test-token")

	_, err := client.CreateEnvironment(context.Background(), CreateEnvironmentRequest{Name: "prod", Type: "K8S"})
	require.NoError(t, err)
	_, err = client.ReportSnapshot(context.Background(), "prod", "K8S", map[string]any{})
	require.NoError(t, err)

	require.Equal(t, 0, seen.hits)
}
