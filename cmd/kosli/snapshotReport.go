package main

import (
	"net/http"
	"net/url"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
)

// snapshotReporter reports Snapshots of Environments to Kosli.
type snapshotReporter struct {
	client *requests.Client
	host   string
	org    string
	token  string
	dryRun bool
	logger *log.Logger
	// ensure prepares the Environment before its first report, e.g. by
	// creating it under --auto-environment.
	ensure func(envName, envType string) error
	// ensured holds the Environments already prepared, so watch mode and
	// multi-environment runs prepare each one only before its first report.
	ensured map[string]bool
}

func newSnapshotReporter() *snapshotReporter {
	return &snapshotReporter{
		client: kosliClient,
		host:   global.Host,
		org:    global.Org,
		token:  global.ApiToken,
		dryRun: global.DryRun,
		logger: logger,
		ensure: ensureEnvironment,
	}
}

// snapshotCollector gathers a Snapshot: the request payload for the
// Environment type, and the line to log once Kosli has recorded it.
type snapshotCollector func() (payload any, reported string, err error)

func (r *snapshotReporter) report(envName, envType string, collect snapshotCollector) error {
	if !r.ensured[envName] {
		if err := r.ensure(envName, envType); err != nil {
			return err
		}
		if r.ensured == nil {
			r.ensured = map[string]bool{}
		}
		r.ensured[envName] = true
	}

	endpoint, err := url.JoinPath(r.host, "api/v2/environments", r.org, envName, "report", envType)
	if err != nil {
		return err
	}

	payload, reported, err := collect()
	if err != nil {
		return err
	}

	_, err = r.client.Do(&requests.RequestParams{
		Method:  http.MethodPut,
		URL:     endpoint,
		Payload: payload,
		DryRun:  r.dryRun,
		Token:   r.token,
	})
	if err == nil && !r.dryRun {
		r.logger.Info("%s", reported)
	}
	return err
}
