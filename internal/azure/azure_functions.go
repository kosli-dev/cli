package azure

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

type AzureStaticCredentials struct {
	TenantId          string
	ClientId          string
	ClientSecret      string
	SubscriptionId    string
	ResourceGroupName string
}

func (staticCreds *AzureStaticCredentials) GetWebAppsInfo() ([]*armappservice.Site, error) {
	credentials, err := azidentity.NewClientSecretCredential(staticCreds.TenantId, staticCreds.ClientId, staticCreds.ClientSecret, nil)
	if err != nil {
		return nil, err
	}

	// Docs: https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/appservice/armappservice/README.md
	appserviceClientFactory, err := armappservice.NewClientFactory(staticCreds.SubscriptionId, credentials, nil)
	if err != nil {
		return nil, err
	}
	webAppsClient := appserviceClientFactory.NewWebAppsClient()

	ctx := context.Background()
	webappsPager := webAppsClient.NewListByResourceGroupPager(staticCreds.ResourceGroupName, nil)

	var webAppsInfo []*armappservice.Site
	for webappsPager.More() {
		response, err := webappsPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		webAppsInfo = append(webAppsInfo, response.Value...)
	}
	return webAppsInfo, nil
}

func (staticCreds *AzureStaticCredentials) GetDockerLogs(appServiceName string) error {
	credentials, err := azidentity.NewClientSecretCredential(staticCreds.TenantId, staticCreds.ClientId, staticCreds.ClientSecret, nil)
	if err != nil {
		return err
	}

	// Docs: https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/appservice/armappservice/README.md
	appserviceClientFactory, err := armappservice.NewClientFactory(staticCreds.SubscriptionId, credentials, nil)
	if err != nil {
		return err
	}
	webAppsClient := appserviceClientFactory.NewWebAppsClient()

	ctx := context.Background()
	fmt.Println("Getting logs for app service: ", appServiceName)
	// response, err := webAppsClient.GetContainerLogsZip(ctx, staticCreds.ResourceGroupName, appServiceName, nil)
	response, err := webAppsClient.GetWebSiteContainerLogs(ctx, staticCreds.ResourceGroupName, appServiceName, nil)
	if err != nil {
		return err
	}
	fmt.Println("Got logs for app service: ", appServiceName)
	if response.Body != nil {
		defer response.Body.Close()
	}
	fmt.Println("Reading logs for app service: ", appServiceName)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))

	// out, err := os.Create("somelogs.zip")
	// if err != nil {
	// 	return err
	// }
	// defer out.Close()
	// io.Copy(out, response.Body)

	// fmt.Println(len(body))
	// fmt.Println(string(body))
	return nil
}
