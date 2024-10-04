package main

import (
	"io"

	"github.com/spf13/cobra"
)

const joinDesc = `All Kosli join commands.`

func newJoinCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: joinDesc,
		Long:  joinDesc,
	}

	// Join subcommands
	cmd.AddCommand(
		newJoinEnvironmentCmd(out),
	)
	return cmd
}
