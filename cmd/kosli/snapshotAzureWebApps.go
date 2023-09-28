package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/azure"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotAzureFunctionsShortDesc = `Report a snapshot of running Azure Web apps in an Azure resource group to Kosli.  `

const snapshotAzureFunctionsLongDesc = snapshotAzureFunctionsShortDesc + `
The reported data includes Azure web app names, container image digests and creation timestamps.` + azureAuthDesc

const snapshotAzureFunctionsExample = `
kosli snapshot azure-webapps yourEnvironmentName \
	--azure-client-id yourAzureClientID \
	--azure-client-secret yourAzureClientSecret \
	--azure-tenant-id yourAzureTenantID \
	--azure-subscription-id yourAzureSubscriptionID \
	--azure-resource-group-name yourAzureResourceGroupName \
	--api-token yourAPIToken \
	--org yourOrgName
`

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
		Hidden:  true,
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

	cmd.Flags().StringVar(&o.azureStaticCredentials.ClientId, "azure-client-id", "", azureClientIdFlag)
	cmd.Flags().StringVar(&o.azureStaticCredentials.ClientSecret, "azure-client-secret", "", azureClientSecretFlag)
	cmd.Flags().StringVar(&o.azureStaticCredentials.TenantId, "azure-tenant-id", "", azureTenantIdFlag)
	cmd.Flags().StringVar(&o.azureStaticCredentials.SubscriptionId, "azure-subscription-id", "", azureSubscriptionIdFlag)
	cmd.Flags().StringVar(&o.azureStaticCredentials.ResourceGroupName, "azure-resource-group-name", "", azureResourceGroupNameFlag)
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
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/azure-web-app", global.Host, global.Org, envName)

	webAppsData, err := o.azureStaticCredentials.GetWebAppsData(logger)
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
