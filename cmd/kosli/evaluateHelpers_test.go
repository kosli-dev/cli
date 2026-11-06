package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kosli-dev/cli/internal/evaluate"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
)

func TestEvaluateResultKeysListsEveryKeyTheOutputCanHave(t *testing.T) {
	var out bytes.Buffer
	result := &evaluate.Result{Allow: true}
	err := printEvaluateResult(&out, result, map[string]any{}, "json", true, map[string]any{}, false, "decision-id")
	require.NoError(t, err)

	var printed map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &printed))
	for key := range printed {
		require.Contains(t, evaluateResultKeys, key)
	}
}

func TestFetchRemotePolicyNamesTheStatusItWasRefusedWith(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	_, err := fetchRemotePolicy(server.URL + "/policy.rego")

	require.EqualError(t, err, "failed to fetch policy from "+server.URL+"/policy.rego: HTTP 404 Not Found")
}

func TestServerSideRequestErrorNamesTheStatus(t *testing.T) {
	require.EqualError(t,
		serverSideRequestError(&requests.APIError{StatusCode: http.StatusForbidden, Message: "not entitled", HasServerMessage: true}),
		"the Kosli server answered HTTP 403 Forbidden: not entitled")
	require.EqualError(t,
		serverSideRequestError(&requests.APIError{StatusCode: http.StatusNotFound}),
		"the Kosli server answered HTTP 404 Not Found")
}
