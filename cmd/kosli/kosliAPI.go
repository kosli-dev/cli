package main

import "github.com/kosli-dev/cli/internal/kosli"

// kosliAPI is every Kosli host the running command talks to.
func kosliAPI() *kosli.Kosli {
	return kosli.New(kosliClient, logger, global.Host, global.ApiToken, global.DryRun)
}
