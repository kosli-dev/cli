package main

import (
	"io"

	"github.com/spf13/cobra"
)

const getDesc = `All Kosli get commands.`

func newGetCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   getDesc,
		Long:    getDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newGetApiKeyCmd(out),
		newGetApprovalCmd(out),
		newGetArtifactCmd(out),
		newGetEnvironmentCmd(out),
		newGetFlowCmd(out),
		newGetSnapshotCmd(out),
		newGetTrailCmd(out),
		newGetPolicyCmd(out),
		newGetAttestationTypeCmd(out),
		newGetControlCmd(out),
		newGetAttestationCmd(out),
		newGetRepoCmd(out),
		newGetServiceAccountCmd(out),
		newGetDefaultOrgCmd(out),
	)
	return cmd
}
