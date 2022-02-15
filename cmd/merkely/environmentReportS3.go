package main

import (
	"fmt"
	"io"

	"github.com/merkely-development/reporter/internal/aws"
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
		Short:   "Report artifact from AWS S3 bucket to Merkely.",
		Long:    environmentReportS3Desc,
		Example: environmentReportS3Example,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only environment name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("environment name is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.bucket, "bucket", "C", "", "The name of the S3 bucket.")
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
	// envName := args[0]

	// url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
	creds := aws.AWSCredentials(o.accessKey, o.secretKey)
	sha256, err := aws.GetS3Digest(o.bucket, creds, o.region)
	if err != nil {
		return err
	}
	fmt.Printf("Sha256: %s\n", sha256)

	// tasksData, err := aws.GetEcsTasksData(client, o.cluster, o.serviceName)
	// if err != nil {
	// 	return err
	// }

	// requestBody := &aws.EcsEnvRequest{
	// 	Artifacts: tasksData,
	// 	Type:      "S3",
	// 	Id:        o.id,
	// }

	// _, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
	// 	global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}
