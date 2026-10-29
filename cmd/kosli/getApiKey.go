package main

import (
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getApiKeyShortDesc = `Get an API key's metadata for a service account, or one of your own personal API keys.`

const getApiKeyLongDesc = getApiKeyShortDesc + `

Only the metadata of the API key is returned; the key value itself is never
returned (it is only shown once, at creation or rotation time).`

const getApiKeyExample = `
# get the metadata of an API key:
kosli get api-key yourApiKeyID \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# get the metadata of one of your own personal API keys:
kosli get api-key yourApiKeyID \
	--personal \
	--api-token yourAPIToken
`

type getApiKeyOptions struct {
	owner  apiKeyOwner
	output string
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
			if err := o.owner.validate(); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	o.owner.addFlags(cmd)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *getApiKeyOptions) run(out io.Writer, args []string) error {
	url, err := o.owner.url(args[0])
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
