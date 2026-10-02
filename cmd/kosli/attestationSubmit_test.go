package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
)

type recordedAttestationRequest struct {
	method string
	path   string
}

func newFakeAttestationServer(t *testing.T) (*httptest.Server, *[]recordedAttestationRequest) {
	t.Helper()
	var recorded []recordedAttestationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded = append(recorded, recordedAttestationRequest{method: r.Method, path: r.URL.Path})
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)
	return server, &recorded
}

func newTestAttestationSubmitter(t *testing.T, host string, dryRun bool) (*attestationSubmitter, *bytes.Buffer) {
	t.Helper()
	var out bytes.Buffer
	l := log.NewLogger(&out, &out, false)
	client, err := requests.NewKosliClient("", 0, false, l)
	require.NoError(t, err)
	return &attestationSubmitter{
		client: client,
		host:   host,
		org:    "acme",
		token:  "secret",
		dryRun: dryRun,
		logger: l,
	}, &out
}

func TestAttestationSubmitterPostsToTheAttestationTypeURL(t *testing.T) {
	server, recorded := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)

	err := submitter.submit(attestationSubmission{
		flow:    "my-flow",
		trail:   "my-trail",
		slug:    "generic",
		label:   "generic",
		payload: &GenericAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{AttestationName: "unit-tests"}},
	})

	require.NoError(t, err)
	require.Len(t, *recorded, 1)
	require.Equal(t, http.MethodPost, (*recorded)[0].method)
	require.Equal(t, "/api/v2/attestations/acme/my-flow/trail/my-trail/generic", (*recorded)[0].path)
}
