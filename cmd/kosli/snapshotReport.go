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
}

// snapshotCollector gathers a Snapshot: the request payload for the
// Environment type, and the line to log once Kosli has recorded it.
type snapshotCollector func() (payload any, reported string, err error)

func (r *snapshotReporter) report(envName, envType string, collect snapshotCollector) error {
	endpoint, err := url.JoinPath(r.host, "api/v2/environments", r.org, envName, "report", envType)
	if err != nil {
		return err
	}

	payload, _, err := collect()
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
	return err
}
