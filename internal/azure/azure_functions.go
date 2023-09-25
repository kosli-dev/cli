package azure

import (
	"context"
	"io"
	"net/http"

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

func (staticCreds *AzureStaticCredentials) GetDockerLogs() error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://tsha256.scm.azurewebsites.net/api/vfs/LogFiles/2023_09_14_10-30-0-8_docker.log", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(staticCreds.ClientId, staticCreds.ClientSecret)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	println(string(body))

	return nil
}
