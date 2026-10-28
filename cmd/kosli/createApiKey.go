package main

import (
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createApiKeyShortDesc = `Create an API key for a service account, or a personal API key for yourself.`

const createApiKeyLongDesc = createApiKeyShortDesc + `

The key value is only returned once, at creation time, so make sure to store it securely.

Every API key expires. Without ^--expires-at^ the key is given the maximum allowed
lifetime, and an ^--expires-at^ beyond that maximum is capped to it.

With ^--personal^, a key created using another API key is revoked along with that key,
and expires no later than it.`

const createApiKeyExample = `
# create an API key for a service account:
kosli create api-key \
	--service-account yourServiceAccountName \
	--description "key for CI" \
	--api-token yourAPIToken \
	--org yourOrgName

# create an API key that expires on a given date:
kosli create api-key \
	--service-account yourServiceAccountName \
	--description "key for CI" \
	--expires-at 2026-12-31 \
	--api-token yourAPIToken \
	--org yourOrgName

# create a personal API key for yourself:
kosli create api-key \
	--personal \
	--description "key for CI" \
	--api-token yourAPIToken
`

type createApiKeyOptions struct {
	owner     apiKeyOwner
	expiresAt string
	output    string
	payload   createApiKeyPayload
}

type createApiKeyPayload struct {
	Description string `json:"description"`
	ExpiresAt   *int64 `json:"expires_at,omitempty"`
}

func newCreateApiKeyCmd(out io.Writer) *cobra.Command {
	o := new(createApiKeyOptions)
	cmd := &cobra.Command{
		Use:     "api-key",
		Aliases: []string{"ak"},
		Short:   createApiKeyShortDesc,
		Long:    createApiKeyLongDesc,
		Example: createApiKeyExample,
		Args:    cobra.NoArgs,
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
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", apiKeyDescriptionFlag)
	cmd.Flags().StringVarP(&o.expiresAt, "expires-at", "e", "", createApiKeyExpiresAtFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"description"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createApiKeyOptions) run(out io.Writer, args []string) error {
	url, err := o.owner.url()
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
