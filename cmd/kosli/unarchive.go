package main

import (
	"io"

	"github.com/spf13/cobra"
)

const unarchiveDesc = `All Kosli unarchive commands.`

func newUnarchiveCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive",
		Short: unarchiveDesc,
		Long:  unarchiveDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newUnarchiveControlCmd(out),
	)
	return cmd
}
