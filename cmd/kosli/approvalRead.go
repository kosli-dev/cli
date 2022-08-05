package main

import (
	"io"

	"github.com/spf13/cobra"
)

const approvalReadDesc = `All approval read operations.`

func newApprovalReadCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval",
		Short: approvalReadDesc,
		Long:  approvalReadDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newApprovalLsCmd(out),
	)

	return cmd
}
