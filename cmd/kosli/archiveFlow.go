package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const archiveFlowShortDesc = `Archive a Kosli flow.`

const archiveFlowLongDesc = archiveFlowShortDesc + `
The flow will no longer be visible in list of flows, data is still stored in the database.
`

const archiveFlowExample = `
# archive a Kosli flow:
kosli archive flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName 
`

func newArchiveFlowCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "flow FLOW-NAME",
		Short:   archiveFlowShortDesc,
		Long:    archiveFlowLongDesc,
		Example: archiveFlowExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v2/flows/%s/%s/archive", global.Host, global.Org, args[0])

			reqParams := &requests.RequestParams{
				Method: http.MethodPut,
				URL:    url,
				DryRun: global.DryRun,
				Token:  global.ApiToken,
			}
			_, err := kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("flow %s was archived", args[0])
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
