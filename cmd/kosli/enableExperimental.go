package main

import (
	"io"

	"github.com/spf13/cobra"
)

const enableExperimentalDesc = `Enable experimental features.`

func newEnableExperimentalCmd(out io.Writer) *cobra.Command {
	o := new(experimentalOptions)
	cmd := &cobra.Command{
		Use:   "experimental",
		Short: enableExperimentalDesc,
		Long:  enableExperimentalDesc,
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.payload.Enabled = true
			return o.run(args)
		},
	}

	return cmd
}
