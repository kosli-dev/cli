package main

import (
	"io"

	"github.com/spf13/cobra"
)

const dismissDesc = `All Kosli dismiss commands.`

func newDismissCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dismiss",
		Aliases: []string{"dism", "dm"},
		Short:   dismissDesc,
		Long:    dismissDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newDismissProductUpdatesCmd(out),
	)

	return cmd
}
