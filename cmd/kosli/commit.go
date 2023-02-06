package main

import (
	"io"

	"github.com/spf13/cobra"
)

const commitDesc = `All commits report operations in a Kosli pipeline.`

func newCommitCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: commitDesc,
		Long:  commitDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newCommitReportCmd(out),
	)

	return cmd
}
