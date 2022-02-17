package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/aws"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportS3Desc = `
Report the artifact deployed in an AWS S3 bucket and their digests 
and reports it to Merkely. 
`

const environmentReportS3Example = `
* report what's running in an AWS S3 bucket:
merkely environment report s3 prod --api-token 1234 --owner exampleOrg
`

type environmentReportS3Options struct {
	bucket    string
	accessKey string
	secretKey string
	region    string
}

func newEnvironmentReportS3Cmd(out io.Writer) *cobra.Command {
	o := new(environmentReportS3Options)
	cmd := &cobra.Command{
		Use:     "s3 env-name",
		Aliases: []string{"S3"},
		Short:   "Report artifact from AWS S3 bucket to Merkely.",
		Long:    environmentReportS3Desc,
		Example: environmentReportS3Example,
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

	cmd.Flags().StringVar(&o.bucket, "bucket", "", "The name of the S3 bucket.")
	cmd.Flags().StringVar(&o.accessKey, "access-key", "", "The AWS access key")
	cmd.Flags().StringVar(&o.secretKey, "secret-key", "", "The AWS secret key")
	cmd.Flags().StringVar(&o.region, "region", "", "The AWS region")

	err := RequireFlags(cmd, []string{"bucket"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *environmentReportS3Options) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
	creds := aws.AWSCredentials(o.accessKey, o.secretKey)
	s3Data, err := aws.GetS3Data(o.bucket, creds, o.region)
	if err != nil {
		return err
	}
	requestBody := &aws.S3EnvRequest{
		Artifacts: s3Data,
		Type:      "S3",
		Id:        envName,
	}

	_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}
