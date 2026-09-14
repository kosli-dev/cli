package evaluations

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

// newSequencedServer answers each request with the next response in turn, and
// repeats the last one once the sequence runs out. That is what lets a test
// say "pending, pending, then completed" without counting requests itself.
func newSequencedServer(t *testing.T, responses ...fakeResponse) (*httptest.Server, *received) {
	t.Helper()
	seen := &received{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next := responses[min(seen.hits, len(responses)-1)]
		seen.hits++
		seen.method = r.Method
		seen.path = r.URL.Path
		seen.authHeader = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(next.status)
		_, err := w.Write([]byte(next.body))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	return server, seen
}

// Short enough that a test that must wait the budget out finishes quickly, and
// long enough that a test that must not hit the budget never does.
func testWaitOptions() WaitOptions {
	return WaitOptions{
		Timeout: 2 * time.Second,
		Initial: time.Millisecond,
		Max:     2 * time.Millisecond,
	}
}

const (
	pendingBody   = `{"id":"01JABCDEF","status":"pending","requested_at":1757000000.5,"recorded_at":1757000000.5}`
	allowedBody   = `{"id":"01JABCDEF","status":"completed","requested_at":1757000000.5,"recorded_at":1757000000.5,"result":{"allow":true,"violations":[]}}`
	deniedBody    = `{"id":"01JABCDEF","status":"completed","requested_at":1757000000.5,"recorded_at":1757000000.5,"result":{"allow":false,"violations":["change is not approved","no snyk scan"]}}`
	brokenBody    = `{"id":"01JABCDEF","status":"failed","requested_at":1757000000.5,"recorded_at":1757000000.5,"error":{"kind":"compile","message":"policy does not compile: policy.rego:4: rego_parse_error"}}`
	unknownBody   = `{"id":"01JABCDEF","status":"running","requested_at":1757000000.5,"recorded_at":1757000000.5}`
	noViolations  = `{"id":"01JABCDEF","status":"completed","requested_at":1757000000.5,"recorded_at":1757000000.5,"result":{"allow":false}}`
	notFoundBody  = `{"message":"Evaluation '01JABCDEF' does not exist in org 'my-org'"}`
	emptyListBody = `{"id":"01JABCDEF","status":"completed","requested_at":1757000000.5,"recorded_at":1757000000.5,"result":{"allow":true,"violations":[]}}`
)

func TestGetReadsTheEvaluationBack(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, deniedBody)

	evaluation, err := newTestClient(t, server.URL, false).Get("my-org", "01JABCDEF")
	require.NoError(t, err)

	require.Equal(t, http.MethodGet, seen.method)
	require.Equal(t, "/api/v2/evaluations/my-org/01JABCDEF", seen.path)
	require.Equal(t, "Bearer test-token", seen.authHeader)

	require.Equal(t, StatusCompleted, evaluation.Status)
	require.True(t, evaluation.IsTerminal())
	require.NotNil(t, evaluation.Result)
	require.False(t, evaluation.Result.Allow)
	require.Equal(t, []string{"change is not approved", "no snyk scan"}, evaluation.Result.Violations)
	require.Nil(t, evaluation.Failure)
}

func TestGetReadsAClassifiedFailure(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusOK, brokenBody)

	evaluation, err := newTestClient(t, server.URL, false).Get("my-org", "01JABCDEF")
	require.NoError(t, err)

	require.Equal(t, StatusFailed, evaluation.Status)
	require.True(t, evaluation.IsTerminal())
	require.Nil(t, evaluation.Result)
	require.NotNil(t, evaluation.Failure)
	require.Equal(t, "compile", evaluation.Failure.Kind)
	require.Contains(t, evaluation.Failure.Message, "does not compile")
}

// The two shapes the API can use for "nothing was violated" are kept apart
// here rather than flattened, because the local evaluator prints a missing
// list as null and an empty list as []. Whichever the server sends, the
// command has to reconcile it before printing, or the two paths disagree on
// the page while agreeing on the verdict.
func TestGetKeepsTheViolationsShapeItWasSent(t *testing.T) {
	t.Run("an absent list stays absent", func(t *testing.T) {
		server, _ := newFakeServer(t, http.StatusOK, noViolations)

		evaluation, err := newTestClient(t, server.URL, false).Get("my-org", "01JABCDEF")
		require.NoError(t, err)
		require.Nil(t, evaluation.Result.Violations)
	})

	t.Run("an empty list stays empty and present", func(t *testing.T) {
		server, _ := newFakeServer(t, http.StatusOK, emptyListBody)

		evaluation, err := newTestClient(t, server.URL, false).Get("my-org", "01JABCDEF")
		require.NoError(t, err)
		require.NotNil(t, evaluation.Result.Violations)
		require.Empty(t, evaluation.Result.Violations)
	})
}

func TestGetSurfacesAMissingEvaluation(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusNotFound, notFoundBody)

	evaluation, err := newTestClient(t, server.URL, false).Get("my-org", "01JABCDEF")

	require.Nil(t, evaluation)
	var apiError *requests.APIError
	require.True(t, errors.As(err, &apiError))
	require.Equal(t, http.StatusNotFound, apiError.StatusCode)
	require.Equal(t, "Evaluation '01JABCDEF' does not exist in org 'my-org'", apiError.Message)
}

func TestWaitReturnsTheVerdictOnceItArrives(t *testing.T) {
	server, seen := newSequencedServer(t,
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusOK, allowedBody},
	)

	evaluation, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "01JABCDEF", testWaitOptions())

	require.NoError(t, err)
	require.Equal(t, StatusCompleted, evaluation.Status)
	require.True(t, evaluation.Result.Allow)
	require.Equal(t, 3, seen.hits)
}

func TestWaitReturnsAClassifiedFailureAsATerminalOutcome(t *testing.T) {
	server, _ := newSequencedServer(t,
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusOK, brokenBody},
	)

	evaluation, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "01JABCDEF", testWaitOptions())

	require.NoError(t, err)
	require.Equal(t, StatusFailed, evaluation.Status)
	require.Equal(t, "compile", evaluation.Failure.Kind)
}

// An unrecognised status means the server has added one; reading it as
// terminal would turn a running evaluation into a verdict.
func TestWaitKeepsGoingThroughAnUnrecognisedStatus(t *testing.T) {
	server, seen := newSequencedServer(t,
		fakeResponse{http.StatusOK, unknownBody},
		fakeResponse{http.StatusOK, unknownBody},
		fakeResponse{http.StatusOK, deniedBody},
	)

	evaluation, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "01JABCDEF", testWaitOptions())

	require.NoError(t, err)
	require.False(t, evaluation.Result.Allow)
	require.Equal(t, 3, seen.hits)
}

func TestWaitGivesUpWithTheEvaluationIdAndNoVerdict(t *testing.T) {
	server, _ := newSequencedServer(t, fakeResponse{http.StatusOK, pendingBody})

	options := testWaitOptions()
	options.Timeout = 20 * time.Millisecond

	evaluation, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "01JABCDEF", options)

	require.Nil(t, evaluation, "a wait that expired must never carry a verdict")

	var stillPending *StillPendingError
	require.True(t, errors.As(err, &stillPending), "expected a still-pending error, got %T", err)
	require.Equal(t, "01JABCDEF", stillPending.ID)
	require.Contains(t, err.Error(), "01JABCDEF")
	require.Contains(t, err.Error(), "still pending")
	require.NotContains(t, err.Error(), "denied")
}

func TestWaitStopsWhenTheCallerGivesUp(t *testing.T) {
	server, seen := newSequencedServer(t, fakeResponse{http.StatusOK, pendingBody})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	evaluation, err := newTestClient(t, server.URL, false).
		WaitForTerminal(ctx, "my-org", "01JABCDEF", testWaitOptions())

	require.Nil(t, evaluation)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 0, seen.hits, "a cancelled wait asks the server for nothing")
}

func TestWaitStopsOnAReadItCannotRetry(t *testing.T) {
	server, seen := newSequencedServer(t,
		fakeResponse{http.StatusOK, pendingBody},
		fakeResponse{http.StatusNotFound, notFoundBody},
	)

	evaluation, err := newTestClient(t, server.URL, false).
		WaitForTerminal(context.Background(), "my-org", "01JABCDEF", testWaitOptions())

	require.Nil(t, evaluation)
	var apiError *requests.APIError
	require.True(t, errors.As(err, &apiError))
	require.Equal(t, http.StatusNotFound, apiError.StatusCode)
	require.Equal(t, 2, seen.hits, "a failed read ends the wait rather than being polled through")
}

func TestNextIntervalDoublesUpToTheCeiling(t *testing.T) {
	const max = 5 * time.Second

	for _, test := range []struct {
		current time.Duration
		want    time.Duration
	}{
		{500 * time.Millisecond, time.Second},
		{time.Second, 2 * time.Second},
		{2 * time.Second, 4 * time.Second},
		{4 * time.Second, max},
		{max, max},
	} {
		require.Equal(t, test.want, nextInterval(test.current, max),
			"doubling %s", test.current)
	}
}

func TestWaitOptionsFallBackToTheAgreedBudget(t *testing.T) {
	filled := WaitOptions{}.withDefaults()

	require.Equal(t, DefaultWaitOptions, filled)
	require.Equal(t, 30*time.Second, filled.Timeout, "the enqueue-to-terminal ceiling")
}
