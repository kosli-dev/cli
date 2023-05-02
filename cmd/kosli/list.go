package main

import (
	"io"

	"github.com/spf13/cobra"
)

const listDesc = `All Kosli list commands.`

type listOptions struct {
	output     string
	pageNumber int
	pageLimit  int
}

func (o *listOptions) validate(cmd *cobra.Command) error {
	if o.pageNumber <= 0 {
		return ErrorBeforePrintingUsage(cmd, "page number must be a positive integer")
	}
	if o.pageLimit <= 0 {
		return ErrorBeforePrintingUsage(cmd, "page limit must be a positive integer")
	}
	return nil
}

func newListCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   listDesc,
		Long:    listDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newListApprovalsCmd(out),
		newListArtifactsCmd(out),
		newListDeploymentsCmd(out),
		newListEnvironmentsCmd(out),
		newListFlowsCmd(out),
		newListSnapshotsCmd(out),
		newListAuditTrailsCmd(out),
		newListWorkflowsCmd(out),
	)

	return cmd
}
