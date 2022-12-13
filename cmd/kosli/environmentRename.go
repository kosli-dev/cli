package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentRenameShortDesc = `Rename a Kosli environment. `

const environmentRenameLongDesc = environmentRenameShortDesc + `
The environment will remain available under its old name until that name is taken by another environment.
`

const environmentRenameExample = `
# rename a Kosli environment:
kosli environment rename oldName newName \
	--api-token yourAPIToken \
	--owner yourOrgName 
`

type RenameEnvironmentPayload struct {
	NewName string `json:"new_name"`
}

func newEnvironmentRenameCmd(out io.Writer) *cobra.Command {
	payload := new(RenameEnvironmentPayload)
	cmd := &cobra.Command{
		Use:     "rename OLD_NAME NEW_NAME",
		Short:   environmentRenameShortDesc,
		Long:    environmentRenameLongDesc,
		Example: environmentRenameExample,
		Args:    cobra.MinimumNArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v1/environments/%s/%s/rename", global.Host, global.Owner, args[0])
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
				logger.Info("environment %s was renamed to %s", args[0], payload.NewName)
			}
			return err
		},
	}
	return cmd
}
