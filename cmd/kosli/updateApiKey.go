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

const updateApiKeyShortDesc = `Update one or more API keys for a service account.`

const updateApiKeyLongDesc = updateApiKeyShortDesc + `

Currently the only supported update is rotation, requested with ^--rotate^.

When rotating, a new API key is generated immediately. The old key remains valid for a
grace period to allow time to update dependent systems; the length of that grace period is
server-managed unless overridden with ^--grace-period-hours^. The new key value is only
returned once, so make sure to store it securely.`

const updateApiKeyExample = `
# rotate an API key for a service account:
kosli update api-key yourApiKeyID \
	--rotate \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# rotate multiple API keys at once:
kosli update api-key keyID1 keyID2 \
	--rotate \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# rotate an API key, keeping the old key valid for 48 hours:
kosli update api-key yourApiKeyID \
	--rotate \
	--grace-period-hours 48 \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName
`

type updateApiKeyOptions struct {
	serviceAccount      string
	rotate              bool
	expiresAt           string
	gracePeriodHours    int
	gracePeriodHoursSet bool
	output              string
	payload             rotateApiKeyPayload
}

type rotateApiKeyPayload struct {
	GracePeriodHours *int   `json:"grace_period_hours,omitempty"`
	ExpiresAt        *int64 `json:"expires_at,omitempty"`
}

func newUpdateApiKeyCmd(out io.Writer) *cobra.Command {
	o := new(updateApiKeyOptions)
	cmd := &cobra.Command{
		Use:     "api-key KEY-ID [KEY-ID...]",
		Aliases: []string{"ak"},
		Short:   updateApiKeyShortDesc,
		Long:    updateApiKeyLongDesc,
		Example: updateApiKeyExample,
		Args:    cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.gracePeriodHoursSet = cmd.Flags().Changed("grace-period-hours")
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.serviceAccount, "service-account", "s", "", serviceAccountNameFlag)
	cmd.Flags().BoolVarP(&o.rotate, "rotate", "R", false, apiKeyRotateFlag)
	cmd.Flags().StringVarP(&o.expiresAt, "expires-at", "e", "", apiKeyExpiresAtFlag)
	cmd.Flags().IntVarP(&o.gracePeriodHours, "grace-period-hours", "g", 0, apiKeyGracePeriodHoursFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"service-account"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *updateApiKeyOptions) run(out io.Writer, args []string) error {
	if !o.rotate {
		return fmt.Errorf("nothing to update: use --rotate to rotate the API key(s)")
	}

	// Only send grace_period_hours when the user explicitly set it; otherwise
	// let the server apply its own default.
	if o.gracePeriodHoursSet {
		o.payload.GracePeriodHours = &o.gracePeriodHours
	}
	if o.expiresAt != "" {
		expiresAt, err := parseExpiresAt(o.expiresAt)
		if err != nil {
			return err
		}
		o.payload.ExpiresAt = &expiresAt
	}

	// Rotated key values are only returned once, so collect each successful
	// response and print what we have even if a later key fails (rather than
	// losing the new keys that were already rotated).
	keys := make([]json.RawMessage, 0, len(args))
	var runErr error
	for _, keyID := range args {
		url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, o.serviceAccount, "api-keys", keyID, "rotate")
		if err != nil {
			runErr = err
			break
		}

		reqParams := &requests.RequestParams{
			Method:  http.MethodPost,
			URL:     url,
			Payload: o.payload,
			DryRun:  global.DryRun,
			Token:   global.ApiToken,
		}
		response, err := kosliClient.Do(reqParams)
		if err != nil {
			runErr = err
			break
		}
		if !global.DryRun {
			keys = append(keys, json.RawMessage(response.Body))
		}
	}

	if !global.DryRun && len(keys) > 0 {
		raw, err := json.Marshal(keys)
		if err != nil {
			return err
		}
		if err := output.FormattedPrint(string(raw), o.output, out, 0,
			map[string]output.FormatOutputFunc{
				"table": printApiKeysAsTable,
				"json":  output.PrintJson,
			}); err != nil {
			return err
		}
	}

	return runErr
}
