package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type expectDeploymentOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	userDataFile       string
	payload            ExpectDeploymentPayload
}

type ExpectDeploymentPayload struct {
	Fingerprint string      `json:"artifact_sha256"`
	Description string      `json:"description"`
	Environment string      `json:"environment"`
	UserData    interface{} `json:"user_data"`
	BuildUrl    string      `json:"build_url"`
}

func newExpectDeploymentCmd(out io.Writer) *cobra.Command {
	o := new(expectDeploymentOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "deployment [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short: deploymentReportShortDesc,
		Long:  deploymentReportLongDesc,
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Fingerprint, false)
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
	cmd.Flags().StringVarP(&o.payload.Fingerprint, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.Environment, "environment", "e", "", environmentNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", artifactDescriptionFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", deploymentUserDataFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "build-url", "environment"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *expectDeploymentOptions) run(args []string) error {
	var err error
	if o.payload.Fingerprint == "" {
		o.payload.Fingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/deployments/", global.Host, global.Owner, o.flowName)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("deployment of artifact %s was reported to: %s", o.payload.Fingerprint, o.payload.Environment)
	}
	return err
}
