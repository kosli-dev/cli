package main

import (
	"io"

	"github.com/spf13/cobra"
)

const commitReportDesc = `All commits reporting operations in a Kosli pipeline.`

func newCommitReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: commitReportDesc,
		Long:  commitReportDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newCommitEvidenceCmd(out),
	)

	return cmd
}
