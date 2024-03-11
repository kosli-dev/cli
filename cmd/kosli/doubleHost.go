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
)

func isDoubleHost(args []string) bool {
	// Returns true iff the CLI execution is double-host, double-api-token
	opts := getDoubleOpts(args)
	return len(opts.hosts) == 2 && len(opts.apiTokens) == 2
}

func runDoubleHost(args []string) (string, error) {

	// Calls "innerMain" twice with the 0th call taking precedence over the 1st call.
	//  - Call first with the 0th host/api-token
	//  - Call next with the 1st host/api-token
	//
	// Always returns the 0th call output.
	// The aim is to make it look like only the 0th call occurred
	//   - return the 1st call output only in debug mode
	// If the 1st call fails:
	// 	- return its error message, making its host clear
	// 	- return a non-zero exit-code, so errors are not silently ignored

	opts := getDoubleOpts(args)

	argsAppendHostApiTokenFlags := func(n int) []string {
		// Return args appended with the given host and api-token.
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

	stdOut := output0
	if opts.debug {
		stdOut += fmt.Sprintf("[debug] %s\n", opts.hosts[1])
		stdOut += output1
	}

	var errorMessage string
	if err0 != nil {
		errorMessage += err0.Error()
	}
	if err1 != nil {
		errorMessage += fmt.Sprintf("\n%s\n\t%s", opts.hosts[1], err1.Error())
	}

	if errorMessage == "" {
		return stdOut, nil
	} else {
		return stdOut, fmt.Errorf("%s", errorMessage)
	}
}

func runBufferedInnerMain(args []string) (string, error) {
	// There is a logger.Error(..) call in main. It must be restored to use
	// the non-buffered global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(original *log.Logger) { *globalLogger = original }(logger)

	// Use a buffered Writer so output printing is decided by the caller.
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger = log.NewLogger(writer, writer, false)

	// newRootCmd(out, args) does _not_ use its args parameter.
	// Viper must be reading os.Args.
	// So we have to set os.Args here.
	defer func(original []string) { os.Args = original }(os.Args)
	os.Args = args

	// Reset global back when done
	globalPtr := &global
	defer func(original *GlobalOpts) { *globalPtr = original }(global)

	// inner_main uses its argument for custom error messages
	err := innerMain(os.Args)
	return fmt.Sprint(&buffer), err
}

type DoubleOpts struct {
	hosts     []string
	apiTokens []string
	debug     bool
}

func getDoubleOpts(args []string) DoubleOpts {
	// There is a logger.Error(..) call in main. It must be restored to use
	// the non-buffered global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(original *log.Logger) { *globalLogger = original }(logger)
	// Use a buffered Writer so output swallowed.
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger = log.NewLogger(writer, writer, false)

	// Append --dry-run; we don't want to execute, we just want to set global
	defer func(original []string) { os.Args = original }(os.Args)
	os.Args = append(args, "--dry-run")

	// Reset global back when done
	globalPtr := &global
	defer func(original *GlobalOpts) { *globalPtr = original }(global)

	// These three statements set the fields in global
	cmd, _ := newRootCmd(logger.Out, os.Args[1:])
	_ = cmd.Execute()
	initialize(cmd, writer)

	return DoubleOpts{
		hosts:     strings.Split(global.Host, ","),
		apiTokens: strings.Split(global.ApiToken, ","),
		debug:     global.Debug,
	}
}
