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

type fingerprintOptions struct {
	artifactType     string
	registryProvider string
	registryUsername string
	registryPassword string
}

func newFingerprintCmd(out io.Writer) *cobra.Command {
	o := new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "fingerprint",
		Short: "Print the SHA256 fingerprint of an artifact.",
		Long:  fingerprintDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("docker image name or file/dir path is required")
			}
			if (o.registryProvider != "" && o.registryPassword == "") ||
				(o.registryProvider == "" && o.registryPassword != "") {
				return fmt.Errorf("both --registry-provider, --registry-username and registry-password are required if you want to get the digest from a remote registry")

			}

			if o.registryProvider != "dockerhub" && o.registryProvider != "github" && o.registryProvider != "" {
				return fmt.Errorf("%s is not a supported registry for getting the docker image digest remotely", o.registryProvider)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, out)
		},
	}

	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact to calculate its SHA256 fingerprint.")
	cmd.Flags().StringVar(&o.registryProvider, "registry-provider", "", "The docker registry provider. Allowed options are [dockerhub, github].")
	cmd.Flags().StringVar(&o.registryUsername, "registry-username", "", "The docker registry username.")
	cmd.Flags().StringVar(&o.registryPassword, "registry-password", "", "The docker registry password or access token.")
	err := RequireFlags(cmd, []string{"artifact-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}
	return cmd
}

func (o *fingerprintOptions) run(args []string, out io.Writer) error {
	fingerprint, err := GetSha256Digest(o.artifactType, args[0], o.registryProvider, o.registryUsername, o.registryPassword)
	if err != nil {
		return err
	}
	fmt.Fprint(out, fingerprint)
	return nil
}
