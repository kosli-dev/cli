package main

import (
	"io"

	"github.com/spf13/cobra"
)

const environmentReportDesc = `
Report artifacts running in an environemt to Kosli.
`

func newEnvironmentReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report artifacts running in an environemt to Kosli.",
		Long:  environmentReportDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvironmentReportK8SCmd(out),
		newEnvironmentReportECSCmd(out),
		newEnvironmentReportServerCmd(out),
		newEnvironmentReportS3Cmd(out),
		newEnvironmentReportLambdaCmd(out),
	)

	return cmd
}
