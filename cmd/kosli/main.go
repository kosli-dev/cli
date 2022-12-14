package main

import (
	"os"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var logger *log.Logger
var kosliClient *requests.Client

func main() {
	out := os.Stdout
	cmd, err := newRootCmd(out, os.Args[1:])
	if err != nil {
		logger.Error("%+v", err)
	}
	if err := cmd.Execute(); err != nil {
		if global.DryRun {
			logger.Warning("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
			os.Exit(0)
		}
		os.Exit(1)
	}
}
