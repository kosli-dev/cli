package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const renameEnvironmentShortDesc = `Rename a Kosli environment.`

const renameEnvironmentLongDesc = renameEnvironmentShortDesc + `

After renaming, the old name still resolves to this environment for both reads and writes:
^get^, ^list^ and similar queries return it, and ^snapshot^ commands that still use the
old name report into it. The only operation that requires the current name is a further rename.

The old name stops resolving once another environment is created with, or renamed to, that name.
From then on, commands using the old name act on the new environment.
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
			url, err := url.JoinPath(global.Host, "api/v2/environments", global.Org, args[0], "rename")
			if err != nil {
				return err
			}
			payload.NewName = args[1]

			reqParams := &requests.RequestParams{
				Method:  http.MethodPut,
				URL:     url,
				Payload: payload,
				DryRun:  global.DryRun,
				Token:   global.ApiToken,
			}
			_, err = kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("environment %s was renamed to %s", args[0], payload.NewName)
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
