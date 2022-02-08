package main

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type allowedArtifactsCreationOptions struct {
	fingerprintOptions *fingerprintOptions
	payload            AllowlistPayload
}

type AllowlistPayload struct {
	Sha256      string `json:"sha256"`
	Filename    string `json:"artifact_name"`
	Reason      string `json:"description"`
	Environment string `json:"environment_name"`
}

func newAllowedArtifactsCreateCmd(out io.Writer) *cobra.Command {
	o := new(allowedArtifactsCreationOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "add ARTIFACT-NAME-OR-PATH",
		Short: "Add an artifact to an environment's allowlist. ",
		Long:  allowedArtifactsCreationDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Sha256, true)
			if err != nil {
				return err
			}
			return ValidateRegisteryFlags(o.fingerprintOptions)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.payload.Sha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.")
	cmd.Flags().StringVarP(&o.payload.Environment, "environment", "e", "", "The environment name for which the artifact is allowlisted.")
	cmd.Flags().StringVar(&o.payload.Reason, "reason", "", "The reason why this artifact is allowlisted.")

	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"environment", "reason"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *allowedArtifactsCreationOptions) run(args []string) error {
	if o.payload.Sha256 != "" {
		o.payload.Filename = args[0]
	} else {
		var err error
		o.payload.Sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
		if o.fingerprintOptions.artifactType == "dir" || o.fingerprintOptions.artifactType == "file" {
			o.payload.Filename = filepath.Base(args[0])
		} else {
			o.payload.Filename = args[0]
		}
	}

	url := fmt.Sprintf("%s/api/v1/policies/%s/allowedartifacts/", global.Host, global.Owner)

	_, err := requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}

func allowedArtifactsCreationDesc() string {
	return `
   Add an artifact to an environment's allowlist. 
   The artifact SHA256 fingerprint is calculated and reported 
   or, alternatively, can be provided directly. 
   `
}
