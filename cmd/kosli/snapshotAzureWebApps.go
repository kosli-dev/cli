package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/azure"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotAzureFunctionsShortDesc = ``

const snapshotAzureFunctionsLongDesc = snapshotAzureFunctionsShortDesc + ``

const snapshotAzureFunctionsExample = ``

type snapshotAzureFunctionsOptions struct {
	azureStaticCredentials *azure.AzureStaticCredentials
}

func newSnapshotAzureWebAppsCmd(out io.Writer) *cobra.Command {
	o := new(snapshotAzureFunctionsOptions)
	o.azureStaticCredentials = new(azure.AzureStaticCredentials)
	cmd := &cobra.Command{
		Use:     "azure-webapps ENVIRONMENT-NAME",
		Short:   snapshotAzureFunctionsShortDesc,
		Long:    snapshotAzureFunctionsLongDesc,
		Example: snapshotAzureFunctionsExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.azureStaticCredentials.ClientId, "azure-client-id", "", "")
	cmd.Flags().StringVar(&o.azureStaticCredentials.ClientSecret, "azure-client-secret", "", "")
	cmd.Flags().StringVar(&o.azureStaticCredentials.TenantId, "azure-tenant-id", "", "")
	cmd.Flags().StringVar(&o.azureStaticCredentials.SubscriptionId, "azure-subscription-id", "", "")
	cmd.Flags().StringVar(&o.azureStaticCredentials.ResourceGroupName, "azure-resource-group-name", "", "")
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"azure-client-id", "azure-client-secret",
		"azure-tenant-id", "azure-subscription-id", "azure-resource-group-name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotAzureFunctionsOptions) run(args []string) error {
	envName := args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/azure-webapps", global.Host, global.Org, envName)

	webAppsData, err := o.azureStaticCredentials.GetWebAppsData()
	if err != nil {
		return err
	}
	payload := &azure.AzureWebAppsRequest{
		Artifacts: webAppsData,
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
		logger.Info("%d azure web apps were reported to environment %s", len(webAppsData), envName)
	}
	return err
}
