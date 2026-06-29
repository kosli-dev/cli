package main

import (
	"io"

	"github.com/spf13/cobra"
)

const updateDesc = `All Kosli update commands.`

func newUpdateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"u", "up"},
		Short:   updateDesc,
		Long:    updateDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newUpdateServiceAccountCmd(out),
		newUpdateDefaultOrgCmd(out),
	)

	return cmd
}
