package main

import (
	"io"

	"github.com/spf13/cobra"
)

const diffDesc = `All Kosli diff commands.`

func newDiffCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: diffDesc,
		Long:  diffDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newDiffSnapshotsCmd(out),
	)
	return cmd
}
