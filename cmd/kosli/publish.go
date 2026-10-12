package main

import (
	"io"

	"github.com/spf13/cobra"
)

const publishDesc = `All Kosli publish commands.`

func newPublishCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: publishDesc,
		Long:  publishDesc,
	}

	cmd.AddCommand(
		newPublishRegoPolicyCmd(out),
	)
	return cmd
}
