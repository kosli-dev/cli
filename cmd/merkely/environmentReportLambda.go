package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportLambdaDesc = `
Report the artifact deployed in an AWS Lambda and its digest to Kosli. 
`

const environmentReportLambdaExample = `
# report what is running in the latest version AWS Lambda function (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report lambda myEnvironment \
	--function-name yourFunctionName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a specific version of an AWS Lambda function (AWS auth provided in flags):
kosli environment report lambda myEnvironment \
	--function-name yourFunctionName \
	--function-version yourFunctionVersion \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--owner yourOrgName
`

type environmentReportLambdaOptions struct {
	functionName    string
	functionVersion string
	accessKey       string
	secretKey       string
	region          string
}

func newEnvironmentReportLambdaCmd(out io.Writer) *cobra.Command {
	o := new(environmentReportLambdaOptions)
	cmd := &cobra.Command{
		Use:     "lambda env-name",
		Short:   "Report artifact from AWS Lambda to Kosli.",
		Long:    environmentReportLambdaDesc,
		Example: environmentReportLambdaExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return ErrorAfterPrintingHelp(cmd, "only env-name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return ErrorAfterPrintingHelp(cmd, "env-name argument is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.functionName, "function-name", "", functionNameFlag)
	cmd.Flags().StringVar(&o.functionVersion, "function-version", "", functionVersionFlag)
	cmd.Flags().StringVar(&o.accessKey, "aws-key-id", "", awsKeyIdFlag)
	cmd.Flags().StringVar(&o.secretKey, "aws-secret-key", "", awsSecretKeyFlag)
	cmd.Flags().StringVar(&o.region, "aws-region", "", awsRegionFlag)

	err := RequireFlags(cmd, []string{"function-name"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *environmentReportLambdaOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
	creds := aws.AWSCredentials(o.accessKey, o.secretKey)
	lambdaData, err := aws.GetLambdaPackageData(o.functionName, o.functionVersion, creds, o.region)
	if err != nil {
		return err
	}

	requestBody := &aws.LambdaEnvRequest{
		Artifacts: lambdaData,
		Type:      "lambda",
		Id:        envName,
	}

	_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}
