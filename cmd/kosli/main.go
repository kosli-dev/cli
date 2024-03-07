package main

import (
	"fmt"
	"os"
	"os/exec"
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

func DoubleHostCyberDojoCallArgs() bool {
	cmd, err := newRootCmd(logger.Out, os.Args[1:])
	if err == nil {
		host := cmd.Flag("host")
		apiToken := cmd.Flag("api-token")
		org := cmd.Flag("org")
		logger.Info("Host: %v", host)
		logger.Info("ApiToken: %v", apiToken)
		logger.Info("Org: %v", org)
		/*if host.Value.String() == "https://app.kosli.com,https://staging.app.kosli.com" {
			logger.Info("Host is Doubled")
		}*/
		return true
	}
	return false
}

func DoubleHostCyberDojoCalls() int {

	// cmd, err := newRootCmd(logger.Out, os.Args[1:])
	// TODO: Get real host/api-token values

	// Hard-wire the two host/api-token values
	// Run exec.command() twice
	//     add --host=... and --api-token=...
	//     add --debug --dry-run
	//     capture output and err and handle as per Issue

	prodArgs := append(os.Args[1:], "--dry-run", "--debug")
	logger.Info("prodCmd: %s", os.Args[0])
	logger.Info("prodArgs: %v", prodArgs)
	prodCmd := exec.Command(os.Args[0], prodArgs...)
	prodOutput, prodError := prodCmd.Output()
	// TODO: get to here by preventing infinite loop!
	logger.Info("prodOutput %s\n", prodOutput)

	stagingArgs := append(os.Args[1:], "--dry-run", "--debug")
	logger.Info("stagingCmd: %s", os.Args[0])
	logger.Info("stagingArgs: %v", stagingArgs)
	stagingCmd := exec.Command(os.Args[0], stagingArgs...)
	stagingOutput, stagingError := stagingCmd.Output()
	if stagingError != nil {
		fmt.Printf("%s", stagingOutput)
	}

	status := 0
	if prodError != nil && stagingError != nil {
		status = 42
	}
	return status
}

func main() {
	if DoubleHostCyberDojoCallArgs() {
		status := DoubleHostCyberDojoCalls()
		os.Exit(status)
	}
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
