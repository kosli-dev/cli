package main

import (
	"io"

	"github.com/spf13/cobra"
)

const artifactEvidenceDesc = `All artifacts evidence operations in a Merkely pipeline.`

func newArtifactEvidenceCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evidence",
		Short: artifactEvidenceDesc,
		Long:  artifactEvidenceDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newGenericEvidenceCmd(out),
		newTestEvidenceCmd(out),
		newPullRequestEvidenceCmd(out),
		newPullRequestEvidenceGithubCmd(out),
	)

	return cmd
}
