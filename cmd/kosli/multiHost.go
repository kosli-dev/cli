package main

/*
Kosli uses CI pipelines in the cyber-dojo Org repos [*] for two purposes:
1. public facing documentation
2. private development purposes

All Kosli CLI calls in [*] are made to _three_ servers (because of 2)
  - https://app.kosli.com
  - https://staging.app.kosli.com
  - https://app.us.kosli.com

Explicitly making each Kosli CLI call in [*] three times is not an option (because of 1)
Duplicating the entire CI workflows is complex, eg, deployments must not be duplicated.
The least-worst option is to allow KOSLI_HOST and KOSLI_API_TOKEN to specify multiple
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

// multiHost reports whether the CLI execution is multi-host, multi-api-token.
func (o MultiOpts) multiHost() bool {
	return hasMultipleHosts(o.hosts, o.apiTokens)
}

// runsPerHost reports whether main should run the command once per host with
// runMultiHost. attest artifact with a build command is the exception: it
// attests to every host itself, so its build runs once.
func (o MultiOpts) runsPerHost() bool {
	return o.multiHost() && !o.buildCommand
}

func hasMultipleHosts(hosts, apiTokens []string) bool {
	return len(hosts) > 1 && len(apiTokens) > 1 && len(hosts) == len(apiTokens)
}

func runMultiHost(args []string) (string, string, error) {
	// Calls "innerMain" at least twice:
	//  - with the 0th host/api-token (primary)
	//  - with the 1st, 2nd, 3rd, etc host/api-token (secondaries)

	opts := getMultiOpts()

	argsAppendHostApiTokenFlags := func(n int) []string {
		// Return args appended with the [n]th host/api-token.
		// No need to strip existing --host/--api-token flags from args
		// as appended flags take precedence.
		hostFlag := fmt.Sprintf("--host=%s", opts.hosts[n])
		apiTokenFlag := fmt.Sprintf("--api-token=%s", opts.apiTokens[n])
		return append(args, hostFlag, apiTokenFlag)
	}

	args0 := argsAppendHostApiTokenFlags(0)
	stdOut, stdErr, err0 := runBufferedInnerMain(args0)

	var errorMessage string

	if err0 != nil {
		errorMessage = fmt.Sprintf("[%s]\n", opts.hosts[0])
		errorMessage += err0.Error()
	}

	secondariesOut := ""
	secondariesErr := ""
	for i := 1; i < len(opts.hosts); i++ {
		argsN := argsAppendHostApiTokenFlags(i)
		outputN, stdErrN, errN := runBufferedInnerMain(argsN)

		// Return subsidiary-call's output in debug mode only.
		if opts.debug && outputN != "" {
			secondariesOut += fmt.Sprintf("\n[debug] [%s]", opts.hosts[i])
			secondariesOut += fmt.Sprintf("\n%s", outputN)
		}
		if opts.debug && stdErrN != "" {
			secondariesErr += fmt.Sprintf("\n[debug] [%s]", opts.hosts[i])
			secondariesErr += fmt.Sprintf("\n%s", stdErrN)
		}

		// Make origin of subsidiary-call failure clear.
		if errN != nil {
			errorMessage += fmt.Sprintf("\n[%s]", opts.hosts[i])
			errorMessage += fmt.Sprintf("\n%s", errN.Error())
		}
	}

	if secondariesOut != "" {
		// Ensure stdOut is prefixed with the primary hostname ONLY if we are
		// in debug mode - we could be producing JSON.
		stdOut = fmt.Sprintf("\n[debug] [%s]", opts.hosts[0]) + stdOut + "\n" + secondariesOut
	}
	if secondariesErr != "" {
		stdErr = fmt.Sprintf("\n[debug] [%s]\n", opts.hosts[0]) + stdErr + secondariesErr
	}

	var err error
	if errorMessage != "" {
		err = errors.New(errorMessage)
	}

	return stdOut, stdErr, err
}

func runBufferedInnerMain(args []string) (string, string, error) {
	// There is a logger.Error(..) call at the end of main. It must be restored to
	// the original global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(original *log.Logger) { *globalLogger = original }(logger)

	// Use buffered Writers so output printing is decided by the caller.
	// stdout and stderr are kept apart so a warning cannot corrupt, eg, JSON output.
	var outBuffer, errBuffer bytes.Buffer
	logger = log.NewLogger(&outBuffer, &errBuffer, false)

	// newRootCmd(out, args) does _not_ use its args parameter.
	// So we have to set os.Args here.
	defer func(original []string) { os.Args = original }(os.Args)
	os.Args = args

	// Create a cmd writing to the buffered Writers
	cmd, err := newRootCmd(logger.Out, logger.ErrOut, nil)
	if err != nil {
		return "", "", err
	}
	cmd.SetOut(&outBuffer)
	cmd.SetErr(&errBuffer)

	// Reset global back when done
	globalPtr := &global
	defer func(original *GlobalOpts) { *globalPtr = original }(global)

	// innerMain uses its argument for custom error messages
	err = innerMain(cmd, args)
	// SetOut also redirects cobra's deprecated-command notice, so move it back.
	executedCmd, _, _ := cmd.Find(normalizeBoolFlagArgs(cmd, args[1:]))
	stdOut, stdErr := moveDeprecationNoticeToStderr(executedCmd, fmt.Sprint(&outBuffer), fmt.Sprint(&errBuffer))
	return stdOut, stdErr, err
}

// moveDeprecationNoticeToStderr puts cobra's deprecated-command notice where a
// user sees it. Cobra writes it to OutOrStderr, which is stderr when SetOut has
// not been called (single-host main) but stdout once it has, as in
// runBufferedInnerMain and the test harness.
func moveDeprecationNoticeToStderr(c *cobra.Command, stdout, stderr string) (string, string) {
	if c == nil || c.Deprecated == "" {
		return stdout, stderr
	}
	notice := fmt.Sprintf("Command %q is deprecated, %s\n", c.Name(), c.Deprecated)
	if !strings.HasPrefix(stdout, notice) {
		return stdout, stderr
	}
	return strings.TrimPrefix(stdout, notice), notice + stderr
}

type MultiOpts struct {
	hosts     []string
	apiTokens []string
	debug     bool
	// buildCommand is set for attest artifact with a build command; see runsPerHost.
	buildCommand bool
}

func getMultiOpts() MultiOpts {
	// Return a MultiOpts struct with:
	//   - hosts set to H, split on comma, where H is the normal value of KOSLI_HOST/--host
	//   - apiTokens set to A, split on comma, where A is the normal value of KOSLI_API_TOKEN/--api-token
	//
	// For any error, return MultiOpts{} which will have
	//   - hosts == nil, so len(hosts) == 0
	//   - apiTokens == nil, so len(apiTokens) == 0
	// so multiHost() will return false.

	// There is a logger.Error(..) call at the end of main. Restore it to
	// the original global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(original *log.Logger) { *globalLogger = original }(logger)

	// Set the global logger to use a buffered Writer so any use of it produces no output.
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	logger = log.NewLogger(writer, writer, false)

	// Create a cmd object. Note: newRootCmd(out, args) does _not_ use its args parameter.
	cmd, err := newRootCmd(logger.Out, logger.ErrOut, nil)
	if err != nil {
		return MultiOpts{}
	}

	// Ensure cmd.Execute() prints nothing, even for a [kosli] call
	cmd.SetOut(writer)
	cmd.SetErr(writer)

	// Parse the same normalized args innerMain will run with, so a boolean flag
	// written in the space form (eg --debug false) does not leave a stray
	// positional argument that fails arg validation and empties MultiOpts.
	if len(os.Args) > 1 {
		cmd.SetArgs(normalizeBoolFlagArgs(cmd, os.Args[1:]))
	}

	fakeError := errors.New("")
	var buildCommand bool

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Call initialize() to bind cobra and viper
		err := initialize(cmd, writer, writer)
		if err != nil {
			return err
		}
		if cmd.CommandPath() == "kosli attest artifact" {
			_, buildCmd, _ := splitArtifactAndBuildArgs(cmd.ArgsLenAtDash(), args)
			buildCommand = len(buildCmd) > 0
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
		return MultiOpts{}
	}

	return MultiOpts{
		hosts:        strings.Split(global.Host, ","),
		apiTokens:    strings.Split(global.ApiToken, ","),
		debug:        global.Debug,
		buildCommand: buildCommand,
	}
}
