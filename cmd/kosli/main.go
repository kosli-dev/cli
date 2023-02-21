package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	logger      *log.Logger
	kosliClient *requests.Client
)

func init() {
	logger = log.NewStandardLogger()
	kosliClient = requests.NewStandardKosliClient()
}

func main() {
	cmd, err := newRootCmd(logger.Out, os.Args[1:])
	if err != nil {
		logger.Error(err.Error())
	}

	if err := cmd.Execute(); err != nil {
		// cobra does not capture unknown/missing commands, see https://github.com/spf13/cobra/issues/706
		// so we handle this here until it is fixed in cobra
		if strings.Contains(err.Error(), "unknown flag:") {
			c, flags, err := cmd.Traverse(os.Args[1:])
			if err != nil {
				logger.Error(err.Error())
			}
			if c.HasSubCommands() {
				errMessage := ""
				if strings.HasPrefix(flags[0], "-") {
					errMessage = "missing subcommand"
				} else {
					errMessage = fmt.Sprintf("unknown command: %s", flags[0])
				}
				availableSubcommands := []string{}
				for _, sc := range c.Commands() {
					if !sc.Hidden {
						availableSubcommands = append(availableSubcommands, strings.Split(sc.Use, " ")[0])
					}
				}
				logger.Error("%s\navailable subcommands are: %s", errMessage, strings.Join(availableSubcommands, " | "))
			}
		}

		if global.DryRun {
			logger.Info("Error: %s", err.Error())
			logger.Warning("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
			os.Exit(0)
		}
		logger.Error(err.Error())
	}
}
