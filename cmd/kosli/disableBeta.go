package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const disableBetaDesc = `Disable beta features for an organization.`

const disableLongBetaDesc = disableBetaDesc + `
Currently, the only beta feature is audit-trails.
`

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
		Long:    disableLongBetaDesc,
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
	var err error
	url := fmt.Sprintf("%s/api/v2/organizations/%s/experimental_features", global.Host, global.Org)
	action := "enabled"
	if !o.payload.Enabled {
		action = "disabled"
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("beta features have been %s for organization: %s", action, global.Org)
	}
	return err
}
