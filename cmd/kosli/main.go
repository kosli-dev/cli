package main

import (
	"os"

	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var log = logrus.New()

func main() {
	out := os.Stdout
	log.Out = out
	log.Formatter = &logrus.TextFormatter{
		DisableTimestamp: true,
	}
	cmd, err := newRootCmd(out, os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
	if err := cmd.Execute(); err != nil {
		if global.DryRun {
			log.Infof("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
			os.Exit(0)
		}
		os.Exit(1)
	}
}
