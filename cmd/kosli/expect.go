package main

import (
	"io"

	"github.com/spf13/cobra"
)

const expectDesc = `All Kosli expect commands.`

func newExpectCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expect",
		Short: expectDesc,
		Long:  expectDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newExpectDeploymentCmd(out),
	)

	return cmd
}
