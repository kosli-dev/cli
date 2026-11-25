package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/kosli"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const updateDefaultOrgShortDesc = `Set the default organization for the current user.`

const updateDefaultOrgLongDesc = updateDefaultOrgShortDesc + `
The default organization is the one selected by default in the Kosli Web UI when you log in.
`

const updateDefaultOrgExample = `
# set the default organization for the current user:
kosli update default-org yourOrgName \
	--api-token yourAPIToken
`

func newUpdateDefaultOrgCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "default-org ORG-NAME",
		Short:   updateDefaultOrgShortDesc,
		Long:    updateDefaultOrgLongDesc,
		Example: updateDefaultOrgExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			if args[0] == "" {
				return ErrorBeforePrintingUsage(cmd, "ORG-NAME argument is required")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return kosliAPI().Send(func(h kosli.Host) error {
				url, err := url.JoinPath(h.URL, "api/v2/user", args[0])
				if err != nil {
					return err
				}

				reqParams := &requests.RequestParams{
					Method: http.MethodPut,
					URL:    url,
					DryRun: h.DryRun,
					Token:  h.Token,
				}
				_, err = h.Do(reqParams)
				if err == nil && !h.DryRun {
					h.Info("default organization is set to: %s", args[0])
				}
				return err
			})
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
