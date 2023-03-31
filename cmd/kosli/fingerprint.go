package main

import (
	"io"

	"github.com/spf13/cobra"
)

const fingerprintShortDesc = `Calculate the SHA256 fingerprint of an artifact.`

const fingerprintLongDesc = fingerprintShortDesc + `
Requires artifact type flag to be set.
Artifact type can be one of: "file" for files, "dir" for directories, "docker" for docker images.`

type fingerprintOptions struct {
	artifactType     string
	registryProvider string
	registryUsername string
	registryPassword string
	excludePaths     []string
}

func newFingerprintCmd(out io.Writer) *cobra.Command {
	o := new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "fingerprint {IMAGE-NAME | FILE-PATH | DIR-PATH}",
		Short: fingerprintShortDesc,
		Long:  fingerprintLongDesc,
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateRegistryFlags(cmd, o)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, out)
		},
	}

	addFingerprintFlags(cmd, o)
	cmd.Flags().StringSliceVarP(&o.excludePaths, "exclude", "e", []string{}, excludePathsFlag)
	err := RequireFlags(cmd, []string{"artifact-type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}
	return cmd
}

func (o *fingerprintOptions) run(args []string, out io.Writer) error {
	fingerprint, err := GetSha256Digest(args[0], o, logger)
	if err != nil {
		return err
	}
	logger.Info(fingerprint)
	return nil
}
