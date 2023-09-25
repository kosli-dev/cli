package azure

import (
	"bufio"
	"bytes"
	"context"
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

type AzureClient struct {
	Credentials       AzureStaticCredentials
	AppServiceFactory *armappservice.ClientFactory
}

func (staticCreds *AzureStaticCredentials) NewAzureClient() (*AzureClient, error) {
	credentials, err := azidentity.NewClientSecretCredential(staticCreds.TenantId, staticCreds.ClientId, staticCreds.ClientSecret, nil)
	if err != nil {
		return nil, err
	}

	// Docs: https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/appservice/armappservice/README.md
	appserviceFactory, err := armappservice.NewClientFactory(staticCreds.SubscriptionId, credentials, nil)
	if err != nil {
		return nil, err
	}

	return &AzureClient{
		Credentials:       *staticCreds,
		AppServiceFactory: appserviceFactory,
	}, nil
}

func (azureClient *AzureClient) GetWebAppsInfo() ([]*armappservice.Site, error) {
	webAppsClient := azureClient.AppServiceFactory.NewWebAppsClient()

	ctx := context.Background()
	webappsPager := webAppsClient.NewListByResourceGroupPager(azureClient.Credentials.ResourceGroupName, nil)

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

func (azureClient *AzureClient) GetDockerLogsForWebApp(appServiceName string) (logs []byte, error error) {
	webAppsClient := azureClient.AppServiceFactory.NewWebAppsClient()

	ctx := context.Background()

	response, err := webAppsClient.GetWebSiteContainerLogs(ctx, azureClient.Credentials.ResourceGroupName, appServiceName, nil)
	if err != nil {
		return nil, err
	}

	if response.Body != nil {
		defer response.Body.Close()
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func ExractImageFingerprintFromLogs(logs []byte) (string, error) {
	logsReader := bytes.NewReader(logs)
	scanner := bufio.NewScanner(logsReader)
	var lastDigestLine []byte
	searchedByteLine := []byte("Digest: sha256:")
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.Contains(line, searchedByteLine) {
			lastDigestLine = line
		}
	}
	if lastDigestLine == nil {
		return "", nil
	}

	lastDigestLineString := string(lastDigestLine)
	startIndex := len(lastDigestLineString) - 64
	extractedDigest := lastDigestLineString[startIndex:]
	return extractedDigest, nil
}
