package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/azure"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotAzureAppsShortDesc = `Report a snapshot of running Azure service apps and function apps in an Azure resource group to Kosli.  `

const snapshotAzureAppsLongDesc = snapshotAzureAppsShortDesc + `
The reported data includes Azure app names, container image digests and creation timestamps.` + azureAuthDesc

const snapshotAzureAppsExample = `
# Use Docker logs of Azure apps to get the digests for artifacts in a snapshot
kosli snapshot azure yourEnvironmentName \
	--azure-client-id yourAzureClientID \
	--azure-client-secret yourAzureClientSecret \
	--azure-tenant-id yourAzureTenantID \
	--azure-subscription-id yourAzureSubscriptionID \
	--azure-resource-group-name yourAzureResourceGroupName \
	--digests-source logs \
	--api-token yourAPIToken \
	--org yourOrgName

# Use Azure Container Registry to get the digests for artifacts in a snapshot
kosli snapshot azure yourEnvironmentName \
	--azure-client-id yourAzureClientID \
	--azure-client-secret yourAzureClientSecret \
	--azure-tenant-id yourAzureTenantID \
	--azure-subscription-id yourAzureSubscriptionID \
	--azure-resource-group-name yourAzureResourceGroupName \
	--digests-source acr \
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotAzureAppsOptions struct {
	azureStaticCredentials *azure.AzureStaticCredentials
}

func newSnapshotAzureAppsCmd(out io.Writer) *cobra.Command {
	o := new(snapshotAzureAppsOptions)
	o.azureStaticCredentials = new(azure.AzureStaticCredentials)
	cmd := &cobra.Command{
		Use:     "azure ENVIRONMENT-NAME",
		Short:   snapshotAzureAppsShortDesc,
		Long:    snapshotAzureAppsLongDesc,
		Example: snapshotAzureAppsExample,
		Hidden:  true,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			if o.azureStaticCredentials.DigestsSource != "acr" && o.azureStaticCredentials.DigestsSource != "logs" {
				return fmt.Errorf("invalid value for --digests-source flag. Valid values are 'acr' and 'logs'")
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
	cmd.Flags().BoolVar(&o.azureStaticCredentials.DownloadLogsAsZip, "zip", false, "Download logs from Azure as zip files")
	cmd.Flags().StringVar(&o.azureStaticCredentials.DigestsSource, "digests-source", "acr", azureDigestsSourceFlag)
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

func (o *snapshotAzureAppsOptions) run(args []string) error {
	envName := args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/azure-apps", global.Host, global.Org, envName)

	webAppsData, err := o.azureStaticCredentials.GetAzureAppsData(logger)
	if err != nil {
		return err
	}
	payload := &azure.AzureAppsRequest{
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
		logger.Info("%d azure apps were reported to environment %s", len(webAppsData), envName)
	}
	return err
}
