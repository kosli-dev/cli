package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const disableBetaDesc = `Disable beta features for an organization.`

type betaOptions struct {
	payload betaFeaturesPayload
}

type betaFeaturesPayload struct {
	Enabled bool `json:"experimental_features_enabled"`
}

func newDisableExperimentalCmd(out io.Writer) *cobra.Command {
	o := new(betaOptions)
	cmd := &cobra.Command{
		Use:     "beta",
		Aliases: []string{"experimental"},
		Short:   disableBetaDesc,
		Long:    disableBetaDesc,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.payload.Enabled = false
			return o.run(args)
		},
	}

	return cmd
}

func (o *betaOptions) run(args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/organizations", global.Org, "experimental_features")
	if err != nil {
		return err
	}
	action := "enabled"
	if !o.payload.Enabled {
		action = "disabled"
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("beta features have been %s for organization: %s", action, global.Org)
	}
	return err
}
