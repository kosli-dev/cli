package regopolicies

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
)

type received struct {
	hits        int
	method      string
	path        string
	escapedPath string
	authHeader  string
	contentType string
	body        map[string]any
}

func newFakeServer(t *testing.T, status int, responseBody string) (*httptest.Server, *received) {
	t.Helper()
	seen := &received{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.hits++
		seen.method = r.Method
		seen.path = r.URL.Path
		seen.escapedPath = r.URL.EscapedPath()
		seen.authHeader = r.Header.Get("Authorization")
		seen.contentType = r.Header.Get("Content-Type")

		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		if len(raw) > 0 {
			require.NoError(t, json.Unmarshal(raw, &seen.body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, err = w.Write([]byte(responseBody))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	return server, seen
}

func newTestClient(t *testing.T, host string, dryRun bool) *Client {
	t.Helper()
	// No retries, so a 5xx surfaces as itself instead of being retried away.
	httpClient, err := requests.NewKosliClient("", 0, false, logger.NewLogger(io.Discard, io.Discard, false))
	require.NoError(t, err)
	return NewClient(httpClient, host, "test-token", dryRun)
}

func aPublishRequest() PublishRequest {
	return PublishRequest{
		Identifier:  "pr-approved",
		InputSchema: "trail",
		Comment:     "Initial version",
		Files:       map[string]string{"policy.rego": "package policy\n\nallow := true\n"},
	}
}

const pendingBody = `{"identifier":"pr-approved","version":1,"status":"pending",` +
	`"digest":"sha256:abc123","input_schema":"trail","comment":"Initial version",` +
	`"created_at":1757000000.5,"created_by":"Jane Doe"}`

func TestPublishSendsTheRequestTheAPIExpects(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusAccepted, pendingBody)

	_, err := newTestClient(t, server.URL, false).Publish("my-org", aPublishRequest())
	require.NoError(t, err)

	require.Equal(t, 1, seen.hits)
	require.Equal(t, http.MethodPut, seen.method)
	require.Equal(t, "/api/v2/rego-policies/my-org/pr-approved", seen.path)
	require.Equal(t, "Bearer test-token", seen.authHeader)
	require.Contains(t, seen.contentType, "application/json")
}

func TestPublishSendsExactlyTheAgreedPayload(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusAccepted, pendingBody)

	_, err := newTestClient(t, server.URL, false).Publish("my-org", aPublishRequest())
	require.NoError(t, err)

	// The server forbids unknown fields, so anything extra is a refusal.
	require.Equal(t, map[string]any{
		"input_schema": "trail",
		"comment":      "Initial version",
		"files":        map[string]any{"policy.rego": "package policy\n\nallow := true\n"},
	}, seen.body)
}

func TestPublishSendsAnEmptyCommentRatherThanLeavingItOut(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusAccepted, pendingBody)
	request := aPublishRequest()
	request.Comment = ""

	_, err := newTestClient(t, server.URL, false).Publish("my-org", request)
	require.NoError(t, err)

	require.Contains(t, seen.body, "comment")
	require.Equal(t, "", seen.body["comment"])
}

func TestPublishSendsTheNameAndDescriptionOnlyWhenGiven(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusAccepted, pendingBody)
	request := aPublishRequest()
	request.Name = "Pull request approved"
	request.Description = "Every change reached main through an approved PR"

	_, err := newTestClient(t, server.URL, false).Publish("my-org", request)
	require.NoError(t, err)

	require.Equal(t, "Pull request approved", seen.body["name"])
	require.Equal(t, "Every change reached main through an approved PR", seen.body["description"])
}

func TestPublishReadsThePendingVersionBack(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusAccepted, pendingBody)

	version, err := newTestClient(t, server.URL, false).Publish("my-org", aPublishRequest())
	require.NoError(t, err)

	require.Equal(t, &Version{
		Identifier:  "pr-approved",
		Version:     1,
		Status:      StatusPending,
		Digest:      "sha256:abc123",
		InputSchema: "trail",
		Comment:     "Initial version",
		CreatedAt:   1757000000.5,
		CreatedBy:   "Jane Doe",
	}, version)
}

func TestPublishSurfacesARefusalInTheServersWords(t *testing.T) {
	for _, test := range []struct {
		name    string
		status  int
		message string
	}{
		{"bad input", http.StatusBadRequest, "Input schema 'many' is not one of 'trail', 'trails'"},
		{"not an admin", http.StatusForbidden,
			"User 'jane' doesn't have permission to publish rego policies in organization 'my-org'"},
		{"over a cap", http.StatusRequestEntityTooLarge,
			"Policy bundle is 1100000 bytes, over the 1048576 byte limit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, _ := newFakeServer(t, test.status, `{"message":"`+test.message+`"}`)

			version, err := newTestClient(t, server.URL, false).Publish("my-org", aPublishRequest())

			require.Nil(t, version)
			var apiError *requests.APIError
			require.True(t, errors.As(err, &apiError), "expected an API error, got %T", err)
			require.Equal(t, test.status, apiError.StatusCode)
			require.Equal(t, test.message, apiError.Message)
		})
	}
}

func TestPublishSendsNothingOnADryRun(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusAccepted, pendingBody)

	version, err := newTestClient(t, server.URL, true).Publish("my-org", aPublishRequest())

	require.NoError(t, err)
	require.Nil(t, version)
	require.Equal(t, 0, seen.hits)
}

func TestPublishKeepsTheIdentifierInOneSegment(t *testing.T) {
	for identifier, escaped := range map[string]string{
		"a/b":        "/api/v2/rego-policies/my-org/a%2Fb",
		"../other/x": "/api/v2/rego-policies/my-org/..%2Fother%2Fx",
	} {
		t.Run(identifier, func(t *testing.T) {
			server, seen := newFakeServer(t, http.StatusAccepted, pendingBody)
			request := aPublishRequest()
			request.Identifier = identifier

			_, err := newTestClient(t, server.URL, false).Publish("my-org", request)
			require.NoError(t, err)

			require.Equal(t, escaped, seen.escapedPath)
		})
	}
}

func TestPublishRefusesAnIdentifierThatIsNotASegment(t *testing.T) {
	for _, identifier := range []string{"", ".", ".."} {
		t.Run(identifier, func(t *testing.T) {
			server, seen := newFakeServer(t, http.StatusAccepted, pendingBody)
			request := aPublishRequest()
			request.Identifier = identifier

			version, err := newTestClient(t, server.URL, false).Publish("my-org", request)

			require.Nil(t, version)
			require.EqualError(t, err, fmt.Sprintf("%q is not a valid rego policy identifier", identifier))
			require.Equal(t, 0, seen.hits)
		})
	}
}

func TestPublishDoesNotRetryAConflict(t *testing.T) {
	const message = "Rego policy 'pr-approved' is being published by another request"
	server, seen := newFakeServer(t, http.StatusConflict, `{"message":"`+message+`"}`)
	httpClient, err := requests.NewKosliClient("", 2, false, logger.NewLogger(io.Discard, io.Discard, false))
	require.NoError(t, err)

	version, err := NewClient(httpClient, server.URL, "test-token", false).Publish("my-org", aPublishRequest())

	require.Nil(t, version)
	var apiError *requests.APIError
	require.True(t, errors.As(err, &apiError), "expected an API error, got %T: %v", err, err)
	require.Equal(t, http.StatusConflict, apiError.StatusCode)
	require.Equal(t, message, apiError.Message)
	require.Equal(t, 1, seen.hits)
}
