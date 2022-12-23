package main

import (
	"fmt"
	"os"
	"strings"

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
		// TODO: logger is most likely not initialized at this stage if the command does not exist or flags had problem, so can't be used
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
	}

	if err := cmd.Execute(); err != nil {
		// cobra does not capture unknown/missing commands, see https://github.com/spf13/cobra/issues/706
		// so we handle this here until it is fixed in cobra

		if strings.Contains(err.Error(), "unknown flag:") {
			c, flags, err := cmd.Traverse(os.Args[1:])
			if err != nil {
				c.PrintErrln("Error:", err.Error())
				// logger.Error(err.Error())
			}
			if c.HasSubCommands() {
				errMessage := ""
				if strings.HasPrefix(flags[0], "-") {
					errMessage = "Error: missing subcommand"
					// fmt.Fprintf(os.Stderr, "Error: missing subcommand\n")
				} else {
					errMessage = fmt.Sprintf("Error: unknown command: %s", flags[0])
					// fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", flags[0])
				}
				availableSubcommands := []string{}
				for _, sc := range c.Commands() {
					if !sc.Hidden {
						availableSubcommands = append(availableSubcommands, strings.Split(sc.Use, " ")[0])
					}
				}
				c.PrintErrf("%s\navailable subcommands are: %s\n", errMessage, strings.Join(availableSubcommands, " | "))
				// logger.Error("%s\navailable subcommands are: %s", errMessage, availableSubcommands)
				// fmt.Fprintf(os.Stderr, "available subcommands are: %s\n", strings.Join(availableSubcommands, " | "))
				os.Exit(1)
			}
		}

		cmd.PrintErrln("Error:", err.Error())
		// logger.Error(err.Error())
		if global.DryRun {
			logger.Warning("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
			os.Exit(0)
		}
		os.Exit(1)
	}
}
