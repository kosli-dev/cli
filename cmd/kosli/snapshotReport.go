package main

import (
	"context"
	"sync"

	"github.com/kosli-dev/cli/internal/kosli"
	log "github.com/kosli-dev/cli/internal/logger"
)

// snapshotReporter reports Snapshots of Environments to Kosli.
type snapshotReporter struct {
	api *kosli.Client
	// dryRun only silences the success line: api sends nothing in a dry run.
	dryRun bool
	logger *log.Logger
	// ensure prepares the Environment before its first report, e.g. by
	// creating it under --auto-environment.
	ensure func(ctx context.Context, envName, envType string) error
	// ensured holds the Environment name and type pairs already prepared, so
	// watch mode and multi-environment runs prepare each one only before its
	// first report. A failed ensure is not recorded, so the next report retries.
	// mu guards it and is held across ensure on purpose: snapshot paths --watch
	// reports from one goroutine per path, and two concurrent first reports
	// must not both create the Environment.
	mu      sync.Mutex
	ensured map[string]bool
}

func newSnapshotReporter() *snapshotReporter {
	return &snapshotReporter{
		api:    kosliAPI(),
		dryRun: global.DryRun,
		logger: logger,
		ensure: ensureEnvironment,
	}
}

// snapshotCollector gathers a Snapshot: the request payload for the
// Environment type, and the line to log once Kosli has recorded it.
type snapshotCollector func() (payload any, reported string, err error)

func (r *snapshotReporter) report(ctx context.Context, envName, envType string, collect snapshotCollector) error {
	if err := r.ensureOnce(ctx, envName, envType); err != nil {
		return err
	}

	payload, reported, err := collect()
	if err != nil {
		return err
	}

	_, err = r.api.ReportSnapshot(ctx, envName, envType, payload)
	if err == nil && !r.dryRun {
		r.logger.Info("%s", reported)
	}
	return err
}

func (r *snapshotReporter) ensureOnce(ctx context.Context, envName, envType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := envName + "/" + envType
	if r.ensured[key] {
		return nil
	}
	if err := r.ensure(ctx, envName, envType); err != nil {
		return err
	}
	if r.ensured == nil {
		r.ensured = map[string]bool{}
	}
	r.ensured[key] = true
	return nil
}
