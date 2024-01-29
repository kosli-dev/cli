package main

import (
	"io"

	"github.com/spf13/cobra"
)

const enableBetaDesc = `Enable beta features for an organization.`

func newEnableBetaCmd(out io.Writer) *cobra.Command {
	o := new(betaOptions)
	cmd := &cobra.Command{
		Use:     "beta",
		Aliases: []string{"experimental"},
		Short:   enableBetaDesc,
		Long:    enableBetaDesc,
		Args:    cobra.NoArgs,
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
