package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type deploymentOptions struct {
	artifactType string
	inputSha256  string
	pipelineName string
	userDataFile string
	payload      DeploymentPayload
}

type DeploymentPayload struct {
	Sha256      string                 `json:"artifact_sha256"`
	Description string                 `json:"description"`
	Environment string                 `json:"environment"`
	UserData    map[string]interface{} `json:"user_data"`
	BuildUrl    string                 `json:"build_url"`
}

func newDeploymentCmd(out io.Writer) *cobra.Command {
	o := new(deploymentOptions)
	cmd := &cobra.Command{
		Use:   "deployment ARTIFACT-NAME-OR-PATH",
		Short: "Report/Log a deployment to Merkely. ",
		Long:  deploymentDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("docker image name or file/dir path is required")
			}

			if o.artifactType == "" && o.inputSha256 == "" {
				return fmt.Errorf("either --type or --sha256 must be specified")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			if o.inputSha256 != "" {
				if err := digest.ValidateDigest(o.inputSha256); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if o.inputSha256 != "" {
				o.payload.Sha256 = o.inputSha256
			} else {
				o.payload.Sha256, err = GetSha256Digest(o.artifactType, args[0])
				if err != nil {
					return err
				}
			}

			o.payload.UserData, err = LoadUserData(o.userDataFile)
			if err != nil {
				return err
			}

			url := fmt.Sprintf("%s/api/v1/projects/%s/%s/deployments/", global.Host, global.Owner, o.pipelineName)

			js, _ := json.MarshalIndent(o.payload, "", "    ")

			return requests.SendPayload(js, url, global.ApiToken,
				global.MaxAPIRetries, global.DryRun, "POST", log)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.artifactType, "type", "t", "", "The type of the artifact. Options are [dir, file, docker].")
	cmd.Flags().StringVarP(&o.inputSha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.payload.Environment, "environment", "e", "", "The environment name.")
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", "[optional] The artifact description.")
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), "The url of CI pipeline that built the artifact.")
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", "The path to a JSON file containing additional data you would like to attach to this deployment.")

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "environment"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func deploymentDesc() string {
	return `
   Report a deployment of an artifact to an environment in Merkely. 
   The artifact SHA256 fingerprint is calculated and reported 
   or,alternatively, can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"build-url"})
}
