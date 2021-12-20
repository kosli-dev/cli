package main

import (
	"io"

	"github.com/spf13/cobra"
)

const environmentReportDesc = `
Report artifacts running in an environemt to Merkely.
`

func newEnvironmentReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Report artifacts running in an environemt to Merkely.",
		Long:  environmentReportDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvironmentReportK8SCmd(out),
		newEnvironmentReportECSCmd(out),
		newEnvironmentReportServerCmd(out),
	)

	return cmd
}
