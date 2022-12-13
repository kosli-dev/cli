package main

import "github.com/spf13/cobra"

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

type awsAuthOptions struct {
	accessKey string
	secretKey string
	region    string
}

func addAWSAuthFlags(cmd *cobra.Command, o *awsAuthOptions) {
	cmd.Flags().StringVar(&o.accessKey, "aws-key-id", "", awsKeyIdFlag)
	cmd.Flags().StringVar(&o.secretKey, "aws-secret-key", "", awsSecretKeyFlag)
	cmd.Flags().StringVar(&o.region, "aws-region", "", awsRegionFlag)
}

func addDryRunFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&global.DryRun, "dry-run", "D", false, dryRunFlag)
}
