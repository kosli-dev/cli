package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type deploymentReportOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	userDataFile       string
	payload            DeploymentPayload
}

type DeploymentPayload struct {
	Sha256      string      `json:"artifact_sha256"`
	Description string      `json:"description"`
	Environment string      `json:"environment"`
	UserData    interface{} `json:"user_data"`
	BuildUrl    string      `json:"build_url"`
}

func newDeploymentReportCmd(out io.Writer) *cobra.Command {
	o := new(deploymentReportOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:        "report [ARTIFACT-NAME-OR-PATH]",
		Short:      "Report a deployment to Kosli. ",
		Long:       deploymentReportDesc(),
		Deprecated: "use \"kosli expect deployment\" instead.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Sha256, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.payload.Sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.Environment, "environment", "e", "", environmentNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", artifactDescriptionFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", deploymentUserDataFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "environment"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *deploymentReportOptions) run(args []string) error {
	var err error
	if o.payload.Sha256 == "" {
		o.payload.Sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
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

func deploymentReportDesc() string {
	return `
   Report a deployment of an artifact to an environment in Kosli. 
   The artifact SHA256 fingerprint is calculated and reported 
   or,alternatively, can be provided directly. 
   `
}
