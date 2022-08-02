package main

import (
	"io"

	"github.com/spf13/cobra"
)

const artifactReadDesc = `All artifacts read operations in a Kosli pipeline.`

func newArtifactReadCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: artifactReadDesc,
		Long:  artifactReadDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newArtifactGetCmd(out),
		newArtifactLsCmd(out),
	)

	return cmd
}
