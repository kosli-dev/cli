package regopolicies

import (
	"errors"
	"net/http"
	"testing"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
)

const validBody = `{"identifier":"pr-approved","version":3,"status":"valid",` +
	`"digest":"sha256:abc123","input_schema":"trail","comment":"Initial version",` +
	`"created_at":1757000000.5,"created_by":"Jane Doe","validated_at":1757000002.25,` +
	`"opa_version":"1.21.1","entrypoint":"policy",` +
	`"params_schema":{"type":"object","properties":{"threshold":{"type":"integer"}}}}`

func TestGetVersionReadsAValidVersionBack(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, validBody)

	version, err := newTestClient(t, server.URL, false).GetVersion("my-org", "pr-approved", 3)
	require.NoError(t, err)

	require.Equal(t, http.MethodGet, seen.method)
	require.Equal(t, "/api/v2/rego-policies/my-org/pr-approved/versions/3", seen.path)
	require.Equal(t, "Bearer test-token", seen.authHeader)

	validatedAt := 1757000002.25
	require.Equal(t, &Version{
		Identifier:  "pr-approved",
		Version:     3,
		Status:      StatusValid,
		Digest:      "sha256:abc123",
		InputSchema: "trail",
		Comment:     "Initial version",
		CreatedAt:   1757000000.5,
		CreatedBy:   "Jane Doe",
		ValidatedAt: &validatedAt,
		OPAVersion:  "1.21.1",
		Entrypoint:  "policy",
		ParamsSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"threshold": map[string]any{"type": "integer"}},
		},
	}, version)
}

const rejectedBody = `{"identifier":"pr-approved","version":4,"status":"rejected",` +
	`"digest":"sha256:def456","input_schema":"trail","comment":"",` +
	`"created_at":1757000000.5,"created_by":"Jane Doe","validated_at":1757000002.25,` +
	`"opa_version":"1.21.1","error":{"kind":"compile",` +
	`"message":"policy.rego:4: rego_parse_error: unexpected eof token"}}`

func TestGetVersionReadsWhyAVersionIsNotValid(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusOK, rejectedBody)

	version, err := newTestClient(t, server.URL, false).GetVersion("my-org", "pr-approved", 4)
	require.NoError(t, err)

	require.Equal(t, StatusRejected, version.Status)
	require.Equal(t, &ValidationError{
		Kind:    "compile",
		Message: "policy.rego:4: rego_parse_error: unexpected eof token",
	}, version.Error)
	require.Empty(t, version.Entrypoint)
	require.Nil(t, version.ParamsSchema)
}

func TestGetVersionSurfacesAMissingVersion(t *testing.T) {
	const message = "Version 9 of rego policy 'pr-approved' does not exist in org 'my-org'"
	server, _ := newFakeServer(t, http.StatusNotFound, `{"message":"`+message+`"}`)

	version, err := newTestClient(t, server.URL, false).GetVersion("my-org", "pr-approved", 9)

	require.Nil(t, version)
	var apiError *requests.APIError
	require.True(t, errors.As(err, &apiError), "expected an API error, got %T", err)
	require.Equal(t, http.StatusNotFound, apiError.StatusCode)
	require.Equal(t, message, apiError.Message)
}

// An unknown status means the server has added one; reading it as terminal
// would report an unfinished validation as an outcome.
func TestOnlyAValidatedVersionIsTerminal(t *testing.T) {
	for status, terminal := range map[string]bool{
		StatusPending:  false,
		StatusValid:    true,
		StatusRejected: true,
		StatusFailed:   true,
		"validating":   false,
	} {
		require.Equal(t, terminal, (&Version{Status: status}).IsTerminal(), status)
	}
}
