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

const prodHostURL = "https://app.kosli.com"
const stagingHostURL = "https://staging.app.kosli.com"

func CyberDojoProdAndStagingCallArgs(args []string) ([]string, []string) {
	// If the args call is a double-host, double-api-token cyber-dojo call then
	// return two []string args, modified so a call with those args targets two hosts:
	//  - https://app.kosli.com
	//  - https://staging.app.kosli.com
	// Otherwise return nil, nil to indicate this is not a doubled-call.

	orgs := splitGlobal(args, getOrg)
	hosts := splitGlobal(args, getHost)
	apiTokens := splitGlobal(args, getApiToken)

	isCyberDojo := len(orgs) == 1 && orgs[0] == "cyber-dojo"
	isDoubledHost := len(hosts) == 2 && hosts[0] == prodHostURL && hosts[1] == stagingHostURL
	isDoubledApiToken := len(apiTokens) == 2

	if isCyberDojo && isDoubledHost && isDoubledApiToken {

		argsAppendHostApiToken := func(n int) []string {
			// No need to strip existing --host/--api-token flags from args
			// as we are appending new flag values which take precedence.
			hostProd := fmt.Sprintf("--host=%s", hosts[n])
			apiTokenProd := fmt.Sprintf("--api-token=%s", apiTokens[n])
			return append(args, hostProd, apiTokenProd)
		}

		argsProd := argsAppendHostApiToken(0)
		argsStaging := argsAppendHostApiToken(1)
		// fmt.Printf("argsProd == %s\n", strings.Join(argsProd, " "))
		// fmt.Printf("argsStaging == %s\n", strings.Join(argsStaging, " "))
		return argsProd, argsStaging
	} else {
		return nil, nil
	}
}

func CyberDojoRunProdAndStagingCalls(prodArgs []string, stagingArgs []string) error {
	// Calls "inner_main" twice:
	//  - with prodArgs targetting https://app.kosli.com
	//  - with stagingArgs targetting https://staging.app.kosli.com
	// If the prod-call and the staging-call succeed:
	// 	- do NOT print the staging-call output, so it looks as-if only the prod call occurred.
	// If the staging-call fails:
	// 	- print its error message, making it clear it is from staging
	// 	- return a non-zero exit-code, so staging errors are not silently ignored

	prodOutput, prodErr := runBufferedInnerMain(prodArgs)
	fmt.Print(prodOutput)
	_, stagingErr := runBufferedInnerMain(stagingArgs)

	var errorMessage string
	if prodErr != nil {
		errorMessage += prodErr.Error()
	}
	if stagingErr != nil {
		errorMessage += fmt.Sprintf("\n%s\n\t%s", stagingHostURL, stagingErr.Error())
	}

	if errorMessage == "" {
		return nil
	} else {
		return fmt.Errorf("%s", errorMessage)
	}
}

func runBufferedInnerMain(args []string) (string, error) {
	// There is a logger.Error(..) call in main. It must be restored to use
	// the non-buffered global logger so the error messages actually appear.
	globalLogger := &logger
	defer func(logger *log.Logger) { *globalLogger = logger }(logger)
	// Use a buffered Writer so output printing is decided by the caller.
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	// Set global logger
	logger = log.NewLogger(writer, writer, false)
	// We have to reset os.Args here.
	// Presumably because viper is reading os.Args.
	// Note that newRootCmd(args) does not use its args parameter.
	os.Args = args
	// Ensure prod/staging calls do not interfere with each other.
	resetGlobal()
	// Finally!
	err := inner_main(args)
	return fmt.Sprint(&buffer), err
}

type getter func() string

func splitGlobal(args []string, g getter) []string {
	defer resetGlobal()
	// Ignore any error, we want only to set the global fields.
	_ = nullCmd(args).Execute()
	// Note: cyber-dojo must ensure its api-tokens do not contain commas.
	return strings.Split(g(), ",")
}

func resetGlobal() {
	global = new(GlobalOpts)
}

func nullCmd(args []string) *cobra.Command {
	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	cmd, _ := newRootCmd(writer, args)
	return cmd
}

func getOrg() string {
	return global.Org
}

func getHost() string {
	return global.Host
}

func getApiToken() string {
	return global.ApiToken
}
