package main

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/kosli-dev/cli/internal/kosli"
	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
)

// snapshotReporter reports Snapshots of Environments to Kosli.
type snapshotReporter struct {
	kosli  *kosli.Kosli
	org    string
	logger *log.Logger
	// ensure prepares the Environment on a host before its first report, e.g.
	// by creating it under --auto-environment.
	ensure func(h kosli.Host, envName, envType string) error
	// ensured holds the host, Environment name and type triples already
	// prepared, so watch mode and multi-environment runs prepare each one only
	// before its first report. A failed ensure is not recorded, so the next
	// report retries. mu guards it and is held across ensure on purpose:
	// snapshot paths --watch reports from one goroutine per path, and two
	// concurrent first reports must not both create the Environment.
	mu      sync.Mutex
	ensured map[string]bool
}

func newSnapshotReporter() *snapshotReporter {
	return &snapshotReporter{
		kosli:  kosliAPI(),
		org:    global.Org,
		logger: logger,
		ensure: ensureEnvironment,
	}
}

// snapshotCollector gathers a Snapshot: the request payload for the
// Environment type, and the line to log once Kosli has recorded it.
type snapshotCollector func() (payload any, reported string, err error)

// report collects the Snapshot once, when the first host has its Environment
// ready, and reports it to every such host.
func (r *snapshotReporter) report(envName, envType string, collect snapshotCollector) error {
	var collectErr error
	collectOnce := sync.OnceValues(func() (snapshot, error) {
		payload, reported, err := collect()
		collectErr = err
		return snapshot{payload, reported}, err
	})
	err := r.kosli.Send(func(h kosli.Host) error {
		if err := r.ensureOnce(h, envName, envType); err != nil {
			return err
		}
		collected, err := collectOnce()
		if err != nil {
			return err
		}
		if err := r.put(h, envName, envType, collected.payload); err != nil {
			return err
		}
		if !h.DryRun {
			h.Info("%s", collected.reported)
		}
		return nil
	})
	// A failed collection is the same failure on every host, so it is not
	// repeated per host.
	if collectErr != nil {
		return collectErr
	}
	return err
}

type snapshot struct {
	payload  any
	reported string
}

func (r *snapshotReporter) put(h kosli.Host, envName, envType string, payload any) error {
	endpoint, err := url.JoinPath(h.URL, "api/v2/environments", r.org, envName, "report", envType)
	if err != nil {
		return err
	}
	_, err = h.Do(&requests.RequestParams{
		Method:  http.MethodPut,
		URL:     endpoint,
		Payload: payload,
		DryRun:  h.DryRun,
		Token:   h.Token,
	})
	return err
}

func (r *snapshotReporter) ensureOnce(h kosli.Host, envName, envType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := h.URL + "/" + envName + "/" + envType
	if r.ensured[key] {
		return nil
	}
	if err := r.ensure(h, envName, envType); err != nil {
		return err
	}
	if r.ensured == nil {
		r.ensured = map[string]bool{}
	}
	r.ensured[key] = true
	return nil
}
