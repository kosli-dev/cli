package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/kosli"
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

func newArchiveEnvironmentCmd(out io.Writer) *cobra.Command {
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
			return kosliAPI().Send(func(h kosli.Host) error {
				url, err := url.JoinPath(h.URL, "api/v2/environments", global.Org, args[0], "archive")
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
					h.Info("environment %s was archived", args[0])
				}
				return err
			})
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
