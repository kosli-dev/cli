package main

import (
	"io"

	"github.com/spf13/cobra"
)

const requestDesc = `All Kosli request commands.`

func newRequestCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "request",
		Short:      requestDesc,
		Long:       requestDesc,
		Deprecated: "this command is deprecated and will be removed in a future release.",
	}

	// Add subcommands
	cmd.AddCommand(
		newRequestApprovalCmd(out),
	)

	return cmd
}
