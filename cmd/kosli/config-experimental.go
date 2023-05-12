package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const enableDesc = `All Kosli feature toggles commands.`

type enableOptions struct {
	enable  bool
	disable bool
	payload experimentalFeaturesPayload
}

type experimentalFeaturesPayload struct {
	Enabled bool `json:"experimental_features_enabled"`
}

func newConfigExperimentalCmd(out io.Writer) *cobra.Command {
	o := new(enableOptions)
	cmd := &cobra.Command{
		Use:    "config-experimental",
		Short:  enableDesc,
		Long:   enableDesc,
		Hidden: true,
		Args:   cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().BoolVar(&o.disable, "disable", false, experimentalDisableFlag)
	cmd.Flags().BoolVar(&o.enable, "enable", true, experimentalEnableFlag)

	return cmd
}

func (o *enableOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/organizations/%s/experimental_features", global.Host, global.Org)
	action := "enabled"
	if o.enable {
		o.payload.Enabled = true
	}
	if o.disable {
		o.payload.Enabled = false
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
		logger.Info("experimental features have been %s for organization: %s", action, global.Org)
	}
	return err
}
