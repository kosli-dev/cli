package main

import (
	"fmt"
	"io"
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

// findUnknownCommand detects the unknown-command case that cobra's
// TraverseChildren routing swallows (issue #1043). It resolves args exactly as
// cobra will — on a throwaway command tree, so the real command's flag state is
// left untouched — and returns an error when resolution lands on a command that
// only groups subcommands (is not runnable) while an unconsumed non-flag token
// remains. It returns nil, deferring to cobra/ExecuteC, when args reach a
// runnable command, when nothing non-flag is left over (e.g. `kosli list`
// printing its own group help), or when Traverse itself errors (an unknown
// flag, --help, etc. — those paths are already handled downstream).
func findUnknownCommand(args []string) error {
	probe, err := newRootCmd(io.Discard, io.Discard, args)
	if err != nil {
		return nil
	}
	c, leftover, err := probe.Traverse(args)
	if err != nil || c.Runnable() {
		return nil
	}
	token := ""
	for _, a := range leftover {
		if !strings.HasPrefix(a, "-") {
			token = a
			break
		}
	}
	if token == "" {
		return nil
	}
	// If the leftover token names a real subcommand, traversal stopped for some
	// other reason (e.g. an unparseable flag ahead of it), not because the
	// command is unknown. Defer to cobra/ExecuteC so the existing flag
	// diagnostics still fire.
	for _, sc := range c.Commands() {
		if sc.Name() == token {
			return nil
		}
		for _, alias := range sc.Aliases {
			if alias == token {
				return nil
			}
		}
	}
	availableSubcommands := []string{}
	for _, sc := range c.Commands() {
		if !sc.Hidden {
			availableSubcommands = append(availableSubcommands, strings.Split(sc.Use, " ")[0])
		}
	}
	return fmt.Errorf("unknown command: %s\navailable subcommands are: %s",
		token, strings.Join(availableSubcommands, " | "))
}

// innerMain runs cmd against args (args[0] being the program name) and turns
// the outcome into the process-level error, printing the update notice on the
// --version path and reporting errors in the friendliest available form.
func innerMain(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		cmd.SetArgs(normalizeBoolFlagArgs(cmd, args[1:]))
		// cobra's TraverseChildren routing hands an unrecognized command token to
		// the nearest group command with no error; that command then prints help
		// and exits 0, so a typo'd command silently reports success (issue #1043).
		// Catch it before executing so the process exits non-zero with a
		// diagnostic instead.
		if err := findUnknownCommand(args[1:]); err != nil {
			return err
		}
	}
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
			// logger.Error terminates the process (os.Exit), so the line below is
			// not reached today. We return explicitly to keep this branch
			// self-contained: the unknown-subcommand case is already reported in a
			// friendly form, so it must not also fall through to the generic
			// enriched-error print below.
			return nil
		}
	}
	if global.DryRun {
		logger.Info("Error: %s", enrichError(executedCmd, err).Error())
		logger.Warn("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
		return nil
	}
	return enrichError(executedCmd, err)
}
