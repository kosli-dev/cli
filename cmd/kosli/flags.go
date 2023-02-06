package main

import (
	"github.com/kosli-dev/cli/internal/aws"
	"github.com/spf13/cobra"
)

type fingerprintOptions struct {
	artifactType     string
	registryProvider string
	registryUsername string
	registryPassword string
}

func addFingerprintFlags(cmd *cobra.Command, o *fingerprintOptions) {
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", artifactTypeFlag)
	cmd.Flags().StringVar(&o.registryProvider, "registry-provider", "", registryProviderFlag)
	cmd.Flags().StringVar(&o.registryUsername, "registry-username", "", registryUsernameFlag)
	cmd.Flags().StringVar(&o.registryPassword, "registry-password", "", registryPasswordFlag)
}

func addAWSAuthFlags(cmd *cobra.Command, o *aws.AWSStaticCreds) {
	cmd.Flags().StringVar(&o.AccessKeyID, "aws-key-id", "", awsKeyIdFlag)
	cmd.Flags().StringVar(&o.SecretAccessKey, "aws-secret-key", "", awsSecretKeyFlag)
	cmd.Flags().StringVar(&o.Region, "aws-region", "", awsRegionFlag)
}

func addDryRunFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&global.DryRun, "dry-run", "D", false, dryRunFlag)
}
