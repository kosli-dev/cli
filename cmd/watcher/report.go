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
		Use:   "report [event type] [args] [flags]",
		Short: "report compliance events to Merkely.",
		Long:  reportDesc,
		// RunE: func(cmd *cobra.Command, args []string) error {
		// 	for _, c := range cmd.Commands() {
		// 		log.Println(c)
		// 	}
		// 	return nil
		// },
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvCmd(out),
	)

	return cmd
}
