package main

import (
	"io"

	"github.com/spf13/cobra"
)

const snapshotDesc = `All environments snapshot operations in Kosli.`

func newSnapshotCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "snapshot",
		Aliases: []string{"snap"},
		Short:   snapshotDesc,
		Long:    snapshotDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newSnapshotLsCmd(out),
	)

	return cmd
}
