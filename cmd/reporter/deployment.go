package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type deploymentOptions struct {
	artifactType string
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
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return ValidateArtifactArg(args, o.artifactType, o.payload.Sha256)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact. Options are [dir, file, docker].")
	cmd.Flags().StringVarP(&o.payload.Sha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.payload.Environment, "environment", "e", "", "The environment name.")
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", "[optional] The artifact description.")
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), "The url of CI pipeline that built the artifact.")
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", "[optional] The path to a JSON file containing additional data you would like to attach to this deployment.")

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "environment"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *deploymentOptions) run(args []string) error {
	var err error
	if o.payload.Sha256 == "" {
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

	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPost, log)
	return err
}

func deploymentDesc() string {
	return `
   Report a deployment of an artifact to an environment in Merkely. 
   The artifact SHA256 fingerprint is calculated and reported 
   or,alternatively, can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"build-url"})
}
