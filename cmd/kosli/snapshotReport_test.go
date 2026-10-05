package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedSnapshotRequest struct {
	method string
	path   string
	body   map[string]any
}

type fakeSnapshotServer struct {
	recorded []recordedSnapshotRequest
	// status and body answer every request; zero status means 201.
	status int
	body   string
}

func newFakeSnapshotServer(t *testing.T) (*httptest.Server, *fakeSnapshotServer) {
	t.Helper()
	fake := &fakeSnapshotServer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The handler runs off the test goroutine, where require's FailNow
		// is unsafe, so failures are reported with assert and answered.
		raw, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			http.Error(w, "unreadable body", http.StatusInternalServerError)
			return
		}
		req := recordedSnapshotRequest{method: r.Method, path: r.URL.Path}
		if !assert.NoError(t, json.Unmarshal(raw, &req.body)) {
			http.Error(w, "body is not JSON", http.StatusInternalServerError)
			return
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

// ensureCalls records what the reporter asked to ensure, in order.
type ensureCalls struct {
	calls []string
	err   error
}

func (e *ensureCalls) ensure(envName, envType string) error {
	e.calls = append(e.calls, envName+"/"+envType)
	return e.err
}

func newTestSnapshotReporter(t *testing.T, host string, dryRun bool) (*snapshotReporter, *ensureCalls, *bytes.Buffer) {
	t.Helper()
	var out bytes.Buffer
	l := log.NewLogger(&out, &out, false)
	client, err := requests.NewKosliClient("", 0, false, l)
	require.NoError(t, err)
	ensured := &ensureCalls{}
	return &snapshotReporter{
		client: client,
		host:   host,
		org:    "acme",
		token:  "secret",
		dryRun: dryRun,
		logger: l,
		ensure: ensured.ensure,
	}, ensured, &out
}

func twoContainers() (any, string, error) {
	payload := map[string]any{"artifacts": []map[string]any{{"name": "a"}, {"name": "b"}}}
	return payload, "[2] containers were reported to environment prod", nil
}

func TestSnapshotReporterPutsThePayloadToTheEnvironmentTypeURL(t *testing.T) {
	server, fake := newFakeSnapshotServer(t)
	reporter, _, _ := newTestSnapshotReporter(t, server.URL, false)

	err := reporter.report("prod", "docker", twoContainers)

	require.NoError(t, err)
	require.Len(t, fake.recorded, 1)
	require.Equal(t, http.MethodPut, fake.recorded[0].method)
	require.Equal(t, "/api/v2/environments/acme/prod/report/docker", fake.recorded[0].path)
	require.Len(t, fake.recorded[0].body["artifacts"], 2)
}
