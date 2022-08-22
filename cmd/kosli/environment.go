package main

import (
	"io"

	"github.com/spf13/cobra"
)

const environmentDesc = `All environment operations in Kosli.`

func newEnvironmentCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"env"},
		Short:   environmentDesc,
		Long:    environmentDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvironmentDeclareCmd(out),
		newEnvironmentReportCmd(out),
		newEnvironmentLsCmd(out),
		newEnvironmentDiffCmd(out),
		newAllowedArtifactsCmd(out),
		newEnvironmentInspectCmd(out),
		newEnvironmentEventsLogCmd(out),
		newEnvironmentGetCmd(out),
		newEnvironmentLogCmd(out),
	)

	return cmd
}
