package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listServiceAccountsShortDesc = `List service accounts in an organization.`

const listServiceAccountsExample = `
# list the service accounts in an organization:
kosli list service-accounts \
	--api-token yourAPIToken \
	--org yourOrgName
`

type listServiceAccountsOptions struct {
	output string
}

func newListServiceAccountsCmd(out io.Writer) *cobra.Command {
	o := new(listServiceAccountsOptions)
	cmd := &cobra.Command{
		Use:     "service-accounts",
		Aliases: []string{"sa", "sas", "service-account"},
		Short:   listServiceAccountsShortDesc,
		Long:    listServiceAccountsShortDesc,
		Example: listServiceAccountsExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *listServiceAccountsOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printServiceAccountsListAsTable,
			"json":  output.PrintJson,
		})
}
