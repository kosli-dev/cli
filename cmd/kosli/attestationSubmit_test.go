package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
)

type recordedAttestationRequest struct {
	method   string
	path     string
	dataJSON map[string]any
	// files maps each file part's field name to its uploaded file name.
	files map[string]string
}

type fakeAttestationServer struct {
	recorded []recordedAttestationRequest
	// status and body answer every request; zero status means 201.
	status int
	body   string
}

func newFakeAttestationServer(t *testing.T) (*httptest.Server, *fakeAttestationServer) {
	t.Helper()
	fake := &fakeAttestationServer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(1<<20))
		req := recordedAttestationRequest{method: r.Method, path: r.URL.Path, files: map[string]string{}}
		require.NoError(t, json.Unmarshal([]byte(r.FormValue("data_json")), &req.dataJSON))
		for field, headers := range r.MultipartForm.File {
			req.files[field] = headers[0].Filename
		}
		fake.recorded = append(fake.recorded, req)
		status := fake.status
		if status == 0 {
			status = http.StatusCreated
		}
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, fake.body)
	}))
	t.Cleanup(server.Close)
	return server, fake
}

func genericSubmission(name string) attestationSubmission {
	return attestationSubmission{
		flow:    "my-flow",
		trail:   "my-trail",
		slug:    "generic",
		label:   "generic",
		payload: &GenericAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{AttestationName: name}},
	}
}

// tarballsLeftIn lists evidence tarballs still on disk under tmp, where
// utils.Tar creates them when TMPDIR points at tmp.
func tarballsLeftIn(t *testing.T, tmp string) []string {
	t.Helper()
	left, err := filepath.Glob(filepath.Join(tmp, "*", "evidence.tgz"))
	require.NoError(t, err)
	return left
}

func twoEvidenceFiles(t *testing.T) []string {
	t.Helper()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	require.NoError(t, os.WriteFile(a, []byte("a"), 0o600))
	require.NoError(t, os.WriteFile(b, []byte("b"), 0o600))
	return []string{a, b}
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
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)

	err := submitter.submit(genericSubmission("unit-tests"))

	require.NoError(t, err)
	require.Len(t, fake.recorded, 1)
	require.Equal(t, http.MethodPost, fake.recorded[0].method)
	require.Equal(t, "/api/v2/attestations/acme/my-flow/trail/my-trail/generic", fake.recorded[0].path)
}

func TestAttestationSubmitterSendsThePayloadAsDataJSONWithoutAttachments(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)

	err := submitter.submit(genericSubmission("unit-tests"))

	require.NoError(t, err)
	require.Len(t, fake.recorded, 1)
	require.Equal(t, "unit-tests", fake.recorded[0].dataJSON["attestation_name"])
	require.Empty(t, fake.recorded[0].files)
}

func TestAttestationSubmitterUploadsAttachmentsAndRemovesTheTarball(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = twoEvidenceFiles(t)

	err := submitter.submit(submission)

	require.NoError(t, err)
	require.Len(t, fake.recorded, 1)
	require.Equal(t, map[string]string{"attachment_file": "evidence.tgz"}, fake.recorded[0].files)
	require.Empty(t, tarballsLeftIn(t, tmp))
}

func TestAttestationSubmitterRemovesTheTarballWhenTheServerRejects(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	server, fake := newFakeAttestationServer(t)
	fake.status = http.StatusBadRequest
	fake.body = `{"message":"rejected"}`
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = twoEvidenceFiles(t)

	err := submitter.submit(submission)

	require.EqualError(t, err, "rejected")
	require.Empty(t, tarballsLeftIn(t, tmp))
}
