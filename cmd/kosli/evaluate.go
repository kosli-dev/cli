package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateDesc = `All Kosli evaluate commands.`

func newEvaluateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evaluate",
		Short: evaluateDesc,
		Long:  evaluateDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEvaluateTrailCmd(out),
		newEvaluateTrailsCmd(out),
	)

	return cmd
}
