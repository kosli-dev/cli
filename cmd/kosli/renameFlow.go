package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const renameFlowShortDesc = `Rename a Kosli flow.`

const renameFlowLongDesc = renameFlowShortDesc + `

After renaming, the old name still resolves to this flow for both reads and writes:
^kosli get flow^, ^kosli list trails^ and similar queries return it, and ^kosli begin trail^
and ^kosli attest^ commands that still use the old name record into it rather than
auto-creating a new flow.
The only operation that requires the current name is a further rename.

The old name stops resolving once another flow is created with, or renamed to, that name.
From then on, commands using the old name act on the new flow.
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
			url, err := url.JoinPath(global.Host, "api/v2/flows", global.Org, args[0], "rename")
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
				logger.Info("flow %s was renamed to %s", args[0], payload.NewName)
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
