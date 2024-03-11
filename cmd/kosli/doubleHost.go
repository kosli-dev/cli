package main

/*
Kosli uses CI pipelines in the cyber-dojo Org repos [*] for two purposes:
1. public facing documentation
2. private development purposes

All Kosli CLI calls in [*] are made to _two_ servers (because of 2)
  - https://app.kosli.com
  - https://staging.app.kosli.com

Explicitly making each Kosli CLI call in [*] twice is not an option (because of 1)
The least-worst option is to allow KOSLI_HOST and KOSLI_API_TOKEN to specify two values.
Note: cyber-dojo must ensure its api-tokens do not contain commas.
*/

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/spf13/cobra"
)

func isDoubleHost() bool {
	// Returns true iff the CLI execution is double-host, double-api-token
	return len(getHosts()) == 2 && len(getApiTokens()) == 2
}

func runDoubleHost(args []string) error {
	// Calls "innerMain" twice with the 0th call taking precedence over the 1st call.
	//  - Call first with the 0th host/api-token
	//  - Call next with the 1st host/api-token
	//
	// Always print the 0th call output.
	// The aim is to make it look like only the 0th call occurred
	//   - print the 1st call output only in debug mode
	// If the 1st call fails:
	// 	- print its error message, making its host clear
	// 	- return a non-zero exit-code, so errors are not silently ignored

	hosts := getHosts()
	apiTokens := getApiTokens()

	argsAppendHostApiTokenFlags := func(n int) []string {
		// Return args appended with the given host and api-token.
		// No need to strip existing --host/--api-token flags from args
		// as appended flags take precedence.
		hostFlag := fmt.Sprintf("--host=%s", hosts[n])
		apiTokenFlag := fmt.Sprintf("--api-token=%s", apiTokens[n])
		return append(args, hostFlag, apiTokenFlag)
	}

	args0 := argsAppendHostApiTokenFlags(0)
	output0, _, err0 := runBufferedInnerMain(args0)
	fmt.Print(output0)

	args1 := argsAppendHostApiTokenFlags(1)
	output1, global1, err1 := runBufferedInnerMain(args1)
	if global1.Debug {
		fmt.Print(output1)
	}

	var errorMessage string
	if err0 != nil {
		errorMessage += err0.Error()
	}
	if err1 != nil {
		errorMessage += fmt.Sprintf("\n%s\n\t%s", hosts[1], err1.Error())
	}

	if errorMessage == "" {
		return nil
	} else {
		return fmt.Errorf("%s", errorMessage)
	}
}

func runBufferedInnerMain(args []string) (string, *GlobalOpts, error) {
	// There is a logger.Error(..) call in main. It must be restored to use
	// the non-buffered global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(logger *log.Logger) { *globalLogger = logger }(logger)

	// Use a buffered Writer so output printing is decided by the caller.
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger = log.NewLogger(writer, writer, false)

	// newRootCmd(out, args) does _not_ use its args parameter.
	// Viper must be reading os.Args.
	// So we have to set os.Args here.
	defer func(args []string) { os.Args = args }(os.Args)
	os.Args = args

	// Ensure we reset global
	globalPtr := &global
	defer func(p *GlobalOpts) { *globalPtr = p }(global)

	// inner_main uses its argument for custom error messages
	err := innerMain(os.Args)
	return fmt.Sprint(&buffer), global, err
}

func getHosts() []string {
	g := func() string { return global.Host }
	return splitGlobal(g)
}

func getApiTokens() []string {
	g := func() string { return global.ApiToken }
	return splitGlobal(g)
}

func splitGlobal(g func() string) []string {
	// Execute() sets global, so ensure we reset it
	globalPtr := &global
	defer func(p *GlobalOpts) { *globalPtr = p }(global)
	// Ignore any error, we want only to set the global fields.
	_ = nullCmd().Execute()
	return strings.Split(g(), ",")
}

func nullCmd() *cobra.Command {
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	cmd, _ := newRootCmd(writer, nil)
	return cmd
}
