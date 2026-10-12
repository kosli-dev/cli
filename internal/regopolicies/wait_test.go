package regopolicies

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
)

type fakeResponse struct {
	status int
	body   string
}

// newSequencedServer repeats the last response once the sequence runs out.
func newSequencedServer(t *testing.T, responses ...fakeResponse) (*httptest.Server, *received) {
	t.Helper()
	seen := &received{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next := responses[min(seen.hits, len(responses)-1)]
		seen.hits++
		seen.path = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(next.status)
		_, err := w.Write([]byte(next.body))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	return server, seen
}

// testWaitOptions has a Timeout that expires only when a test shortens it.
func testWaitOptions() WaitOptions {
	return WaitOptions{
		Timeout: 2 * time.Second,
		Initial: time.Millisecond,
		Max:     2 * time.Millisecond,
	}
}

const unknownStatusBody = `{"identifier":"pr-approved","version":1,"status":"validating",` +
	`"digest":"sha256:abc123","input_schema":"trail","comment":"",` +
	`"created_at":1757000000.5,"created_by":"Jane Doe"}`

func TestWaitReturnsTheVersionOnceItIsValidated(t *testing.T) {
	server, seen := newSequencedServer(t,
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusOK, validBody},
	)

	version, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "pr-approved", 3, testWaitOptions())

	require.NoError(t, err)
	require.Equal(t, StatusValid, version.Status)
	require.Equal(t, 3, seen.hits)
	require.Equal(t, "/api/v2/rego-policies/my-org/pr-approved/versions/3", seen.path)
}

func TestWaitReturnsARejectedVersionAsAnOutcome(t *testing.T) {
	server, _ := newSequencedServer(t,
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusOK, rejectedBody},
	)

	version, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "pr-approved", 4, testWaitOptions())

	require.NoError(t, err)
	require.Equal(t, StatusRejected, version.Status)
	require.Equal(t, "compile", version.Error.Kind)
}

func TestWaitKeepsGoingThroughAnUnrecognisedStatus(t *testing.T) {
	server, seen := newSequencedServer(t,
		fakeResponse{http.StatusOK, unknownStatusBody},
		fakeResponse{http.StatusOK, validBody},
	)

	version, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "pr-approved", 3, testWaitOptions())

	require.NoError(t, err)
	require.Equal(t, StatusValid, version.Status)
	require.Equal(t, 2, seen.hits)
}

func TestWaitGivesUpNamingTheVersionAndHowToReadItLater(t *testing.T) {
	server, _ := newSequencedServer(t, fakeResponse{http.StatusOK, pendingBody})
	options := testWaitOptions()
	options.Timeout = 20 * time.Millisecond

	version, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "pr-approved", 1, options)

	require.Nil(t, version, "a wait that expired must never carry a version")
	var stillPending *StillPendingError
	require.True(t, errors.As(err, &stillPending), "expected a still-pending error, got %T", err)
	require.Equal(t, "my-org", stillPending.Org)
	require.Equal(t, "pr-approved", stillPending.Identifier)
	require.Equal(t, 1, stillPending.Version)
	require.GreaterOrEqual(t, stillPending.Waited+time.Millisecond, options.Timeout)
	require.Equal(t,
		"version 1 of rego policy pr-approved is still pending after "+stillPending.Waited.String()+
			"; read it later with GET /api/v2/rego-policies/my-org/pr-approved/versions/1",
		err.Error())
}

func TestWaitStopsOnAReadItCannotRetry(t *testing.T) {
	server, seen := newSequencedServer(t,
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusNotFound, `{"message":"Rego policy 'pr-approved' does not exist"}`},
	)

	version, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "pr-approved", 1, testWaitOptions())

	require.Nil(t, version)
	var apiError *requests.APIError
	require.True(t, errors.As(err, &apiError))
	require.Equal(t, http.StatusNotFound, apiError.StatusCode)
	require.Equal(t, 2, seen.hits)
}

func TestWaitOptionsFallBackToTheDefaultBudget(t *testing.T) {
	filled := WaitOptions{}.withDefaults()

	require.Equal(t, DefaultWaitOptions, filled)
	require.Equal(t, 30*time.Second, filled.Timeout)
}

func TestStillPendingNamesTheEscapedRoute(t *testing.T) {
	err := &StillPendingError{Org: "my-org", Identifier: "a/b", Version: 2, Waited: time.Second}

	require.Equal(t,
		"version 2 of rego policy a/b is still pending after 1s; "+
			"read it later with GET /api/v2/rego-policies/my-org/a%2Fb/versions/2",
		err.Error())
}
