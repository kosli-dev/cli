package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/version"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	logger      *log.Logger
	kosliClient *requests.Client
)

func init() {
	logger = log.NewStandardLogger()
	// needed for some tests, actual CLI client is initialized in root.go
	kosliClient, _ = requests.NewKosliClient("", 3, false, logger)
}

func main() {
	var err error
	if isMultiHost() {
		var output string
		output, err = runMultiHost(os.Args)
		fmt.Print(output)
	} else {
		var cmd *cobra.Command
		cmd, err = newRootCmd(logger.Out, logger.ErrOut, os.Args[1:])
		if err == nil {
			err = innerMain(cmd, os.Args)
		}
	}
	if err != nil {
		logger.Error(err.Error())
	}
}

// enrichError prefixes err with the failing command's identity so users
// running scripts with several similar commands (e.g. two `kosli attest snyk`
// calls) can tell which invocation failed. It prepends the command path and,
// when present and non-empty, the --flow and --trail flag values. env-provided
// values (KOSLI_FLOW/KOSLI_TRAIL) are included too, since bindFlags sets them
// on the flag. Returns err unchanged when cmd or err is nil.
func enrichError(cmd *cobra.Command, err error) error {
	if cmd == nil || err == nil {
		return err
	}
	parts := []string{cmd.CommandPath()}
	for _, name := range []string{"flow", "trail"} {
		if f := cmd.Flags().Lookup(name); f != nil && f.Value.String() != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", name, f.Value.String()))
		}
	}
	return fmt.Errorf("[%s] %w", strings.Join(parts, " "), err)
}

func innerMain(cmd *cobra.Command, args []string) error {
	executedCmd, err := cmd.ExecuteC()
	if err == nil {
		// Cobra handles --version internally and bypasses all hooks, so we print
		// the update notice here after the fact.
		// initialize() also never runs, so global.Debug is not set — check
		// the flag and KOSLI_DEBUG env var directly.
		if cmd.Root().Flags().Changed("version") {
			debugEnabled := cmd.Root().Flags().Changed("debug")
			// match Viper internal bool env coercion
			if !debugEnabled {
				if v, err := strconv.ParseBool(os.Getenv("KOSLI_DEBUG")); err == nil {
					debugEnabled = v
				}
			}
			if !debugEnabled {
				notice, _ := version.CheckForUpdate(version.GetVersion())
				if notice != "" {
					_, _ = fmt.Fprint(logger.ErrOut, notice)
				}
			}
		}

		return nil
	}

	// cobra does not capture unknown/missing commands, see https://github.com/spf13/cobra/issues/706
	// so we handle this here until it is fixed in cobra
	if strings.Contains(err.Error(), "unknown flag:") {
		c, flags, err := cmd.Traverse(args[1:])
		if err != nil {
			return err
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
		logger.Info("Error: %s", enrichError(executedCmd, err).Error())
		logger.Warn("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
		return nil
	}
	return enrichError(executedCmd, err)
}
