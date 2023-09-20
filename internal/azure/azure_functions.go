package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

type AzureFunctionsCredentials struct {
	TenantId          string
	ClientId          string
	ClientSecret      string
	SubscriptionId    string
	ResourceGroupName string
}

func (staticCreds *AzureFunctionsCredentials) GetWebAppsInfo() ([]byte, error) {
	fmt.Printf("TenantId: %s\n", staticCreds.TenantId)
	fmt.Printf("ClientId: %s\n", staticCreds.ClientId)
	fmt.Printf("ClientSecret: %s\n", staticCreds.ClientSecret)
	cred, err := azidentity.NewClientSecretCredential(staticCreds.TenantId, staticCreds.ClientId, staticCreds.ClientSecret, nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Reached after new client secret credential\n")

	appserviceClientFactory, err := armappservice.NewClientFactory(staticCreds.SubscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Reached after new client factory\n")
	webAppsClient := appserviceClientFactory.NewWebAppsClient()
	fmt.Printf("Reached after new web apps client\n")
	webappsPager := webAppsClient.NewListByResourceGroupPager(staticCreds.ResourceGroupName, nil)
	fmt.Printf("Reached after new list by resource group pager\n")
	var webappsArray []byte
	ctx := context.Background()
	for webappsPager.More() {
		fmt.Printf("Inside for loop\n")
		var currentPageData []byte
		webappsPager.UnmarshalJSON(currentPageData)
		fmt.Print(currentPageData)
		webappsArray = append(webappsArray, currentPageData...)
		webappsPager.NextPage(ctx)
	}
	fmt.Printf("Reached after unmarshal json\n")
	return webappsArray, nil
}
