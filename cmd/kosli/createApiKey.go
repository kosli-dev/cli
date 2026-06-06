package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createApiKeyShortDesc = `Create an API key for a service account.`

const createApiKeyLongDesc = createApiKeyShortDesc + `

The key value is only returned once, at creation time, so make sure to store it securely.`

const createApiKeyExample = `
# create an API key for a service account:
kosli service-account api-keys create \
	--service-account yourServiceAccountName \
	--description "key for CI" \
	--api-token yourAPIToken \
	--org yourOrgName

# create an API key that expires on a given date:
kosli service-account api-keys create \
	--service-account yourServiceAccountName \
	--description "key for CI" \
	--expires-at 2026-12-31 \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createApiKeyOptions struct {
	serviceAccount string
	expiresAt      string
	output         string
	payload        createApiKeyPayload
}

type createApiKeyPayload struct {
	Description string `json:"description"`
	ExpiresAt   *int64 `json:"expires_at,omitempty"`
}

func newCreateApiKeyCmd(out io.Writer) *cobra.Command {
	o := new(createApiKeyOptions)
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c", "cr"},
		Short:   createApiKeyShortDesc,
		Long:    createApiKeyLongDesc,
		Example: createApiKeyExample,
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
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", apiKeyDescriptionFlag)
	cmd.Flags().StringVarP(&o.expiresAt, "expires-at", "e", "", apiKeyExpiresAtFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"service-account", "description"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createApiKeyOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, o.serviceAccount, "api-keys")
	if err != nil {
		return err
	}

	if o.expiresAt != "" {
		expiresAt, err := parseExpiresAt(o.expiresAt)
		if err != nil {
			return err
		}
		o.payload.ExpiresAt = &expiresAt
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPost,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil || global.DryRun {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printApiKeyAsTable,
			"json":  output.PrintJson,
		})
}
