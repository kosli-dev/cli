package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
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
	mu       sync.Mutex
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
		fake.mu.Lock()
		fake.recorded = append(fake.recorded, req)
		fake.mu.Unlock()
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

func TestSnapshotReporterEnsuresTheEnvironmentBeforeCollecting(t *testing.T) {
	server, _ := newFakeSnapshotServer(t)
	reporter, ensured, _ := newTestSnapshotReporter(t, server.URL, false)
	var ensuredWhenCollecting []string

	err := reporter.report("prod", "docker", func() (any, string, error) {
		ensuredWhenCollecting = append([]string{}, ensured.calls...)
		return twoContainers()
	})

	require.NoError(t, err)
	require.Equal(t, []string{"prod/docker"}, ensuredWhenCollecting)
}

func TestSnapshotReporterStopsWhenTheEnvironmentCannotBeEnsured(t *testing.T) {
	server, fake := newFakeSnapshotServer(t)
	reporter, ensured, _ := newTestSnapshotReporter(t, server.URL, false)
	ensured.err = errors.New("environment prod already exists with type K8S")
	collected := false

	err := reporter.report("prod", "docker", func() (any, string, error) {
		collected = true
		return twoContainers()
	})

	require.EqualError(t, err, "environment prod already exists with type K8S")
	require.False(t, collected)
	require.Empty(t, fake.recorded)
}

func TestSnapshotReporterSendsNothingWhenCollectingFails(t *testing.T) {
	server, fake := newFakeSnapshotServer(t)
	reporter, _, out := newTestSnapshotReporter(t, server.URL, false)

	err := reporter.report("prod", "docker", func() (any, string, error) {
		return nil, "", errors.New("cannot reach docker daemon")
	})

	require.EqualError(t, err, "cannot reach docker daemon")
	require.Empty(t, fake.recorded)
	require.Empty(t, out.String())
}

func TestSnapshotReporterLogsTheReportedLine(t *testing.T) {
	server, _ := newFakeSnapshotServer(t)
	reporter, _, out := newTestSnapshotReporter(t, server.URL, false)

	err := reporter.report("prod", "docker", twoContainers)

	require.NoError(t, err)
	require.Equal(t, "[2] containers were reported to environment prod\n", out.String())
}

func TestSnapshotReporterSendsNothingAndLogsNoSuccessInDryRun(t *testing.T) {
	server, fake := newFakeSnapshotServer(t)
	reporter, _, out := newTestSnapshotReporter(t, server.URL, true)

	err := reporter.report("prod", "docker", twoContainers)

	require.NoError(t, err)
	require.Empty(t, fake.recorded)
	require.Contains(t, out.String(), "THIS IS A DRY-RUN")
	require.NotContains(t, out.String(), "were reported to environment")
}

func TestSnapshotReporterReturnsTheServerErrorAndLogsNoSuccess(t *testing.T) {
	server, fake := newFakeSnapshotServer(t)
	fake.status = http.StatusBadRequest
	fake.body = `{"message":"Environment named 'prod' does not exist"}`
	reporter, _, out := newTestSnapshotReporter(t, server.URL, false)

	err := reporter.report("prod", "docker", twoContainers)

	require.EqualError(t, err, "Environment named 'prod' does not exist")
	require.Empty(t, out.String())
}

func TestSnapshotReporterEnsuresEachEnvironmentOnce(t *testing.T) {
	server, fake := newFakeSnapshotServer(t)
	reporter, ensured, _ := newTestSnapshotReporter(t, server.URL, false)

	require.NoError(t, reporter.report("prod", "server", twoContainers))
	require.NoError(t, reporter.report("prod", "server", twoContainers))
	require.NoError(t, reporter.report("staging", "server", twoContainers))

	require.Equal(t, []string{"prod/server", "staging/server"}, ensured.calls)
	require.Len(t, fake.recorded, 3)
}

// snapshot paths --watch reports from one goroutine per watched path through
// a shared reporter.
func TestSnapshotReporterIsSafeForConcurrentReports(t *testing.T) {
	server, _ := newFakeSnapshotServer(t)
	reporter, ensured, _ := newTestSnapshotReporter(t, server.URL, false)
	var mu sync.Mutex
	reporter.ensure = func(envName, envType string) error {
		mu.Lock()
		defer mu.Unlock()
		return ensured.ensure(envName, envType)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			assert.NoError(t, reporter.report("prod", "server", twoContainers))
		})
	}
	wg.Wait()

	require.Equal(t, []string{"prod/server"}, ensured.calls)
}
