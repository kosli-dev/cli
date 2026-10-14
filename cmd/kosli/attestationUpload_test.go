package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/require"
)

// uploadFake is a server that plays both Kosli and the landing bucket: it
// mints an upload, takes the PUT and records the attestation. Each request is
// kept so a test can assert on what the CLI sent.
type uploadFake struct {
	*httpfake.HTTPFake
	mint    map[string]any
	put     *http.Request
	putBody []byte
	// putStatus and putResponse are what the bucket answers; 200 and empty
	// unless a test sets them.
	putStatus   int
	putResponse string
	attest      map[string]any
	attestReq   *http.Request
}

const (
	fakeUploadID  = "2ee4c953-e472-47af-bef3-999de17c"
	fakeChecksum  = "dQnlvaDHYtK6x/kNdYtbImP6Acy8VCq1498WO+CObKk="
	fakeUploadKey = "/uploads/cli-tests-landing/org/uploads/" + fakeUploadID + "/file1"
)

func newUploadFake(t *testing.T, org, flow, trail string) *uploadFake {
	t.Helper()
	f := &uploadFake{HTTPFake: httpfake.New(), putStatus: http.StatusOK}
	t.Cleanup(f.Close)

	f.NewHandler().Post("/api/v2/attestations/" + org + "/uploads").Handle(func(w http.ResponseWriter, r *http.Request, _ *httpfake.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&f.mint))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"upload_id":             fakeUploadID,
			"url":                   f.Server.URL + fakeUploadKey + "?X-Amz-Signature=abc",
			"method":                "PUT",
			"headers":               map[string]string{"x-amz-checksum-sha256": fakeChecksum},
			"upload_url_expires_at": 1791959246.0,
			"upload_id_expires_at":  1791962246.0,
			"max_size_bytes":        104857600,
		})
	})
	f.NewHandler().Put(fakeUploadKey).Handle(func(w http.ResponseWriter, r *http.Request, _ *httpfake.Request) {
		f.put = r
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		f.putBody = body
		w.WriteHeader(f.putStatus)
		_, _ = w.Write([]byte(f.putResponse))
	})
	f.NewHandler().Post(fmt.Sprintf("/api/v2/attestations/%s/%s/trail/%s/generic", org, flow, trail)).Handle(func(w http.ResponseWriter, r *http.Request, _ *httpfake.Request) {
		f.attestReq = r
		require.NoError(t, r.ParseMultipartForm(1<<20))
		require.NoError(t, json.Unmarshal([]byte(r.FormValue("data_json")), &f.attest))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	})
	return f
}

func TestAttestWithAttachmentMintsAnUploadForTheFile(t *testing.T) {
	fake := newUploadFake(t, "org", "flow", "trail")
	content, err := os.ReadFile("testdata/file1")
	require.NoError(t, err)

	_, _, _, _, err = executeCommandC(
		`attest generic --name foo --attachments testdata/file1 ` +
			`--flow flow --trail trail --org org --api-token tok --host ` + fake.Server.URL)

	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"filename":   "file1",
		"size_bytes": float64(len(content)),
		"sha256":     fmt.Sprintf("%x", sha256.Sum256(content)),
	}, fake.mint)
}

func TestAttestWithAttachmentPutsTheFileAsMinted(t *testing.T) {
	fake := newUploadFake(t, "org", "flow", "trail")
	content, err := os.ReadFile("testdata/file1")
	require.NoError(t, err)

	_, _, _, _, err = executeCommandC(
		`attest generic --name foo --attachments testdata/file1 ` +
			`--flow flow --trail trail --org org --api-token tok --host ` + fake.Server.URL)

	require.NoError(t, err)
	require.NotNil(t, fake.put, "the file was not PUT to the minted URL")
	require.Equal(t, "X-Amz-Signature=abc", fake.put.URL.RawQuery, "the signed query must reach the bucket untouched")
	require.Equal(t, fakeChecksum, fake.put.Header.Get("x-amz-checksum-sha256"))
	require.Empty(t, fake.put.Header.Get("Authorization"), "the Kosli token must not reach the bucket")
	require.Equal(t, int64(len(content)), fake.put.ContentLength)
	require.Equal(t, content, fake.putBody)
}

func TestAttestWithAttachmentSendsTheUploadIDAndNoFile(t *testing.T) {
	fake := newUploadFake(t, "org", "flow", "trail")

	_, _, _, _, err := executeCommandC(
		`attest generic --name foo --attachments testdata/file1 ` +
			`--flow flow --trail trail --org org --api-token tok --host ` + fake.Server.URL)

	require.NoError(t, err)
	require.NotNil(t, fake.attestReq, "the attestation was not recorded")
	require.Equal(t, fakeUploadID, fake.attest["attachment_upload_id"])
	require.Empty(t, fake.attestReq.MultipartForm.File, "the file must not travel with the attestation")
}

func TestAttestWithAttachmentFailsWhenMintIsRefused(t *testing.T) {
	fake := httpfake.New()
	t.Cleanup(fake.Close)
	fake.NewHandler().Post("/api/v2/attestations/org/uploads").Reply(404).
		BodyString(`{"message":"Direct uploads are not available on this server"}`)
	fake.NewHandler().Post("/api/v2/attestations/org/flow/trail/trail/generic").Handle(func(w http.ResponseWriter, r *http.Request, _ *httpfake.Request) {
		t.Error("nothing must be attested when the upload was refused")
	})

	_, _, _, _, err := executeCommandC(
		`attest generic --name foo --attachments testdata/file1 ` +
			`--flow flow --trail trail --org org --api-token tok --host ` + fake.Server.URL)

	require.EqualError(t, err, "Direct uploads are not available on this server")
}

func TestAttestWithAttachmentFailsWhenTheBucketRefusesTheFile(t *testing.T) {
	fake := newUploadFake(t, "org", "flow", "trail")
	fake.putStatus = http.StatusForbidden
	fake.putResponse = `<?xml version="1.0" encoding="UTF-8"?><Error><Code>BadDigest</Code>` +
		`<Message>The SHA256 you specified did not match what we received</Message></Error>`

	_, _, _, _, err := executeCommandC(
		`attest generic --name foo --attachments testdata/file1 ` +
			`--flow flow --trail trail --org org --api-token tok --host ` + fake.Server.URL)

	require.EqualError(t, err, "failed to upload attachment testdata/file1: "+
		"upload refused with HTTP 403: BadDigest: The SHA256 you specified did not match what we received")
	require.Nil(t, fake.attestReq, "nothing must be attested when the upload failed")
}
