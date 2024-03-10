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

const prodHostURL = "https://app.kosli.com"
const stagingHostURL = "https://staging.app.kosli.com"

func IsCyberDojoDoubleHost() bool {
	// Returns true iff the CLI execution is a double-host, double-api-token, org==cyber-dojo
	orgs := getOrgs()
	hosts := getHosts()
	apiTokens := getApiTokens()

	isCyberDojo := len(orgs) == 1 && orgs[0] == "cyber-dojo"
	isDoubledHost := len(hosts) == 2 && hosts[0] == prodHostURL && hosts[1] == stagingHostURL
	isDoubledApiToken := len(apiTokens) == 2

	return isCyberDojo && isDoubledHost && isDoubledApiToken
}

func RunCyberDojoDoubleHost() error {
	// Calls "inner_main" twice:
	//  - with os.Args targetting https://app.kosli.com (prod)
	//  - with os.Args targetting https://staging.app.kosli.com (staging)
	// Always print the prod-call output
	// Print the staging-call output only in debug mode, so it looks as-if only the prod call occurred.
	// If the staging-call fails:
	// 	- print its error message, making it clear it is from staging
	// 	- return a non-zero exit-code, so staging errors are not silently ignored

	prodArgs, stagingArgs := cyberDojoProdAndStagingArgs()

	prodOutput, _, prodErr := runBufferedInnerMain(prodArgs)
	fmt.Print(prodOutput)

	stagingOutput, stagingGlobal, stagingErr := runBufferedInnerMain(stagingArgs)
	if stagingGlobal.Debug {
		fmt.Print(stagingOutput)
	}

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

func cyberDojoProdAndStagingArgs() ([]string, []string) {
	// The CLI execution is a double-host, double-api-token, org==cyber-dojo
	// Return two []string args, modified so a call with those args targets two hosts:
	//  - https://app.kosli.com
	//  - https://staging.app.kosli.com

	hosts := getHosts()
	apiTokens := getApiTokens()

	argsAppendHostApiToken := func(n int) []string {
		// No need to strip existing --host/--api-token flags from os.Args
		// as we are appending new flag values which take precedence.
		hostProd := fmt.Sprintf("--host=%s", hosts[n])
		apiTokenProd := fmt.Sprintf("--api-token=%s", apiTokens[n])
		return append(os.Args, hostProd, apiTokenProd)
	}

	argsProd := argsAppendHostApiToken(0)
	argsStaging := argsAppendHostApiToken(1)
	return argsProd, argsStaging
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
	err := inner_main(os.Args)
	return fmt.Sprint(&buffer), global, err
}

func getOrgs() []string {
	g := func() string { return global.Org }
	return splitGlobal(g)
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
