package main

import (
	"bytes"
	"fmt"
	"io"
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

func prodAndStagingCyberDojoCallArgs(args []string) ([]string, []string) {
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	nullLogger := log.NewLogger(writer, writer, false)

	_, err := newRootCmd(nullLogger.Out, args[1:])
	if err == nil {
		// host := cmd.Flag("host")
		// if host.Value.String() == "https://app.kosli.com,https://staging.app.kosli.com" {
		// 	logger.Info("Host is Doubled")
		// }
		// apiToken := cmd.Flag("api-token")
		// org := cmd.Flag("org")

		// TODO: proper check for doubled host etc
		if true {
			argsProd := append(args[1:], "--host=https://app.kosli.com")            // --api-token=...
			argsStaging := append(args[1:], "--host=https://staging.app.kosli.com") // --api-token=...
			return argsProd, argsStaging
		} else {
			return nil, nil
		}
	}
	return nil, nil
}

func runProdAndStagingCyberDojoCalls(prodArgs []string, stagingArgs []string) error {
	// Kosli uses CI pipelines in the cyber-dojo Org repos [*] for two purposes:
	// 1. public facing documentation
	// 2. private development purposes
	//
	// All Kosli CLI calls in [*] are made to _two_ servers
	//   - https://app.kosli.com
	//   - https://staging.app.kolsi.com  (because of 2)
	//
	// Explicitly making each Kosli CLI call in [*] twice is not an option because of 1)
	// The least-worst option is to allow KOLSI_HOST and KOSLI_API_TOKEN to specify two values.

	// Make double-host calls look as-if only the prod call occurred:
	//   - never print the staging-call output.
	//   - if the staging-call fails:
	//     - print its error message, making it clear it is from staging
	//     - return a non-zero exit-code, so staging errors are not silently ignored

	prodOutput, prodErr := runBufferedInnerMain(prodArgs)
	fmt.Print(prodOutput)
	_, stagingErr := runBufferedInnerMain(stagingArgs)

	var errorMessage string
	if prodErr != nil {
		errorMessage += prodErr.Error()
	}
	if stagingErr != nil {
		errorMessage += fmt.Sprintf("\n%s\n\t%s", "https://staging.app.kosli.com", stagingErr.Error())
	}

	if errorMessage == "" {
		return nil
	} else {
		return fmt.Errorf("%s", errorMessage)
	}
}

func runBufferedInnerMain(args []string) (string, error) {
	// When errors are logged in main() the non-buffered global logger
	// must be restored so the error messages actually appear.
	globalLogger := &logger
	defer func(logger *log.Logger) { *globalLogger = logger }(logger)
	// Use a buffered Writer so the output of the staging call is NOT printed
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger = log.NewLogger(writer, writer, false)
	err := inner_main(args)
	return fmt.Sprint(&buffer), err
}

func main() {
	var err error
	prodArgs, stagingArgs := prodAndStagingCyberDojoCallArgs(os.Args)
	if prodArgs == nil && stagingArgs == nil {
		err = inner_main(os.Args)
	} else {
		err = runProdAndStagingCyberDojoCalls(prodArgs, stagingArgs)
	}
	if err != nil {
		logger.Error(err.Error())
	}
}

func inner_main(args []string) error {
	cmd, err := newRootCmd(logger.Out, args[1:])
	if err != nil {
		return err
	}

	err = cmd.Execute()
	if err == nil {
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
		logger.Info("Error: %s", err.Error())
		logger.Warning("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
		return nil
	}
	return err
}
