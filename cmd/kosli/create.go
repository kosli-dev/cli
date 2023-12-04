package main

import (
	"io"

	"github.com/spf13/cobra"
)

const createDesc = `All Kosli create commands.`

func newCreateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: createDesc,
		Long:  createDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newCreateEnvironmentCmd(out),
		newCreateFlowCmd(out),
		newCreateAuditTrailCmd(out),
		newCreateFlowWithTemplateCmd(out),
	)
	return cmd
}
