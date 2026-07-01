package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const unarchiveControlShortDesc = `Unarchive a Kosli control.`

const unarchiveControlLongDesc = unarchiveControlShortDesc + `
Restores a previously archived control to the active state.
`

const unarchiveControlExample = `
# unarchive a Kosli control:
kosli unarchive control yourControlIdentifier \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newUnarchiveControlCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "control CONTROL-IDENTIFIER",
		Short:       unarchiveControlShortDesc,
		Long:        unarchiveControlLongDesc,
		Example:     unarchiveControlExample,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url, err := url.JoinPath(global.Host, "api/v2/controls", global.Org, args[0], "unarchive")
			if err != nil {
				return err
			}

			reqParams := &requests.RequestParams{
				Method: http.MethodPost,
				URL:    url,
				DryRun: global.DryRun,
				Token:  global.ApiToken,
			}
			_, err = kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("control %s was unarchived", args[0])
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
