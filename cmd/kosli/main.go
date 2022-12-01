package main

import (
	"os"

	log "github.com/kosli-dev/cli/internal/logger"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var logger *log.Logger

// func debug(format string, v ...interface{}) {
// 	if global.Verbose {
// 		format = fmt.Sprintf("[debug] %s\n", format)
// 		log.Output(2, fmt.Sprintf(format, v...))
// 	}
// }

// func warning(format string, v ...interface{}) {
// 	format = fmt.Sprintf("WARNING: %s\n", format)
// 	fmt.Fprintf(os.Stderr, format, v...)
// }

func main() {
	out := os.Stdout
	// log.Out = out
	// log.Formatter = &logrus.TextFormatter{
	// 	DisableTimestamp: true,
	// }
	// log.SetFlags(log.Lshortfile)
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
