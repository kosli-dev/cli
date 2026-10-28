package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listApiKeysShortDesc = `List API keys for a service account, or your own personal API keys.`

const listApiKeysLongDesc = listApiKeysShortDesc + `

Only the metadata of each active API key is returned; the key values themselves are never
listed (they are only shown once, at creation or rotation time).`

const listApiKeysExample = `
# list the API keys for a service account:
kosli list api-keys \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# list your own personal API keys:
kosli list api-keys \
	--personal \
	--api-token yourAPIToken
`

type listApiKeysOptions struct {
	owner  apiKeyOwner
	output string
}

func newListApiKeysCmd(out io.Writer) *cobra.Command {
	o := new(listApiKeysOptions)
	cmd := &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"ak", "aks", "api-key"},
		Short:   listApiKeysShortDesc,
		Long:    listApiKeysLongDesc,
		Example: listApiKeysExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, o.owner.requiredGlobalFlags()); err != nil {
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

func (o *listApiKeysOptions) run(out io.Writer, args []string) error {
	url, err := o.owner.url()
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
			"table": printApiKeysListAsTable,
			"json":  output.PrintJson,
		})
}

func printApiKeysListAsTable(raw string, out io.Writer, page int) error {
	var keys []apiKeyMetadata
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return err
	}

	if len(keys) == 0 {
		logger.Info("No API keys were found.")
		return nil
	}

	header := []string{"ID", "DESCRIPTION", "CREATED", "EXPIRES", "LAST USED"}
	rows := []string{}
	for _, key := range keys {
		createdAt, expiresAt, lastUsedAt, err := apiKeyMetadataTimestamps(key)
		if err != nil {
			return err
		}

		row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", key.Id, key.Description, createdAt, expiresAt, lastUsedAt)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)
	return nil
}
