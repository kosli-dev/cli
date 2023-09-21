package main

import (
	"fmt"
	"io"

	"github.com/kosli-dev/cli/internal/azure"
	"github.com/spf13/cobra"
)

const snapshotAzureFunctionsShortDesc = ``

const snapshotAzureFunctionsLongDesc = snapshotAzureFunctionsShortDesc + ``

const snapshotAzureFunctionsExample = ``

type snapshotAzureFunctionsOptions struct {
	functionNames    []string
	functionVersion  string
	azureCredentials *azure.AzureStaticCredentials
}

func newSnapshotAzureFunctionsCmd(out io.Writer) *cobra.Command {
	o := new(snapshotAzureFunctionsOptions)
	o.azureCredentials = new(azure.AzureStaticCredentials)
	cmd := &cobra.Command{
		Use:     "azure-apps ENVIRONMENT-NAME",
		Short:   snapshotAzureFunctionsShortDesc,
		Long:    snapshotAzureFunctionsLongDesc,
		Example: snapshotAzureFunctionsExample,
		// Args:    cobra.ExactArgs(1),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			// if err != nil {
			// 	return ErrorBeforePrintingUsage(cmd, err.Error())
			// }

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.azureCredentials.ClientId, "azure-client-id", "", "")
	cmd.Flags().StringVar(&o.azureCredentials.ClientSecret, "azure-client-secret", "", "")
	cmd.Flags().StringVar(&o.azureCredentials.TenantId, "azure-tenant-id", "", "")
	cmd.Flags().StringVar(&o.azureCredentials.SubscriptionId, "azure-subscription-id", "", "")
	cmd.Flags().StringVar(&o.azureCredentials.ResourceGroupName, "azure-resource-group-name", "", "")

	addDryRunFlag(cmd)

	return cmd
}

func (o *snapshotAzureFunctionsOptions) run(args []string) error {
	webAppInfo, err := o.azureCredentials.GetWebAppsInfo()
	if err != nil {
		return err
	}
	for _, webapp := range webAppInfo {
		fmt.Printf("Webapp name: %s\n", *webapp.Name)
		fmt.Printf("Webapp image: %s\n", *webapp.Properties.SiteConfig.LinuxFxVersion)
		for _, host := range webapp.Properties.EnabledHostNames {
			fmt.Printf("Webapp host: %s\n", *host)
		}

	}

	// envName := args[0]

	// TODO: Change later for azure function environments
	// payload := &aws.LambdaEnvRequest{
	// 	Artifacts: lambdaData,
	// }
	// url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/lambda", global.Host, global.Org, envName)
	// reqParams := &requests.RequestParams{
	// 	Method:   http.MethodPut,
	// 	URL:      url,
	// 	Payload:  payload,
	// 	DryRun:   global.DryRun,
	// 	Password: global.ApiToken,
	// }
	// _, err = kosliClient.Do(reqParams)
	// if err == nil && !global.DryRun {
	// 	logger.Info("%d lambda functions were reported to environment %s", len(lambdaData), envName)
	// }
	// return err
	return nil
}
