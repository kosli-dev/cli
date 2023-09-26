package azure

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"

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

// WebAppData represents the harvested Azure Web App data
type WebAppData struct {
	WebAppName string            `json:"webAppName"`
	Digests    map[string]string `json:"digests"`
	// StartedAt  int64             `json:"creationTimestamp"` TODO: decide where to get this from
}

// AzureWebAppsRequest represents the PUT request body to be sent to Kosli from CLI
type AzureWebAppsRequest struct {
	Artifacts []*WebAppData `json:"artifacts"`
}

func (staticCreds *AzureStaticCredentials) GetWebAppsData() ([]*WebAppData, error) {
	webAppsData := []*WebAppData{}
	azureClient, err := staticCreds.NewAzureClient()
	if err != nil {
		return nil, err
	}
	webAppInfo, err := azureClient.GetWebAppsInfo()
	if err != nil {
		return nil, err
	}
	for _, webapp := range webAppInfo {
		if strings.ToLower(*webapp.Properties.State) != "running" {
			continue
		}
		// get image name from DOCKER|tookyregistry.azurecr.io/tookyregistry/tooky/sha256:cb29a6
		linuxFxVersion := strings.Split(*webapp.Properties.SiteConfig.LinuxFxVersion, "|")
		imageName := linuxFxVersion[1]
		var fingerprint string
		if linuxFxVersion[0] == "DOCKER" {
			logs, err := azureClient.GetDockerLogsForWebApp(*webapp.Name)
			if err != nil {
				return nil, err
			}
			fingerprint, err = exractImageFingerprintFromLogs(logs)
			if err != nil {
				return nil, err
			}
		} else {
			// TODO: get fingerprint for non-docker images
			fingerprint = ""
		}

		data := &WebAppData{*webapp.Name, map[string]string{imageName: fingerprint}}
		webAppsData = append(webAppsData, data)
	}
	return webAppsData, nil
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
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func exractImageFingerprintFromLogs(logs []byte) (string, error) {
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
