package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportDesc = `All Kosli report commands.`

func newReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "report",
		Short:      reportDesc,
		Long:       reportDesc,
		Deprecated: "this command is deprecated and will be removed in a future release.",
	}

	// Add subcommands
	cmd.AddCommand(
		newReportArtifactCmd(out),
		newReportApprovalCmd(out),
	)

	return cmd
}
