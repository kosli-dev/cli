package main

import (
	"io"

	"github.com/spf13/cobra"
)

const artifactEvidenceDesc = `All artifacts evidence operations in a Kosli pipeline.`

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
		newJUnitEvidenceCmd(out),
		newSnykEvidenceCmd(out),
		newPullRequestEvidenceBitbucketCmd(out),
		newPullRequestEvidenceGithubCmd(out),
	)

	return cmd
}
