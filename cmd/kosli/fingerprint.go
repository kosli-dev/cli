package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const fingerprintDesc = `
Print the SHA256 fingerprint of an artifact. Requires artifact type flag to be set.
Artifact type can be one of: "file" for files, "dir" for directories, "docker" for docker images.
`

func newFingerprintCmd(out io.Writer) *cobra.Command {
	o := new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "fingerprint ARTIFACT-NAME-OR-PATH",
		Short: "Print the SHA256 fingerprint of an artifact.",
		Long:  fingerprintDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return ErrorBeforePrintingUsage(cmd, "only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return ErrorBeforePrintingUsage(cmd, "docker image name or file/dir path is required")
			}

			return ValidateRegistryFlags(cmd, o)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, out)
		},
	}

	addFingerprintFlags(cmd, o)
	err := RequireFlags(cmd, []string{"artifact-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}
	return cmd
}

func (o *fingerprintOptions) run(args []string, out io.Writer) error {
	fingerprint, err := GetSha256Digest(args[0], o)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, fingerprint)
	return nil
}
