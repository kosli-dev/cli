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
	logger      *log.Logger // DROP?
	kosliClient *requests.Client
)

func init() {
	logger = log.NewStandardLogger()
	kosliClient = requests.NewStandardKosliClient()
}

func ProdAndStagingCyberDojoCallArgs(args []string) ([]string, []string) {
	// TODO: new logger here?
	_, err := newRootCmd(logger.Out, args[1:])
	if err == nil {
		/*host := cmd.Flag("host")
		if host.Value.String() == "https://app.kosli.com,https://staging.app.kosli.com" {
			logger.Info("Host is Doubled")
		}
		fmt.Printf("host.Value %+v", host.Value)
		apiToken := cmd.Flag("api-token")
		org := cmd.Flag("org")
		logger.Info("Host: %v", host)
		logger.Info("ApiToken: %v", apiToken)
		logger.Info("Org: %v", org)*/

		// TODO: proper check for doubled host etc
		if true {
			argsProd := append(args[1:], "--dry-run", "--debug", "--host=https://app.kosli.com")
			argsStaging := append(args[1:], "--dry-run", "--debug", "--host=https://staging.app.kosli.com")
			return argsProd, argsStaging
		} else {
			return nil, nil
		}
	}
	return nil, nil
}

func bufferedLogger() (*bytes.Buffer, *log.Logger) {
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	return &buffer, log.NewLogger(writer, writer, false)
}

func main() {
	prodArgs, stagingArgs := ProdAndStagingCyberDojoCallArgs(os.Args)
	if prodArgs != nil && stagingArgs != nil {
		fmt.Printf("Running inner_main() twice\n")

		prodBuffer, prodLogger := bufferedLogger()
		inner_main(prodLogger, prodArgs)
		fmt.Print(prodBuffer)

		stagingBuffer, stagingLogger := bufferedLogger()
		inner_main(stagingLogger, stagingArgs)
		fmt.Print(stagingBuffer)

	} else {
		fmt.Printf("Running inner_main() once\n")
		inner_main(log.NewStandardLogger(), os.Args)
	}
}

func inner_main(log *log.Logger, args []string) {
	// TODO: make this accept logger.Out used for newRootCmd() call
	// TODO: pass in buffered-logger for doubled-calls, logger.Out for normal call
	// TODO: make this return (output, error)
	fmt.Printf("Inside inner_main: %v\n", args)

	//cmd, err := newRootCmd(logger.Out, args[1:])
	cmd, err := newRootCmd(log.Out, args[1:])

	if err != nil {
		log.Error(err.Error())
	}
	if err := cmd.Execute(); err != nil {
		// cobra does not capture unknown/missing commands, see https://github.com/spf13/cobra/issues/706
		// so we handle this here until it is fixed in cobra
		if strings.Contains(err.Error(), "unknown flag:") {
			c, flags, err := cmd.Traverse(args[1:])
			if err != nil {
				log.Error(err.Error())
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
				log.Error("%s\navailable subcommands are: %s", errMessage, strings.Join(availableSubcommands, " | "))
			}
		}

		if global.DryRun {
			log.Info("Error: %s", err.Error())
			log.Warning("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
			os.Exit(0)
		}
		log.Error(err.Error())
	}
}
