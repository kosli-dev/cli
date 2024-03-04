package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportEvidenceDesc = `All Kosli report evidence commands.`

func newReportEvidenceCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "evidence",
		Short:      reportEvidenceDesc,
		Long:       reportEvidenceDesc,
		Deprecated: "see kosli attest commands",
	}

	// Add subcommands
	cmd.AddCommand(
		newReportEvidenceArtifactCmd(out),
		newReportEvidenceCommitCmd(out),
	)

	return cmd
}
