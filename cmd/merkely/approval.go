package main

import (
	"io"

	"github.com/spf13/cobra"
)

const approvalDesc = `All approvals operations in a Merkely pipeline.`

func newApprovalCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval",
		Short: approvalDesc,
		Long:  approvalDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newApprovalRequestCmd(out),
		newApprovalReportCmd(out),
		newApprovalAssertCmd(out),
	)

	return cmd
}
