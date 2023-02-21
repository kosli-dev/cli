package main

import (
	"io"

	"github.com/spf13/cobra"
)

const listDesc = `All Kosli list commands.`

func newListCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: listDesc,
		Long:  listDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newListApprovalsCmd(out),
	)

	return cmd
}
