package main

import (
	"io"

	"github.com/spf13/cobra"
)

const requestDesc = `All Kosli request commands.`

func newRequestCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: reportDesc,
		Long:  reportDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newRequestApprovalCmd(out),
	)

	return cmd
}
