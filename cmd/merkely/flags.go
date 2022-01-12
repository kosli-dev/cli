package main

import "github.com/spf13/cobra"

type fingerprintOptions struct {
	artifactType     string
	registryProvider string
	registryUsername string
	registryPassword string
}

func addFingerprintFlags(cmd *cobra.Command, o *fingerprintOptions) {
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact to calculate its SHA256 fingerprint.")
	cmd.Flags().StringVar(&o.registryProvider, "registry-provider", "", "The docker registry provider. Allowed options are [dockerhub, github].")
	cmd.Flags().StringVar(&o.registryUsername, "registry-username", "", "The docker registry username.")
	cmd.Flags().StringVar(&o.registryPassword, "registry-password", "", "The docker registry password or access token.")
}
