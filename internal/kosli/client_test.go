package kosli

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
)

// received is what the fake server saw, so a test can assert on the request
// rather than only on what came back.
type received struct {
	hits       int
	method     string
	rawPath    string
	authHeader string
}

func newFakeServer(t *testing.T, status int, responseBody string) (*httptest.Server, *received) {
	t.Helper()
	seen := &received{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.hits++
		seen.method = r.Method
		seen.rawPath = r.URL.EscapedPath()
		seen.authHeader = r.Header.Get("Authorization")
		w.WriteHeader(status)
		_, err := w.Write([]byte(responseBody))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	return server, seen
}

func newHTTPClient(t *testing.T) *requests.Client {
	t.Helper()
	// No retries, so a 409 or 5xx surfaces as itself instead of being retried away.
	httpClient, err := requests.NewKosliClient("", 0, false, logger.NewLogger(io.Discard, io.Discard, false))
	require.NoError(t, err)
	return httpClient
}

func newTestClient(t *testing.T, host string) *Client {
	t.Helper()
	return New(newHTTPClient(t), host, "test-org", "test-token")
}

func TestSendRequestsTheEndpointWithTheTokenAndReturnsTheRawBody(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `{"name":"trail-1"}`)
	client := newTestClient(t, server.URL)

	endpoint, err := client.endpoint("trails", "test-org", "flow-1", "trail-1")
	require.NoError(t, err)
	result, err := client.send(context.Background(), &requests.RequestParams{Method: http.MethodGet, URL: endpoint})

	require.NoError(t, err)
	require.Equal(t, http.MethodGet, seen.method)
	require.Equal(t, "/api/v2/trails/test-org/flow-1/trail-1", seen.rawPath)
	require.Equal(t, "Bearer test-token", seen.authHeader)
	require.JSONEq(t, `{"name":"trail-1"}`, string(result.Raw))
}

func TestSendReportsA201AsCreated(t *testing.T) {
	for _, tc := range []struct {
		status  int
		created bool
	}{
		{http.StatusCreated, true},
		{http.StatusOK, false},
	} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			server, _ := newFakeServer(t, tc.status, `"OK"`)
			client := newTestClient(t, server.URL)

			endpoint, err := client.endpoint("flows", "test-org")
			require.NoError(t, err)
			result, err := client.send(context.Background(), &requests.RequestParams{Method: http.MethodPut, URL: endpoint})

			require.NoError(t, err)
			require.Equal(t, tc.created, result.Created)
		})
	}
}

func TestEndpointEscapesEachPathSegment(t *testing.T) {
	client := New(nil, "https://app.kosli.com", "test-org", "test-token")

	endpoint, err := client.endpoint("trails", "test-org", "my flow", "trail?#1")

	require.NoError(t, err)
	require.Equal(t, "https://app.kosli.com/api/v2/trails/test-org/my%20flow/trail%3F%231", endpoint)
}

func TestSendReturnsAnAPIErrorThatTheStatusHelpersRecognise(t *testing.T) {
	for _, tc := range []struct {
		status int
		is     func(error) bool
	}{
		{http.StatusNotFound, IsNotFound},
		// A 409 is retried until retries run out, which ends in a plain error,
		// so only a request that opts out of conflict retries sees it.
		{http.StatusConflict, IsConflict},
		{http.StatusForbidden, IsForbidden},
	} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			server, _ := newFakeServer(t, tc.status, `{"message":"nope"}`)
			client := newTestClient(t, server.URL)

			endpoint, err := client.endpoint("flows", "test-org", "missing")
			require.NoError(t, err)
			_, err = client.send(context.Background(), &requests.RequestParams{
				Method:               http.MethodGet,
				URL:                  endpoint,
				DisableConflictRetry: true,
			})

			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			require.Equal(t, tc.status, apiErr.StatusCode)
			require.True(t, tc.is(err))
		})
	}
}

func TestStatusHelpersDoNotMatchOtherStatusesOrOtherErrors(t *testing.T) {
	serverError := &APIError{StatusCode: http.StatusInternalServerError}
	for _, is := range []func(error) bool{IsNotFound, IsConflict, IsForbidden} {
		require.False(t, is(serverError))
		require.False(t, is(errors.New("connection refused")))
		require.False(t, is(nil))
	}
}
