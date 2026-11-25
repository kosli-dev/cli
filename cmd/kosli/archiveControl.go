package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/kosli"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const archiveControlShortDesc = `Archive a Kosli control.`

const archiveControlLongDesc = archiveControlShortDesc + `
An archived control is no longer active. It remains visible via ^kosli get control^ and
via ^kosli list controls --archived^, and can be restored with ^kosli unarchive control^.
`

const archiveControlExample = `
# archive a Kosli control:
kosli archive control yourControlIdentifier \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newArchiveControlCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "control CONTROL-IDENTIFIER",
		Short:   archiveControlShortDesc,
		Long:    archiveControlLongDesc,
		Example: archiveControlExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return kosliAPI().Send(func(h kosli.Host) error {
				url, err := url.JoinPath(h.URL, "api/v2/controls", global.Org, args[0], "archive")
				if err != nil {
					return err
				}

				reqParams := &requests.RequestParams{
					Method: http.MethodPost,
					URL:    url,
					DryRun: h.DryRun,
					Token:  h.Token,
				}
				_, err = h.Do(reqParams)
				if err == nil && !h.DryRun {
					h.Info("control %s was archived", args[0])
				}
				return err
			})
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
