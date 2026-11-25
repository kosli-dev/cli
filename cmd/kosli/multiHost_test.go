package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/kosli"
	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const multiHostFingerprint = "1bef738d0bb1e690500f99a5b57d958caf3a5eb3e00d9012e1f4369fc6812e01"

// recordingKosliHost records the requests a command makes and answers each with
// the body given for its path, or {} with 200. statuses, keyed by method and
// path, answers with another status.
type recordingKosliHost struct {
	mu       sync.Mutex
	requests []string
	payloads map[string]string
	bodies   map[string]string
	statuses map[string]int
}

func newRecordingKosliHost(t *testing.T, bodies map[string]string) (*httptest.Server, *recordingKosliHost) {
	t.Helper()
	fake := &recordingKosliHost{bodies: bodies, payloads: map[string]string{}, statuses: map[string]int{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		payload, _ := io.ReadAll(r.Body)
		fake.mu.Lock()
		fake.requests = append(fake.requests, key)
		fake.payloads[key] = string(payload)
		status, failing := fake.statuses[key]
		fake.mu.Unlock()
		body, ok := fake.bodies[r.URL.Path]
		if !ok {
			body = "{}"
		}
		w.Header().Set("Content-Type", "application/json")
		if failing {
			w.WriteHeader(status)
		}
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server, fake
}

func multiHostFlags(a, b string) string {
	return fmt.Sprintf("--host %s,%s --api-token one,two --org acme --max-api-retries 0", a, b)
}

func newTwoHostKosli(t *testing.T, a, b string) (*kosli.Kosli, *bytes.Buffer) {
	t.Helper()
	var out bytes.Buffer
	l := log.NewLogger(&out, &out, false)
	client, err := requests.NewKosliClient("", 0, false, l)
	require.NoError(t, err)
	return kosli.New(client, l, a+","+b, "one,two", false), &out
}

func TestAReadAsksOnlyThePrimaryHost(t *testing.T) {
	serverA, fakeA := newRecordingKosliHost(t, nil)
	serverB, fakeB := newRecordingKosliHost(t, nil)

	_, _, _, _, err := executeCommandC("get flow my-flow " + multiHostFlags(serverA.URL, serverB.URL))

	require.NoError(t, err)
	assert.Equal(t, []string{"GET /api/v2/flows/acme/my-flow"}, fakeA.requests)
	assert.Empty(t, fakeB.requests)
}

func TestAWriteIsSentToEveryHost(t *testing.T) {
	serverA, fakeA := newRecordingKosliHost(t, nil)
	serverB, fakeB := newRecordingKosliHost(t, nil)

	_, combined, _, _, err := executeCommandC("archive flow my-flow " + multiHostFlags(serverA.URL, serverB.URL))

	require.NoError(t, err)
	assert.Equal(t, []string{"PUT /api/v2/flows/acme/my-flow/archive"}, fakeA.requests)
	assert.Equal(t, fakeA.requests, fakeB.requests)
	assert.Equal(t, "flow my-flow was archived\n", combined)
}

func TestAnAssertFailsWhenASecondaryHostDisagrees(t *testing.T) {
	path := "/api/v2/asserts/acme/fingerprint/" + multiHostFingerprint
	serverA, _ := newRecordingKosliHost(t, map[string]string{path: `{"compliant": true}`})
	serverB, _ := newRecordingKosliHost(t, map[string]string{path: `{"compliant": false}`})

	_, _, _, _, err := executeCommandC("assert artifact --fingerprint " + multiHostFingerprint + " --output json " +
		multiHostFlags(serverA.URL, serverB.URL))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "["+serverB.URL+"]\nArtifact is not compliant")
}

func TestATagLooksUpTheRepoIDOnEachHost(t *testing.T) {
	serverA, fakeA := newRecordingKosliHost(t, map[string]string{"/api/v2/repos/acme/my-repo": `{"id": "id-on-a"}`})
	serverB, fakeB := newRecordingKosliHost(t, map[string]string{"/api/v2/repos/acme/my-repo": `{"id": "id-on-b"}`})

	_, _, _, _, err := executeCommandC("tag repo my-repo --set team=a " + multiHostFlags(serverA.URL, serverB.URL))

	require.NoError(t, err)
	assert.Contains(t, fakeA.requests, "PATCH /api/v2/tags/acme/repo/id-on-a")
	assert.Contains(t, fakeB.requests, "PATCH /api/v2/tags/acme/repo/id-on-b")
}

func TestAnEmptyHostInTheListIsRefused(t *testing.T) {
	_, _, _, _, err := executeCommandC("status --host http://a,,http://b --api-token 1,2,3")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag '--host' was given an empty value")
}

func TestAttestationSubmitterSendsTheSameAttestationToEveryHost(t *testing.T) {
	serverA, fakeA := newFakeAttestationServer(t)
	serverB, fakeB := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, serverA.URL, false)
	submitter.kosli, _ = newTwoHostKosli(t, serverA.URL, serverB.URL)
	submission := genericSubmission("unit-tests")
	submission.attachments = []string{"testdata/file1", "testdata/folder1"}

	require.NoError(t, submitter.submit(submission))

	require.Len(t, fakeA.recorded, 1)
	require.Len(t, fakeB.recorded, 1)
	assert.Equal(t, fakeA.recorded[0].dataJSON, fakeB.recorded[0].dataJSON)
	assert.Equal(t, fakeA.putBody, fakeB.putBody, "every host is sent the one tarball")
	assert.Empty(t, submission.payload.(*GenericAttestationPayload).AttachmentUploadID,
		"each host's upload id goes in a copy of the payload")
}

func TestAttestationSubmitterNamesTheHostThatRejectedTheAttestation(t *testing.T) {
	serverA, _ := newFakeAttestationServer(t)
	serverB, fakeB := newFakeAttestationServer(t)
	fakeB.status = http.StatusBadRequest
	fakeB.body = `{"message": "no such flow"}`
	submitter, _ := newTestAttestationSubmitter(t, serverA.URL, false)
	submitter.kosli, _ = newTwoHostKosli(t, serverA.URL, serverB.URL)

	err := submitter.submit(genericSubmission("unit-tests"))

	assert.EqualError(t, err, "["+serverB.URL+"]\nno such flow")
}

func TestAttestationSubmitterLogsOnlyThePrimarysSuccess(t *testing.T) {
	serverA, _ := newFakeAttestationServer(t)
	serverB, _ := newFakeAttestationServer(t)
	submitter, _ := newTestAttestationSubmitter(t, serverA.URL, false)
	var out *bytes.Buffer
	submitter.kosli, out = newTwoHostKosli(t, serverA.URL, serverB.URL)

	require.NoError(t, submitter.submit(genericSubmission("unit-tests")))

	assert.Equal(t, "generic attestation 'unit-tests' is reported to trail: my-trail\n", out.String())
}

func TestSnapshotReporterCollectsOnceAndReportsToEveryHost(t *testing.T) {
	serverA, fakeA := newFakeSnapshotServer(t)
	serverB, fakeB := newFakeSnapshotServer(t)
	reporter, ensured, _ := newTestSnapshotReporter(t, serverA.URL, false)
	reporter.kosli, _ = newTwoHostKosli(t, serverA.URL, serverB.URL)
	var collected atomic.Int32

	err := reporter.report("prod", "docker", func() (any, string, error) {
		collected.Add(1)
		return twoContainers()
	})

	require.NoError(t, err)
	assert.Equal(t, int32(1), collected.Load())
	assert.Equal(t, []string{"prod/docker", "prod/docker"}, ensured.calls)
	require.Len(t, fakeA.recorded, 1)
	require.Len(t, fakeB.recorded, 1)
	assert.Equal(t, fakeA.recorded[0].body, fakeB.recorded[0].body)
}

func TestSnapshotReporterStillReportsToTheHostsItCouldPrepare(t *testing.T) {
	serverA, fakeA := newFakeSnapshotServer(t)
	serverB, fakeB := newFakeSnapshotServer(t)
	reporter, _, _ := newTestSnapshotReporter(t, serverA.URL, false)
	reporter.kosli, _ = newTwoHostKosli(t, serverA.URL, serverB.URL)
	reporter.ensure = func(h kosli.Host, envName, envType string) error {
		if h.URL == serverA.URL {
			return errors.New("environment prod already exists with type K8S")
		}
		return nil
	}

	err := reporter.report("prod", "docker", twoContainers)

	assert.EqualError(t, err, "["+serverA.URL+"]\nenvironment prod already exists with type K8S")
	assert.Empty(t, fakeA.recorded)
	assert.Len(t, fakeB.recorded, 1)
}

func TestSnapshotReporterDoesNotCollectWhenNoHostIsReady(t *testing.T) {
	reporter, _, _ := newTestSnapshotReporter(t, "http://a", false)
	reporter.kosli, _ = newTwoHostKosli(t, "http://a", "http://b")
	reporter.ensure = func(kosli.Host, string, string) error { return errors.New("unauthorized") }

	err := reporter.report("prod", "docker", func() (any, string, error) {
		t.Error("collected a snapshot no host could take")
		return nil, "", nil
	})

	assert.EqualError(t, err, "[http://a]\nunauthorized\n[http://b]\nunauthorized")
}

func TestSnapshotReporterReportsAFailedCollectionOnce(t *testing.T) {
	reporter, _, _ := newTestSnapshotReporter(t, "http://a", false)
	reporter.kosli, _ = newTwoHostKosli(t, "http://a", "http://b")

	err := reporter.report("prod", "docker", func() (any, string, error) {
		return nil, "", errors.New("docker is not running")
	})

	assert.EqualError(t, err, "docker is not running")
}

func TestWithAttachmentUploadIDLeavesTheOriginalPayloadAlone(t *testing.T) {
	original := GenericAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{AttestationName: "x"}}

	cloned := withAttachmentUploadID(original, "id-1")

	assert.Empty(t, original.AttachmentUploadID)
	assert.Equal(t, "id-1", cloned.(GenericAttestationPayload).AttachmentUploadID)
	assert.Equal(t, "x", cloned.attestationName())
}

func TestATagSendsEveryHostTheTagsToRemoveAsGiven(t *testing.T) {
	serverA, fakeA := newRecordingKosliHost(t, nil)
	serverB, fakeB := newRecordingKosliHost(t, nil)

	_, combined, _, _, err := executeCommandC("tag flow my-flow --unset t2 --unset t1 --unset t3 " +
		multiHostFlags(serverA.URL, serverB.URL))

	require.NoError(t, err)
	key := "PATCH /api/v2/tags/acme/flow/my-flow"
	assert.Contains(t, fakeA.payloads[key], `"remove_tags":["t2","t1","t3"]`)
	assert.Equal(t, fakeA.payloads[key], fakeB.payloads[key])
	assert.Contains(t, combined, "Tag(s) [t1, t2, t3] removed")
}

func TestStatusAssertFailsWhenASecondaryHostIsDown(t *testing.T) {
	serverA, _ := newRecordingKosliHost(t, map[string]string{"/ready": `"OK"`})
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()

	_, _, _, _, err := executeCommandC("status --assert " + multiHostFlags(serverA.URL, down.URL))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "["+down.URL+"]\nkosli server "+down.URL+" is unresponsive")
}

func TestEvaluateTrailAssertFailsWhenASecondaryHostDenies(t *testing.T) {
	trailPath := "/api/v2/trails/acme/my-flow/my-trail"
	serverA, _ := newRecordingKosliHost(t, map[string]string{trailPath: `{"compliance_status": {"attestations_statuses": [{"attestation_name": "bar", "attestation_id": null}]}}`})
	serverB, _ := newRecordingKosliHost(t, map[string]string{trailPath: `{"compliance_status": {"attestations_statuses": []}}`})

	policy := filepath.Join(t.TempDir(), "needs-bar.rego")
	require.NoError(t, os.WriteFile(policy, []byte(`package policy

default allow := false

allow if input.trail.compliance_status.attestations_statuses.bar

violations contains "trail attestation 'bar' not found" if not input.trail.compliance_status.attestations_statuses.bar
`), 0o600))

	_, combined, _, _, err := executeCommandC("evaluate trail my-trail --flow my-flow --policy " + policy + " " +
		multiHostFlags(serverA.URL, serverB.URL))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "["+serverB.URL+"]\npolicy denied")
	assert.Regexp(t, `RESULT:\s+ALLOWED`, combined)
}

func TestAllDryRunTokensMakeTheWholeRunADryRun(t *testing.T) {
	serverA, fakeA := newRecordingKosliHost(t, nil)
	serverB, fakeB := newRecordingKosliHost(t, nil)

	_, _, _, _, err := executeCommandStdinC(fmt.Sprintf(
		"delete service-account ci --org acme --host %s,%s --api-token DRY_RUN,DRY_RUN", serverA.URL, serverB.URL), "")

	require.NoError(t, err, "a dry run must not ask for confirmation")
	assert.Empty(t, fakeA.requests)
	assert.Empty(t, fakeB.requests)
}

func TestADeleteFailingPartwayOnASecondaryNamesWhatItHadDeleted(t *testing.T) {
	serverA, _ := newRecordingKosliHost(t, nil)
	serverB, fakeB := newRecordingKosliHost(t, nil)
	fakeB.statuses["DELETE /api/v2/service-accounts/acme/s2"] = http.StatusBadRequest

	_, combined, _, _, err := executeCommandC("delete service-account s1 s2 --assume-yes " +
		multiHostFlags(serverA.URL, serverB.URL))

	require.Error(t, err)
	assert.Contains(t, combined, "[warning] ["+serverB.URL+"] Service accounts already deleted before this failure")
}

func TestReportArtifactSendsEachHostItsOwnChangelog(t *testing.T) {
	gv, err := gitview.New("../..")
	require.NoError(t, err)
	head, err := gv.ResolveRevision("HEAD")
	require.NoError(t, err)
	oneBack, err := gv.ResolveRevision("HEAD~1")
	require.NoError(t, err)
	threeBack, err := gv.ResolveRevision("HEAD~3")
	require.NoError(t, err)
	latest := "/api/v2/artifacts/acme/my-flow/" + multiHostFingerprint + "/latest_commit"
	serverA, fakeA := newRecordingKosliHost(t, map[string]string{latest: `{"latest_commit": "` + oneBack + `"}`})
	serverB, fakeB := newRecordingKosliHost(t, map[string]string{latest: `{"latest_commit": "` + threeBack + `"}`})

	_, _, _, _, err = executeCommandC("report artifact app --fingerprint " + multiHostFingerprint +
		" --flow my-flow --git-commit " + head + " --repo-root ../.. --build-url http://b --commit-url http://c " +
		multiHostFlags(serverA.URL, serverB.URL))

	require.NoError(t, err)
	commits := func(fake *recordingKosliHost) int {
		var sent struct {
			CommitsList []any `json:"commits_list"`
		}
		require.NoError(t, json.Unmarshal([]byte(fake.payloads["POST /api/v2/artifacts/acme/my-flow"]), &sent))
		return len(sent.CommitsList)
	}
	assert.Less(t, commits(fakeA), commits(fakeB))
}
