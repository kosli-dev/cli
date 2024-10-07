package main

/*
Kosli uses CI pipelines in the cyber-dojo Org repos [*] for two purposes:
1. public facing documentation
2. private development purposes

All Kosli CLI calls in [*] are made to _two_ servers (because of 2)
  - https://app.kosli.com
  - https://staging.app.kosli.com

Explicitly making each Kosli CLI call in [*] twice is not an option (because of 1)
Duplicating the entire CI workflows is complex because, eg, deployments must not be duplicated.
The least-worst option is to allow KOSLI_HOST and KOSLI_API_TOKEN to specify two
comma-separated values. Note cyber-dojo must ensure its api-tokens do not contain commas.
*/

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/spf13/cobra"
)

func isDoubledHost() bool {
	// Returns true iff the CLI execution is doubled-host, doubled-api-token
	opts := getDoubledOpts()
	return len(opts.hosts) == 2 && len(opts.apiTokens) == 2
}

func runDoubledHost(args []string) (string, error) {
	// Calls "innerMain" twice:
	//  - with the 0th host/api-token (primary)
	//  - with the 1st host/api-token (subsidiary)

	opts := getDoubledOpts()

	argsAppendHostApiTokenFlags := func(n int) []string {
		// Return args appended with the [n]th host/api-token.
		// No need to strip existing --host/--api-token flags from args
		// as appended flags take precedence.
		hostFlag := fmt.Sprintf("--host=%s", opts.hosts[n])
		apiTokenFlag := fmt.Sprintf("--api-token=%s", opts.apiTokens[n])
		return append(args, hostFlag, apiTokenFlag)
	}

	args0 := argsAppendHostApiTokenFlags(0)
	output0, err0 := runBufferedInnerMain(args0)

	args1 := argsAppendHostApiTokenFlags(1)
	output1, err1 := runBufferedInnerMain(args1)

	// Return subsidiary-call's output in debug mode only.
	stdOut := output0
	if opts.debug == "true" && output1 != "" {
		stdOut += fmt.Sprintf("\n[debug] [%s]", opts.hosts[1])
		stdOut += fmt.Sprintf("\n%s", output1)
	}

	// Make origin of subsidiary-call failure clear.
	var errorMessage string
	if err0 != nil {
		errorMessage += err0.Error()
	}
	if err1 != nil {
		errorMessage += fmt.Sprintf("\n[%s]", opts.hosts[1])
		errorMessage += fmt.Sprintf("\n%s", err1.Error())
	}

	var err error
	if errorMessage != "" {
		err = errors.New(errorMessage)
	}

	return stdOut, err
}

func runBufferedInnerMain(args []string) (string, error) {
	// There is a logger.Error(..) call at the end of main. It must be restored to
	// the original global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(original *log.Logger) { *globalLogger = original }(logger)

	// Use a buffered Writer so output printing is decided by the caller.
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger = log.NewLogger(writer, writer, false)

	// newRootCmd(out, args) does _not_ use its args parameter.
	// So we have to set os.Args here.
	defer func(original []string) { os.Args = original }(os.Args)
	os.Args = args

	// Create a cmd writing to the buffered Writer
	cmd, err := newRootCmd(logger.Out, nil)
	if err != nil {
		return "", err
	}
	cmd.SetOut(writer)
	cmd.SetErr(writer)

	// Reset global back when done
	globalPtr := &global
	defer func(original *GlobalOpts) { *globalPtr = original }(global)

	// innerMain uses its argument for custom error messages
	err = innerMain(cmd, args)
	return fmt.Sprint(&buffer), err
}

type DoubledOpts struct {
	hosts     []string
	apiTokens []string
	debug     string
}

func getDoubledOpts() DoubledOpts {
	// Return a DoubledOpts struct with:
	//   - hosts set to H, split on comma, where H is the normal value of KOSLI_HOST/--host
	//   - apiTokens set to A, split on comma, where A is the normal value of KOSLI_API_TOKEN/--api-token
	//
	// For any error, return DoubleOpts{} which will have
	//   - hosts == nil, so len(hosts) == 0
	//   - apiTokens == nil, so len(apiTokens) == 0
	// so isDoubledHost() will return false.

	// There is a logger.Error(..) call at the end of main. Restore it to
	// the original global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(original *log.Logger) { *globalLogger = original }(logger)

	// Set the global logger to use a buffered Writer so any use of it produces no output.
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger = log.NewLogger(writer, writer, false)

	// Create a cmd object. Note: newRootCmd(out, args) does _not_ use its args parameter.
	cmd, err := newRootCmd(logger.Out, nil)
	if err != nil {
		return DoubledOpts{}
	}

	// Ensure cmd.Execute() prints nothing, even for a [kosli] call
	cmd.SetOut(writer)
	cmd.SetErr(writer)

	fakeError := errors.New("")

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Call initialize() to bind cobra and viper
		err := initialize(cmd, writer)
		if err != nil {
			return err
		}
		return fakeError
	}

	// We are setting global's fields. Reset global back when done.
	globalPtr := &global
	defer func(original *GlobalOpts) { *globalPtr = original }(global)

	// Call cmd.Execute() to set global's fields.
	err = cmd.Execute()
	if err != nil && err != fakeError {
		// Genuine error
		// Eg kosli unknownCommand ...
		// Eg kosli status --unknown-flag
		return DoubledOpts{}
	}

	return DoubledOpts{
		hosts:     strings.Split(global.Host, ","),
		apiTokens: strings.Split(global.ApiToken, ","),
		debug:     global.Debug,
	}
}
