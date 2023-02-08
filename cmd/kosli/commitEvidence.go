package main

import (
	"io"

	"github.com/spf13/cobra"
)

const commitEvidenceDesc = `All commit evidence operations in a Kosli pipeline.`

func newCommitEvidenceCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evidence",
		Short: commitEvidenceDesc,
		Long:  commitEvidenceDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newJUnitCommitEvidenceCmd(out),
		newSnykCommitEvidenceCmd(out),
		newGenericCommitEvidenceCmd(out),
		newPullRequestCommitEvidenceGithubCmd(out),
	)

	return cmd
}
