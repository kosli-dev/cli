package main

import (
	"io"

	"github.com/spf13/cobra"
)

const assertStatusDesc = `
Assert the status of Merkely server. Exits with non-zero code if Merkely server down.
`

func newAssertStatusCmd(out io.Writer) *cobra.Command {
	o := &statusOptions{assert: true}
	cmd := &cobra.Command{
		Use:   "status",
		Short: assertStatusDesc,
		Long:  assertStatusDesc,
		Args:  NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}
	return cmd
}
