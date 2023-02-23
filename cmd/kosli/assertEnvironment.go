package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const assertEnvironmentShortDesc = `Assert the compliance status of an environment in Kosli.`

const assertEnvironmentLongDesc = assertEnvironmentShortDesc + `
Exits with non-zero code if the environment has a non-compliant status.
The expected argument is an expression to specify the specific environment snapshot to assert.
It has the format <ENVIRONMENT_NAME>[SEPARATOR][SNAPSHOT_REFERENCE] 

Separators can be:
- '#' to specify a specific snapshot number for the environment that is being asserted.
- '~' to get N-th behind the latest snapshot.

Examples of valid expressions are: 
- prod (latest snapshot of prod)
- prod#10 (snapshot number 10 of prod)
- prod~2 (third latest snapshot of prod)
`

const assertEnvironmentExample = `
kosli assert environment prod#5 \
	--api-token yourAPIToken \
	--owner yourOrgName
`

func newAssertEnvironmentCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment ENVIRONMENT-NAME-OR-EXPRESSION",
		Aliases: []string{"env"},
		Short:   assertEnvironmentShortDesc,
		Long:    assertEnvironmentLongDesc,
		Example: assertEnvironmentExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(out, args)
		},
	}
	addDryRunFlag(cmd)

	return cmd
}

func run(out io.Writer, args []string) error {
	var err error

	url := fmt.Sprintf("%s/api/v1/environments/%s/snapshots/%s", global.Host, global.Owner, url.QueryEscape(args[0]))

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	var environmentData map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &environmentData)
	if err != nil {
		return err
	}

	if environmentData["compliant"].(bool) {
		logger.Info("COMPLIANT")
	} else {
		return fmt.Errorf("INCOMPLIANT")
	}

	return nil
}
