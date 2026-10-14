package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedAttestationRequest struct {
	method   string
	path     string
	dataJSON map[string]any
	// files maps each file part's field name to its uploaded file name.
	files map[string]string
}

const (
	fakeUploadID  = "2ee4c953-e472-47af-bef3-999de17c"
	fakeChecksum  = "dQnlvaDHYtK6x/kNdYtbImP6Acy8VCq1498WO+CObKk="
	fakeUploadKey = "/uploads/landing/acme/uploads/" + fakeUploadID + "/file1"
)

// fakeAttestationServer plays both Kosli and the landing bucket: it mints an
// upload, takes the PUT and records the attestation.
type fakeAttestationServer struct {
	recorded []recordedAttestationRequest
	// status and body answer every attestation request; zero status means 201.
	status int
	body   string
	// mintStatus and mintBody answer the mint request; zero status means 201
	// with a URL back to this server.
	mintStatus int
	mintBody   string
	mint       map[string]any
	// put is the request that delivered the file, answered with putStatus and
	// putResponse; zero status means 200.
	put         *http.Request
	putBody     []byte
	putStatus   int
	putResponse string
}

func newFakeAttestationServer(t *testing.T) (*httptest.Server, *fakeAttestationServer) {
	t.Helper()
	fake := &fakeAttestationServer{}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The handler runs off the test goroutine, where require's FailNow
		// is unsafe, so failures are reported with assert and answered.
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/attestations/acme/uploads":
			fake.serveMint(t, w, r, server.URL)
		case r.Method == http.MethodPut:
			fake.servePut(t, w, r)
		default:
			fake.serveAttestation(t, w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server, fake
}

func (fake *fakeAttestationServer) serveMint(t *testing.T, w http.ResponseWriter, r *http.Request, serverURL string) {
	if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&fake.mint)) {
		http.Error(w, "unparseable mint request", http.StatusInternalServerError)
		return
	}
	if fake.mintStatus != 0 {
		w.WriteHeader(fake.mintStatus)
		_, _ = fmt.Fprint(w, fake.mintBody)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"upload_id":             fakeUploadID,
		"url":                   serverURL + fakeUploadKey + "?X-Amz-Signature=abc",
		"method":                "PUT",
		"headers":               map[string]string{"x-amz-checksum-sha256": fakeChecksum},
		"upload_url_expires_at": 1791959246.0,
		"upload_id_expires_at":  1791962246.0,
		"max_size_bytes":        104857600,
	})
}

func (fake *fakeAttestationServer) servePut(t *testing.T, w http.ResponseWriter, r *http.Request) {
	fake.put = r
	body, err := io.ReadAll(r.Body)
	if !assert.NoError(t, err) {
		http.Error(w, "unreadable body", http.StatusInternalServerError)
		return
	}
	fake.putBody = body
	status := fake.putStatus
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = fmt.Fprint(w, fake.putResponse)
}

func (fake *fakeAttestationServer) serveAttestation(t *testing.T, w http.ResponseWriter, r *http.Request) {
	if !assert.NoError(t, r.ParseMultipartForm(1<<20)) {
		http.Error(w, "unparseable form", http.StatusInternalServerError)
		return
	}
	req := recordedAttestationRequest{method: r.Method, path: r.URL.Path, files: map[string]string{}}
	if !assert.NoError(t, json.Unmarshal([]byte(r.FormValue("data_json")), &req.dataJSON)) {
		http.Error(w, "unparseable data_json", http.StatusInternalServerError)
		return
	}
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

func TestAttestationSubmitterMintsAnUploadForTheAttachment(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = []string{"testdata/file1"}
	content, err := os.ReadFile("testdata/file1")
	require.NoError(t, err)

	err = submitter.submit(submission)

	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"filename":   "file1",
		"size_bytes": float64(len(content)),
		"sha256":     fmt.Sprintf("%x", sha256.Sum256(content)),
	}, fake.mint)
}

func TestAttestationSubmitterPutsTheAttachmentAsMinted(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = []string{"testdata/file1"}
	content, err := os.ReadFile("testdata/file1")
	require.NoError(t, err)

	err = submitter.submit(submission)

	require.NoError(t, err)
	require.NotNil(t, fake.put, "the file was not PUT to the minted URL")
	require.Equal(t, fakeUploadKey, fake.put.URL.Path)
	require.Equal(t, "X-Amz-Signature=abc", fake.put.URL.RawQuery, "the signed query must reach the bucket untouched")
	require.Equal(t, fakeChecksum, fake.put.Header.Get("x-amz-checksum-sha256"))
	require.Empty(t, fake.put.Header.Get("Authorization"), "the Kosli token must not reach the bucket")
	require.Equal(t, int64(len(content)), fake.put.ContentLength)
	require.Equal(t, content, fake.putBody)
}

func TestAttestationSubmitterSendsTheUploadIDInsteadOfTheFile(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = []string{"testdata/file1"}

	err := submitter.submit(submission)

	require.NoError(t, err)
	require.Len(t, fake.recorded, 1)
	require.Equal(t, fakeUploadID, fake.recorded[0].dataJSON["attachment_upload_id"])
	require.Empty(t, fake.recorded[0].files, "the file must not travel with the attestation")
}

func TestAttestationSubmitterUploadsSeveralAttachmentsAsOneTarballAndRemovesIt(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = twoEvidenceFiles(t)

	err := submitter.submit(submission)

	require.NoError(t, err)
	require.Equal(t, "evidence.tgz", fake.mint["filename"])
	require.Equal(t, float64(len(fake.putBody)), fake.mint["size_bytes"])
	require.Equal(t, fmt.Sprintf("%x", sha256.Sum256(fake.putBody)), fake.mint["sha256"])
	require.Len(t, fake.recorded, 1)
	require.Empty(t, tarballsLeftIn(t, tmp))
}

func TestAttestationSubmitterAttestsNothingWhenMintIsRefused(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	fake.mintStatus = http.StatusNotFound
	fake.mintBody = `{"message":"Direct uploads are not available on this server"}`
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = []string{"testdata/file1"}

	err := submitter.submit(submission)

	require.EqualError(t, err, "Direct uploads are not available on this server")
	require.Nil(t, fake.put)
	require.Empty(t, fake.recorded)
}

func TestAttestationSubmitterAttestsNothingWhenTheBucketRefusesTheFile(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	fake.putStatus = http.StatusForbidden
	fake.putResponse = `<?xml version="1.0" encoding="UTF-8"?><Error><Code>BadDigest</Code>` +
		`<Message>The SHA256 you specified did not match what we received</Message></Error>`
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.attachments = []string{"testdata/file1"}

	err := submitter.submit(submission)

	require.EqualError(t, err, "failed to upload attachment testdata/file1: "+
		"upload refused with HTTP 403: BadDigest: The SHA256 you specified did not match what we received")
	require.Empty(t, fake.recorded)
}

func TestAttestationSubmitterShowsBothRequestsAndSendsNothingInDryRun(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, out := newTestAttestationSubmitter(t, server.URL, true)
	submission := genericSubmission("unit-tests")
	submission.attachments = []string{"testdata/file1"}

	err := submitter.submit(submission)

	require.NoError(t, err)
	require.Nil(t, fake.mint, "a dry run must not mint")
	require.Nil(t, fake.put, "a dry run must not upload")
	require.Empty(t, fake.recorded, "a dry run must not attest")
	require.Contains(t, out.String(), server.URL+"/api/v2/attestations/acme/uploads")
	require.Contains(t, out.String(), `"sha256": "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"`)
	require.Contains(t, out.String(), server.URL+"/api/v2/attestations/acme/my-flow/trail/my-trail/generic")
	require.Contains(t, out.String(), `"attachment_upload_id": "`+dryRunUploadID+`"`)
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

func TestAttestationSubmitterLogsTheRecordedAttestation(t *testing.T) {
	server, _ := newFakeAttestationServer(t)
	submitter, out := newTestAttestationSubmitter(t, server.URL, false)

	err := submitter.submit(genericSubmission("unit-tests"))

	require.NoError(t, err)
	require.Equal(t, "generic attestation 'unit-tests' is reported to trail: my-trail\n", out.String())
}

func TestAttestationSubmitterSendsNothingAndLogsNoSuccessInDryRun(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, out := newTestAttestationSubmitter(t, server.URL, true)

	err := submitter.submit(genericSubmission("unit-tests"))

	require.NoError(t, err)
	require.Empty(t, fake.recorded)
	require.Contains(t, out.String(), "THIS IS A DRY-RUN")
	require.NotContains(t, out.String(), "is reported to trail")
}

func TestAttestationSubmitterLogsNoSuccessWhenTheServerRejects(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	fake.status = http.StatusBadRequest
	fake.body = `{"message":"rejected"}`
	submitter, out := newTestAttestationSubmitter(t, server.URL, false)

	err := submitter.submit(genericSubmission("unit-tests"))

	require.EqualError(t, err, "rejected")
	require.NotContains(t, out.String(), "is reported to trail")
}

func TestAttestationSubmitterExplainsHowToBindTheAttestation(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	fake.status = http.StatusBadRequest
	fake.body = `{"message":"Input payload requires at least one of: artifact_fingerprint or git_commit_info."}`
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)

	err := submitter.submit(genericSubmission("unit-tests"))

	require.EqualError(t, err, "Input payload requires at least one of: specifying the fingerprint "+
		"(either by calculating it using the artifact name/path and --artifact-type, or by providing it using --fingerprint) "+
		"or providing --commit (requires an available git repo to access commit details)")
}

func TestAttestationSubmitterReportsAssertFailuresAfterRecording(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, out := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.assertFailures = []error{errors.New("no references found"), errors.New("1 reference not found")}

	err := submitter.submit(submission)

	require.EqualError(t, err, "no references found\nError: 1 reference not found")
	require.Len(t, fake.recorded, 1)
	require.Equal(t, "generic attestation 'unit-tests' is reported to trail: my-trail\n", out.String())
}

func TestAttestationSubmitterJoinsAssertFailuresAfterTheServerError(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	fake.status = http.StatusBadRequest
	fake.body = `{"message":"rejected"}`
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.assertFailures = []error{errors.New("no references found")}

	err := submitter.submit(submission)

	require.EqualError(t, err, "rejected\nError: no references found")
}

func TestAttestationSubmitterDropsAssertFailuresInDryRun(t *testing.T) {
	server, _ := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, true)
	submission := genericSubmission("unit-tests")
	submission.assertFailures = []error{errors.New("no references found")}

	err := submitter.submit(submission)

	require.NoError(t, err)
}

func TestAttestationSubmitterUploadsEvidenceAlreadyInMemory(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.evidence = &requests.FileBytes{Name: "sbom.json", Data: []byte(`{"bomFormat":"CycloneDX"}`)}

	err := submitter.submit(submission)

	require.NoError(t, err)
	require.Len(t, fake.recorded, 1)
	require.Equal(t, "unit-tests", fake.recorded[0].dataJSON["attestation_name"])
	require.Equal(t, map[string]string{"attachment_file": "sbom.json"}, fake.recorded[0].files)
}

func TestAttestationSubmitterRemovesTheStagingDirWhenAnAttachmentIsMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	attachments := twoEvidenceFiles(t)
	submission.attachments = []string{attachments[0], filepath.Join(filepath.Dir(attachments[0]), "missing.txt")}

	err := submitter.submit(submission)

	require.ErrorContains(t, err, "missing.txt")
	require.Empty(t, fake.recorded)
	left, readErr := os.ReadDir(tmp)
	require.NoError(t, readErr)
	require.Empty(t, left, "the staging directory must not outlive a failed copy")
}

func TestAttestationSubmitterRejectsEvidenceTogetherWithAttachments(t *testing.T) {
	server, fake := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, server.URL, false)
	submission := genericSubmission("unit-tests")
	submission.evidence = &requests.FileBytes{Name: "sbom.json", Data: []byte(`{}`)}
	submission.attachments = twoEvidenceFiles(t)

	err := submitter.submit(submission)

	require.Error(t, err)
	require.Empty(t, fake.recorded)
}
