package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const archiveEnvironmentShortDesc = `Archive a Kosli environment.`

const archiveEnvironmentLongDesc = archiveEnvironmentShortDesc + `
The environment will no longer be visible in list of environments, data is still stored in the database.
`

const archiveEnvironmentExample = `
# archive a Kosli environment:
kosli archive environment yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type ArchiveEnvironmentPayload struct {
}

func newArchiveEnvironmentCmd(out io.Writer) *cobra.Command {
	payload := new(ArchiveEnvironmentPayload)
	cmd := &cobra.Command{
		Use:     "environment ENVIRONMENT-NAME",
		Short:   archiveEnvironmentShortDesc,
		Long:    archiveEnvironmentLongDesc,
		Example: archiveEnvironmentExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v2/environments/%s/%s/archive", global.Host, global.Org, args[0])

			reqParams := &requests.RequestParams{
				Method:  http.MethodPut,
				URL:     url,
				Payload: payload,
				DryRun:  global.DryRun,
				Token:   global.ApiToken,
			}
			_, err := kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("environment %s was archived", args[0])
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
