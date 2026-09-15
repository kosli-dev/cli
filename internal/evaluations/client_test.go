package evaluations

import (
	"encoding/json"
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
	hits        int
	method      string
	path        string
	authHeader  string
	contentType string
	body        map[string]interface{}
}

func newFakeServer(t *testing.T, status int, responseBody string) (*httptest.Server, *received) {
	t.Helper()
	seen := &received{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.hits++
		seen.method = r.Method
		seen.path = r.URL.Path
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

func aCreateRequest() CreateRequest {
	return CreateRequest{
		Trails: []TrailRef{{Flow: "release", Trail: "my-trail"}},
		Files:  map[string]string{"policy.rego": "package policy\n\nallow := true\n"},
		Params: map[string]interface{}{"threshold": float64(2)},
	}
}

const createdBody = `{"id":"01JABCDEF","status":"pending","requested_at":1757000000.5,"recorded_at":1757000000.5}`

func TestCreateSendsTheRequestTheAPIExpects(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, createdBody)

	_, err := newTestClient(t, server.URL, false).Create("my-org", aCreateRequest())
	require.NoError(t, err)

	require.Equal(t, 1, seen.hits)
	require.Equal(t, http.MethodPost, seen.method)
	require.Equal(t, "/api/v2/evaluations/my-org", seen.path)
	require.Equal(t, "Bearer test-token", seen.authHeader)
	require.Contains(t, seen.contentType, "application/json")
}

func TestCreateSendsExactlyTheAgreedPayload(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, createdBody)

	_, err := newTestClient(t, server.URL, false).Create("my-org", aCreateRequest())
	require.NoError(t, err)

	// Every model on the server forbids unknown fields, so an extra key here
	// is a 400 rather than something ignored.
	require.ElementsMatch(t, []string{"context", "policy", "params"}, keysOf(seen.body))

	context, ok := seen.body["context"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, []string{"trails"}, keysOf(context))
	require.Equal(t, []interface{}{
		map[string]interface{}{"flow": "release", "trail": "my-trail"},
	}, context["trails"])

	policy, ok := seen.body["policy"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, []string{"files"}, keysOf(policy))
	require.Equal(t, map[string]interface{}{
		"policy.rego": "package policy\n\nallow := true\n",
	}, policy["files"])

	require.Equal(t, map[string]interface{}{"threshold": float64(2)}, seen.body["params"])
}

func TestCreateSendsEveryTrailInOneEvaluation(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, createdBody)

	request := aCreateRequest()
	request.Trails = []TrailRef{
		{Flow: "release", Trail: "first"},
		{Flow: "release", Trail: "second"},
	}
	_, err := newTestClient(t, server.URL, false).Create("my-org", request)
	require.NoError(t, err)

	require.Equal(t, 1, seen.hits)
	context := seen.body["context"].(map[string]interface{})
	require.Equal(t, []interface{}{
		map[string]interface{}{"flow": "release", "trail": "first"},
		map[string]interface{}{"flow": "release", "trail": "second"},
	}, context["trails"])
}

func TestCreateSendsAbsentParamsAsAnEmptyObject(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, createdBody)

	request := aCreateRequest()
	request.Params = nil
	_, err := newTestClient(t, server.URL, false).Create("my-org", request)
	require.NoError(t, err)

	// Not null: the field is a plain dict on the server, so null fails
	// validation where an empty object is the documented default.
	require.Equal(t, map[string]interface{}{}, seen.body["params"])
}

func TestCreateDecodesTheCreatedEvaluation(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusCreated, createdBody)

	evaluation, err := newTestClient(t, server.URL, false).Create("my-org", aCreateRequest())
	require.NoError(t, err)

	require.Equal(t, "01JABCDEF", evaluation.ID)
	require.Equal(t, StatusPending, evaluation.Status)
	require.Equal(t, 1757000000.5, evaluation.RequestedAt)
	require.Equal(t, 1757000000.5, evaluation.RecordedAt)
	require.Nil(t, evaluation.Result)
	require.Nil(t, evaluation.Failure)
}

func TestCreateSurfacesTheServerMessageAndStatus(t *testing.T) {
	for _, test := range []struct {
		name    string
		status  int
		body    string
		message string
	}{
		{
			name:    "the organization is not entitled to the feature",
			status:  http.StatusForbidden,
			body:    `{"message":"Server-side evaluation is not enabled for this organization"}`,
			message: "Server-side evaluation is not enabled for this organization",
		},
		{
			name:    "a named trail does not exist",
			status:  http.StatusNotFound,
			body:    `{"message":"These trails do not exist in org 'my-org': release/nope"}`,
			message: "These trails do not exist in org 'my-org': release/nope",
		},
		{
			name:    "the policy bundle is over the byte cap",
			status:  http.StatusBadRequest,
			body:    `{"message":"policy bundle is 1048600 bytes, over the 1048576 byte limit"}`,
			message: "policy bundle is 1048600 bytes, over the 1048576 byte limit",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, _ := newFakeServer(t, test.status, test.body)

			evaluation, err := newTestClient(t, server.URL, false).Create("my-org", aCreateRequest())

			require.Nil(t, evaluation)
			require.Error(t, err)

			var apiError *requests.APIError
			require.True(t, errors.As(err, &apiError), "expected an API error, got %T", err)
			require.Equal(t, test.status, apiError.StatusCode)
			require.Equal(t, test.message, apiError.Message)
		})
	}
}

// A 5xx is retryable, so the shared HTTP client consumes it: it exhausts the
// retries and reports giving up, discarding the response and the server's
// message with it. Pinned because the enqueue failure the API documents at 503
// therefore cannot reach a user as its own sentence, and any caller wanting to
// name that failure has to say something of its own instead.
func TestCreateLosesTheServerMessageOnAServerError(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusServiceUnavailable,
		`{"message":"Evaluation '01JABCDEF' could not be queued"}`)

	evaluation, err := newTestClient(t, server.URL, false).Create("my-org", aCreateRequest())

	require.Nil(t, evaluation)
	require.Error(t, err)

	var apiError *requests.APIError
	require.False(t, errors.As(err, &apiError), "a 5xx does not arrive as an API error")
	require.NotContains(t, err.Error(), "could not be queued")
	require.Contains(t, err.Error(), "giving up after")
}

func TestCreateSendsNothingOnADryRun(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusCreated, createdBody)

	evaluation, err := newTestClient(t, server.URL, true).Create("my-org", aCreateRequest())

	require.NoError(t, err)
	require.Nil(t, evaluation)
	require.Equal(t, 0, seen.hits)
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// A body carrying the message key with something that is not text has no
// sentence to pass on. Reading one as a string used to panic the whole CLI,
// which is a stack trace in place of the error it arrived with.
func TestCreateSurvivesAMessageThatIsNotText(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "a null message", body: `{"message": null}`},
		{name: "a numeric message", body: `{"message": 42}`},
		{name: "an object message", body: `{"message": {"detail": "nested"}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, _ := newFakeServer(t, http.StatusForbidden, test.body)

			require.NotPanics(t, func() {
				_, err := newTestClient(t, server.URL, false).Create("my-org", aCreateRequest())
				require.Error(t, err)

				var apiError *requests.APIError
				require.True(t, errors.As(err, &apiError))
				require.Equal(t, http.StatusForbidden, apiError.StatusCode)
			})
		})
	}
}
