package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportDesc = `
Report compliance events back to Merkely.
`

func newReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "report",
		Short:   "Report compliance events to Merkely.",
		Long:    reportDesc,
		Aliases: []string{"log"},
		//SuggestFor: []string{"reportenv", "env report", "envreport"},
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvCmd(out),
		newArtifactCmd(out),
	)

	return cmd
}
