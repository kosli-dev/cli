package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/aws"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportLambdaDesc = `
Report the artifact deployed in an AWS Lambda and its digest to Merkely. 
`

const environmentReportLambdaExample = `
* report what's running in the latest version AWS Lambda function:
merkely environment report lambda myEnvironment --function-name lambda-test --api-token 1234 --owner exampleOrg

* report what's running in a specific version of an AWS Lambda function:
merkely environment report lambda myEnvironment --function-name lambda-test --version 1 --api-token 1234 --owner exampleOrg
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
		Short:   "Report artifact from AWS Lambda to Merkely.",
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

	cmd.Flags().StringVar(&o.functionName, "function-name", "", "The name of the AWS Lambda function.")
	cmd.Flags().StringVar(&o.functionVersion, "version", "", "[optional] The version of the AWS Lambda function.")
	cmd.Flags().StringVar(&o.accessKey, "access-key", "", "The AWS access key")
	cmd.Flags().StringVar(&o.secretKey, "secret-key", "", "The AWS secret key")
	cmd.Flags().StringVar(&o.region, "region", "", "The AWS region")

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
