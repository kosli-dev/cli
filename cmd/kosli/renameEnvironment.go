package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const renameEnvironmentShortDesc = `Rename a Kosli environment.`

const renameEnvironmentLongDesc = renameEnvironmentShortDesc + `
The environment will remain accessible under its old name until that name is taken by another environment.
`

const renameEnvironmentExample = `
# rename a Kosli environment:
kosli rename environment oldName newName \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type RenameEnvironmentPayload struct {
	NewName string `json:"new_name"`
}

func newRenameEnvironmentCmd(out io.Writer) *cobra.Command {
	payload := new(RenameEnvironmentPayload)
	cmd := &cobra.Command{
		Use:     "environment OLD_NAME NEW_NAME",
		Aliases: []string{"env"},
		Short:   renameEnvironmentShortDesc,
		Long:    renameEnvironmentLongDesc,
		Example: renameEnvironmentExample,
		Args:    cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v2/environments/%s/%s/rename", global.Host, global.Org, args[0])
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
	addDryRunFlag(cmd)
	return cmd
}
