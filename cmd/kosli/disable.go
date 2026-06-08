package main

import (
	"io"

	"github.com/spf13/cobra"
)

const disableDesc = `Kosli disable commands.`

func newDisableCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: disableDesc,
		Long:  disableDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newDisableExperimentalCmd(out),
	)

	return cmd
}
