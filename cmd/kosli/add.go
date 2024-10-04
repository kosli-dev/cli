package main

import (
	"io"

	"github.com/spf13/cobra"
)

const addDesc = `All Kosli add commands.`

func newAddCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: addDesc,
		Long:  addDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAddEnvironmentCmd(out),
	)
	return cmd
}
