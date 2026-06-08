package main

import (
	"io"

	"github.com/spf13/cobra"
)

const updateDesc = `All Kosli update commands.`

func newUpdateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"up", "u"},
		Short:   updateDesc,
		Long:    updateDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newUpdateApiKeyCmd(out),
	)

	return cmd
}
