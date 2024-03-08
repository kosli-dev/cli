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
		if false {
			argsProd := append(args[1:], "--debug", "--host=https://app.kosli.com")
			argsStaging := append(args[1:], "--debug", "--host=https://staging.app.kosli.com")
			return argsProd, argsStaging
		} else {
			return nil, nil
		}
	}
	return nil, nil
}

func runBufferedInnerMain(args []string) (string, error) {
	// Use a buffered Writer because we want usually dont want
	// to print output for the cyber-dojo staging call
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger := log.NewLogger(writer, writer, false)
	err := inner_main(logger, args)
	return fmt.Sprint(&buffer), err
}

func main() {
	prodArgs, stagingArgs := ProdAndStagingCyberDojoCallArgs(os.Args)
	if prodArgs == nil && stagingArgs == nil {
		// Normal call
		err := inner_main(log.NewStandardLogger(), os.Args)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(42)
		}
	}

	// Kosli uses CI pipelines in the cyber-dojo Org repos for two purposes
	// 1. public facing documentation
	// 2. private development purposes, specifically
	//    all Kosli CLI calls are made twice, to two different servers
	//     - https://app.kosli.com
	//     - https://staging.app.kolsi.com
	//    We do not want to have to explicitly make each Kosli CLI call twice
	//    since that would not serve well for the documentation.
	// The least worst option is to allow multiple KOLSI_HOST and KOSLI_API_TOKEN
	// to specify more than one flag value.

	prodOutput, prodErr := runBufferedInnerMain(prodArgs)
	fmt.Print(prodOutput)
	if prodErr != nil {
		fmt.Printf("%s\n", prodErr.Error())
	}

	stagingOutput, stagingErr := runBufferedInnerMain(stagingArgs)
	if stagingErr != nil {
		// Only show staging output if there is an error
		fmt.Print(stagingOutput)
		fmt.Printf("%s\n", stagingErr.Error())
	}

	if prodErr == nil && stagingErr == nil {
		os.Exit(0)
	} else {
		os.Exit(42)
	}
}

func inner_main(log *log.Logger, args []string) error {
	cmd, err := newRootCmd(log.Out, args[1:])
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
			log.Error("%s\navailable subcommands are: %s", errMessage, strings.Join(availableSubcommands, " | "))
		}
	}
	if global.DryRun {
		log.Info("Error: %s", err.Error())
		log.Warning("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
		return nil
	}
	return err
}
