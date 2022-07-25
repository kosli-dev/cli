package main

import (
	"io"

	"github.com/spf13/cobra"
)

const artifactReportDesc = `All artifacts reporting operations in a Kosli pipeline.`

func newArtifactReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: artifactReportDesc,
		Long:  artifactReportDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newArtifactCreationCmd(out),
		newArtifactEvidenceCmd(out),
	)

	return cmd
}
