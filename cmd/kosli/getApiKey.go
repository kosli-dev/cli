package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getApiKeyShortDesc = `Get an API key's metadata for a service account.`

const getApiKeyLongDesc = getApiKeyShortDesc + `

Only the metadata of the API key is returned; the key value itself is never
returned (it is only shown once, at creation or rotation time).`

const getApiKeyExample = `
# get the metadata of an API key:
kosli get api-key yourApiKeyID \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName
`

type getApiKeyOptions struct {
	serviceAccount string
	output         string
}

func newGetApiKeyCmd(out io.Writer) *cobra.Command {
	o := new(getApiKeyOptions)
	cmd := &cobra.Command{
		Use:     "api-key KEY-ID",
		Aliases: []string{"ak"},
		Short:   getApiKeyShortDesc,
		Long:    getApiKeyLongDesc,
		Example: getApiKeyExample,
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

	cmd.Flags().StringVarP(&o.serviceAccount, "service-account", "s", "", serviceAccountNameFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	err := RequireFlags(cmd, []string{"service-account"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *getApiKeyOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, o.serviceAccount, "api-keys", args[0])
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
			"table": printApiKeyMetadataAsTable,
			"json":  output.PrintJson,
		})
}
