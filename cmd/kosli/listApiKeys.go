package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listApiKeysShortDesc = `List API keys for a service account.`

const listApiKeysLongDesc = listApiKeysShortDesc + `

Only the metadata of each active API key is returned; the key values themselves are never
listed (they are only shown once, at creation or rotation time).`

const listApiKeysExample = `
# list the API keys for a service account:
kosli service-account api-keys list \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName
`

type listApiKeysOptions struct {
	serviceAccount string
	output         string
}

func newListApiKeysCmd(out io.Writer) *cobra.Command {
	o := new(listApiKeysOptions)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   listApiKeysShortDesc,
		Long:    listApiKeysLongDesc,
		Example: listApiKeysExample,
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

	cmd.Flags().StringVarP(&o.serviceAccount, "service-account", "s", "", serviceAccountNameFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	err := RequireFlags(cmd, []string{"service-account"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *listApiKeysOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, o.serviceAccount, "api-keys")
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
	var keys []map[string]interface{}
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
		createdAt, err := formattedTimestamp(key["created_at"], false)
		if err != nil {
			return err
		}
		expiresAt, err := formattedTimestamp(key["expires_at"], false)
		if err != nil {
			return err
		}
		lastUsedAt, err := formattedTimestamp(key["last_used_at"], false)
		if err != nil {
			return err
		}

		row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", key["id"], key["description"], createdAt, expiresAt, lastUsedAt)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)
	return nil
}
