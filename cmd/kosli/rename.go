package main

import (
	"io"

	"github.com/spf13/cobra"
)

const renameDesc = `All Kosli rename commands.`

func newRenameCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rename",
		Aliases: []string{"re"},
		Short:   renameDesc,
		Long:    renameDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newRenameEnvironmentCmd(out),
		newRenameFlowCmd(out),
	)
	return cmd
}
