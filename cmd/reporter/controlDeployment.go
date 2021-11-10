package main

import (
	"fmt"
	"io"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type controlDeploymentOptions struct {
	artifactType string
	sha256       string
	pipelineName string
}

func newControlDeploymentCmd(out io.Writer) *cobra.Command {
	o := new(controlDeploymentOptions)
	cmd := &cobra.Command{
		Use:   "control ARTIFACT-NAME-OR-PATH",
		Short: "Check if an artifact in Merkely has been approved for deployment.",
		Long:  controlDeploymentDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("docker image name or file/dir path is required")
			}

			if o.artifactType == "" && o.sha256 == "" {
				return fmt.Errorf("either --type or --sha256 must be specified")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			if o.sha256 != "" {
				if err := digest.ValidateDigest(o.sha256); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if o.sha256 == "" {
				o.sha256, err = GetSha256Digest(o.artifactType, args[0])
				if err != nil {
					return err
				}
			}

			url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/approvals/", global.Host, global.Owner, o.pipelineName, o.sha256)

			response, err := requests.DoRequest([]byte{}, url, global.ApiToken,
				global.MaxAPIRetries, "GET", log)
			if err != nil {
				return err
			}
			fmt.Println(response.Body)

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact to be approved. Options are [dir, file, docker]. Only required if you don't specify --sha256.")
	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")

	err := RequireFlags(cmd, []string{"pipeline"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func controlDeploymentDesc() string {
	return `Check if an artifact in Merkely has been approved for deployment.
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   `
}
