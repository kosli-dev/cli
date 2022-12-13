package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportLambdaShortDesc = `Report the artifact deployed in an AWS Lambda and its digest to Kosli.`

const environmentReportLambdaLongDesc = environmentReportLambdaShortDesc + `
To authenticate to AWS, you can either export the AWS env vars or use the command flags to pass them.
See the examples below.s`

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
	awsAuthOptions  *awsAuthOptions
}

func newEnvironmentReportLambdaCmd(out io.Writer) *cobra.Command {
	o := new(environmentReportLambdaOptions)
	o.awsAuthOptions = new(awsAuthOptions)
	cmd := &cobra.Command{
		Use:     "lambda ENVIRONMENT-NAME",
		Short:   environmentReportLambdaShortDesc,
		Long:    environmentReportLambdaLongDesc,
		Example: environmentReportLambdaExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.functionName, "function-name", "", functionNameFlag)
	cmd.Flags().StringVar(&o.functionVersion, "function-version", "", functionVersionFlag)
	addAWSAuthFlags(cmd, o.awsAuthOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"function-name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *environmentReportLambdaOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
	creds := aws.AWSCredentials(o.awsAuthOptions.accessKey, o.awsAuthOptions.secretKey)
	lambdaData, err := aws.GetLambdaPackageData(o.functionName, o.functionVersion, creds, o.awsAuthOptions.region)
	if err != nil {
		return err
	}

	payload := &aws.LambdaEnvRequest{
		Artifacts: lambdaData,
		Type:      "lambda",
		Id:        envName,
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("%s lambda function was reported to environment %s", o.functionName, envName)
	}
	return err
}
