package main

import (
	"fmt"
	"os"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var logger *log.Logger
var kosliClient *requests.Client

func main() {
	out := os.Stdout
	cmd, err := newRootCmd(out, os.Args[1:])
	if err != nil {
		logger.Error("%+v", err)
	}

	c, flags, err := cmd.Traverse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v --- %v", c.HasSubCommands(), flags)
	// commands := checkCommands(cmd)
	// for _, arg := range os.Args[1:] {
	// 	if strings.HasPrefix(arg, "-") {
	// 		break
	// 	}
	// 	if !strings.Contains(strings.Join(commands, ""), arg) {
	// 		fmt.Println("unknown command " + arg)
	// 		os.Exit(2)
	// 	}
	// }
	if err := cmd.Execute(); err != nil {
		if global.DryRun {
			logger.Warning("Encountered an error but --dry-run is enabled. Exiting with 0 exit code.")
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func checkCommands(cmd *cobra.Command) []string {
	results := []string{}
	for _, c := range cmd.Commands() {
		results = append(results, checkCommands(c)...)
	}
	results = append(results, cmd.Use)
	return results
}
