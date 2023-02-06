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

const deploymentReportShortDesc = `Report a deployment of an artifact to an environment to Kosli.`

const deploymentReportLongDesc = deploymentReportShortDesc + `
` + sha256Desc

func newDeploymentReportCmd(out io.Writer) *cobra.Command {
	o := new(deploymentReportOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:        "report [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:      deploymentReportShortDesc,
		Long:       deploymentReportLongDesc,
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
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *deploymentReportOptions) run(args []string) error {
	var err error
	if o.payload.Sha256 == "" {
		o.payload.Sha256, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/deployments/", global.Host, global.Owner, o.pipelineName)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("deployment of artifact %s was reported to: %s", o.payload.Sha256, o.payload.Environment)
	}
	return err
}
