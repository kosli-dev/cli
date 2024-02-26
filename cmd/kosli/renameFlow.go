package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const renameFlowShortDesc = `Rename a Kosli flow.`

const renameFlowLongDesc = renameFlowShortDesc + `
The flow will remain accessible under its old name until that name is taken by another flow.
`

const renameFlowExample = `
# rename a Kosli flow:
kosli rename flow oldName newName \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type RenameFlowPayload struct {
	NewName string `json:"new_name"`
}

func newRenameFlowCmd(out io.Writer) *cobra.Command {
	payload := new(RenameFlowPayload)
	cmd := &cobra.Command{
		Use:     "flow OLD_NAME NEW_NAME",
		Short:   renameFlowShortDesc,
		Long:    renameFlowLongDesc,
		Example: renameFlowExample,
		Args:    cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v2/flows/%s/%s/rename", global.Host, global.Org, args[0])
			payload.NewName = args[1]

			reqParams := &requests.RequestParams{
				Method:   http.MethodPut,
				URL:      url,
				Payload:  payload,
				DryRun:   global.DryRun,
				Password: global.ApiToken,
			}
			_, err := kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("flow %s was renamed to %s", args[0], payload.NewName)
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
