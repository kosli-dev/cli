package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportDesc = `All Kosli report commands.`

func newReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: reportDesc,
		Long:  reportDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newReportArtifactCmd(out),
		newReportEvidenceCmd(out),
		newReportApprovalCmd(out),
		newReportWorkflowCmd(out),
	)

	return cmd
}
