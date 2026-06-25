package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getServiceAccountShortDesc = `Get a service account's metadata.`

const getServiceAccountExample = `
# get the metadata of a service account:
kosli get service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName
`

type getServiceAccountOptions struct {
	output string
}

func newGetServiceAccountCmd(out io.Writer) *cobra.Command {
	o := new(getServiceAccountOptions)
	cmd := &cobra.Command{
		Use:     "service-account SERVICE-ACCOUNT-NAME",
		Aliases: []string{"sa"},
		Short:   getServiceAccountShortDesc,
		Long:    getServiceAccountShortDesc,
		Example: getServiceAccountExample,
		Args:    cobra.ExactArgs(1),
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

func (o *getServiceAccountOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, args[0])
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
			"table": printServiceAccountAsTable,
			"json":  output.PrintJson,
		})
}
