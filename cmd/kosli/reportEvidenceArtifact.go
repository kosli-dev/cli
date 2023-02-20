package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportEvidenceArtifactDesc = `All Kosli evidence commands.`

func newReportEvidenceArtifactCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: reportEvidenceArtifactDesc,
		Long:  reportEvidenceArtifactDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newReportEvidenceArtifactPRCmd(out),
		newGenericEvidenceCmd(out),
		newJUnitEvidenceCmd(out),
	)

	return cmd
}
