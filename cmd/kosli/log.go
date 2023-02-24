package main

import (
	"io"

	"github.com/spf13/cobra"
)

const logDesc = `All Kosli log commands.`

func newLogCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: logDesc,
		Long:  logDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newLogEnvironmentCmd(out),
	)

	return cmd
}
