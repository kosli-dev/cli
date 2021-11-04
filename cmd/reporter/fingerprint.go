package main

import (
	"fmt"
	"io"
	"log"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/spf13/cobra"
)

const fingerprintDesc = `
Print the SHA256 fingerprint of an artifact. Requires artifact type flag to be set.
Artifact type can be one of: "file" for files, "dir" for directories, "docker" for docker images.
`

type fingerprintOptions struct {
	artifactType string
}

func newFingerprintCmd(out io.Writer) *cobra.Command {
	o := new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "fingerprint",
		Short: "Print the SHA256 fingerprint of an artifact.",
		Long:  fingerprintDesc,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("docker image name or file/dir path is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var fingerprint string
			switch o.artifactType {
			case "file":
				fingerprint, err = digest.FileSha256(args[0])
			case "dir":
				fingerprint, err = digest.DirSha256(args[0], false)
			case "docker":
				fingerprint, err = digest.DockerImageSha256(args[0])
			default:
				return fmt.Errorf("%s is not a supported artifact type", o.artifactType)
			}

			if err != nil {
				return err
			}
			fmt.Print(fingerprint)

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.artifactType, "type", "t", "", "the type of the artifact to calculate its SHA256 fingerprint")
	err := RequireFlags(cmd, []string{"type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}
	return cmd
}
