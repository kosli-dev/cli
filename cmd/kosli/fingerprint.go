package main

import (
	"io"

	"github.com/spf13/cobra"
)

const fingerprintShortDesc = `Calculate the SHA256 fingerprint of an artifact.`

const fingerprintDirSynopsis = `When fingerprinting a 'dir' artifact, you can exclude certain paths from fingerprint calculation 
using the ^--exclude^ flag.
Excluded paths are relative to the DIR-PATH and can be literal paths or
glob patterns.  
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

` + kosliIgnoreDesc

const fingerprintLongDesc = fingerprintShortDesc + `
Requires ^--artifact-type^ flag to be set.
Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.

Fingerprinting container images can be done using the local docker daemon or the fingerprint can be fetched
from a remote registry.

` + fingerprintDirSynopsis

const fingerprintExamples = `
# fingerprint a file
kosli fingerprint --artifact-type file file.txt

# fingerprint a dir
kosli fingerprint --artifact-type dir mydir

# fingerprint a dir while excluding paths
kosli fingerprint --artifact-type dir --exclude logs --exclude *.exe mydir

# fingerprint a locally available docker image (requires docker daemon running)
kosli fingerprint --artifact-type docker nginx:latest

# fingerprint a public image from a remote registry
kosli fingerprint --artifact-type oci nginx:latest

# fingerprint a private image from a remote registry
kosli fingerprint --artifact-type oci private:latest --registry-username YourUsername --registry-password YourPassword
`

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
		Use:     "fingerprint {IMAGE-NAME | FILE-PATH | DIR-PATH}",
		Short:   fingerprintShortDesc,
		Long:    fingerprintLongDesc,
		Example: fingerprintExamples,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateRegistryFlags(cmd, o)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, out)
		},
	}

	addFingerprintFlags(cmd, o)
	cmd.Flags().StringSliceVarP(&o.excludePaths, "e", "e", []string{}, excludePathsFlag)
	err := RequireFlags(cmd, []string{"artifact-type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	err = DeprecateFlags(cmd, map[string]string{
		"e": "use -x instead",
	})

	if err != nil {
		logger.Error("failed to configure deprecated flags: %v", err)
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
