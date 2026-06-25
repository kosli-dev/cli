package main

import (
	"io"

	"github.com/spf13/cobra"
)

const deleteDesc = `All Kosli delete commands.`

func newDeleteCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d", "de", "rm", "remove"},
		Short:   deleteDesc,
		Long:    deleteDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newDeleteApiKeyCmd(out),
		newDeleteServiceAccountCmd(out),
	)

	return cmd
}
