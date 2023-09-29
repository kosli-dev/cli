package main

import (
	"io"

	"github.com/spf13/cobra"
)

const snapshotDesc = `All Kosli snapshot commands.`

func newSnapshotCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: snapshotDesc,
		Long:  snapshotDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newSnapshotDockerCmd(out),
		newSnapshotECSCmd(out),
		newSnapshotK8SCmd(out),
		newSnapshotServerCmd(out),
		newSnapshotLambdaCmd(out),
		newSnapshotS3Cmd(out),
		newSnapshotAzureAppsCmd(out),
	)

	return cmd
}
