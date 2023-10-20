package azure

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/aws/smithy-go/time"
	"github.com/kosli-dev/cli/internal/logger"
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

// AppData represents the harvested Azure service app and function app data
type AppData struct {
	AppName   string            `json:"appName"`
	AppKind   string            `json:"appKind"`
	Digests   map[string]string `json:"digests"`
	StartedAt int64             `json:"creationTimestamp"`
}

// AzureAppsRequest represents the PUT request body to be sent to Kosli from CLI
type AzureAppsRequest struct {
	Artifacts []*AppData `json:"artifacts"`
}

func (staticCreds *AzureStaticCredentials) GetAzureAppsData(logger *logger.Logger) (appsData []*AppData, err error) {
	azureClient, err := staticCreds.NewAzureClient()
	if err != nil {
		return nil, err
	}

	appsInfo, err := azureClient.GetAppsListForResourceGroup()
	if err != nil {
		return nil, err
	}

	logger.Debug("found %d apps in the resource group %s", len(appsInfo), staticCreds.ResourceGroupName)
	if logger.DebugEnabled {
		logger.Debug("Found apps:")
		for _, app := range appsInfo {
			logger.Debug("  app name=%s, state=%s, kind=%s, linuxFxVersion=%s", *app.Name,
				*app.Properties.State, *app.Kind, *app.Properties.SiteConfig.LinuxFxVersion)
		}
	}

	// run concurrently
	var wg sync.WaitGroup
	errs := make(chan error, 1) // Buffered only for the first error
	appsChan := make(chan *AppData, len(appsInfo))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure it's called to release resources even if no errors

	for _, app := range appsInfo {
		wg.Add(1)
		go func(app *armappservice.Site) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is a must to avoid blocking
			}

			if strings.ToLower(*app.Properties.State) != "running" {
				logger.Debug("app %s is not running, skipping from report", *app.Name)
				return
			}

			data, err := azureClient.NewAppData(app, logger)
			if err != nil {
				select {
				case errs <- err:
				default:
				}
				cancel() // send cancel signal to goroutines
				return
			}
			if !data.IsEmpty() {
				appsChan <- &data
			}
		}(app)
	}

	wg.Wait()
	close(appsChan)

	// Return (first) error, if any:
	if ctx.Err() != nil {
		return appsData, <-errs
	}

	for app := range appsChan {
		appsData = append(appsData, app)
	}

	if appsData == nil {
		appsData = make([]*AppData, 0)
	}
	return appsData, nil
}

func (azureClient *AzureClient) NewAppData(app *armappservice.Site, logger *logger.Logger) (AppData, error) {
	// Construct and return AppData for the provided armappservice.Site

	// get image name from "DOCKER|tookyregistry.azurecr.io/tookyregistry/tooky/sha256:cb29a6"
	linuxFxVersion := strings.Split(*app.Properties.SiteConfig.LinuxFxVersion, "|")
	if len(linuxFxVersion) != 2 || linuxFxVersion[0] != "DOCKER" {
		logger.Debug("app %s is not using a Docker image, skipping from report", *app.Name)
		//  TODO: support other types of images, for now just skip
		return AppData{}, nil
	}
	imageName := linuxFxVersion[1]

	var fingerprint string
	var startedAt int64

	// First try to get image fingerprint from Docker logs.
	// If it is not found, then try to get it from image name.
	// We also need Docker logs to get startedAt timestamp.
	logs, err := azureClient.GetDockerLogsForApp(*app.Name)
	if err != nil {
		return AppData{}, err
	}

	fingerprint, startedAt, err = exractImageFingerprintAndStartedTimestampFromLogs(logs, *app.Name)
	if err != nil {
		return AppData{}, err
	}
	if startedAt == 0 {
		// if startedAt is 0, then the container is not running
		logger.Debug("docker container in app %s is not running, skipping from report", *app.Name)
		return AppData{}, nil
	}

	if fingerprint == "" && strings.Contains(imageName, "@sha256:") {
		// get digest from image if it is pulled by sha256 digest, ie, imageName@sha256:cb29a6edff54216aa3e1d433aa98f0d1a711d17e59004fb6e3afffe0a784e34e"
		fingerprint = strings.Split(imageName, "@sha256:")[1]
	}

	logger.Debug("For app %s found: image=%s, fingerprint=%s, startedAt=%d", *app.Name, imageName, fingerprint, startedAt)

	return AppData{*app.Name, *app.Kind, map[string]string{imageName: fingerprint}, startedAt}, nil
}

func (app *AppData) IsEmpty() bool {
	return app.AppName == "" && len(app.Digests) == 0 && app.StartedAt == 0
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

func (azureClient *AzureClient) GetAppsListForResourceGroup() ([]*armappservice.Site, error) {
	webAppsClient := azureClient.AppServiceFactory.NewWebAppsClient()

	ctx := context.Background()
	appsPager := webAppsClient.NewListByResourceGroupPager(azureClient.Credentials.ResourceGroupName, nil)

	var appsInfo []*armappservice.Site
	for appsPager.More() {
		response, err := appsPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		appsInfo = append(appsInfo, response.Value...)
	}
	return appsInfo, nil
}

func (azureClient *AzureClient) GetDockerLogsForApp(appServiceName string) (logs []byte, error error) {
	appsClient := azureClient.AppServiceFactory.NewWebAppsClient()

	ctx := context.Background()

	response, err := appsClient.GetWebSiteContainerLogs(ctx, azureClient.Credentials.ResourceGroupName, appServiceName, nil)
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

func exractImageFingerprintAndStartedTimestampFromLogs(logs []byte, appName string) (fingerprint string, startedAt int64, error error) {
	logsReader := bytes.NewReader(logs)
	scanner := bufio.NewScanner(logsReader)

	searchedDigestByteArray := []byte("Digest: sha256:")
	containerStartedAtByteArray := []byte(fmt.Sprintf("for site %s initialized successfully and is ready to serve requests.", appName))

	var lastDigestLine []byte
	var lastStartedAtLine []byte
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.Contains(line, searchedDigestByteArray) {
			lastDigestLine = make([]byte, len(line))
			copy(lastDigestLine, line)
		}

		if bytes.Contains(line, containerStartedAtByteArray) {
			lastStartedAtLine = make([]byte, len(line))
			copy(lastStartedAtLine, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", 0, err
	}

	lengthOfTimestamp := 24 // example 2023-09-25T12:21:09.927Z
	var digestLoggedAt string
	if lastDigestLine != nil {
		lastDigestLineString := string(lastDigestLine)
		fingerprintStartIndex := len(lastDigestLineString) - 64
		fingerprint = lastDigestLineString[fingerprintStartIndex:]
		digestLoggedAt = lastDigestLineString[:lengthOfTimestamp]
	}

	var startedAtLoggedAt string
	if lastStartedAtLine != nil {
		startedAtLoggedAt = string(lastStartedAtLine)[:lengthOfTimestamp]
	}

	if digestLoggedAt != "" && startedAtLoggedAt != "" {
		digestLoggedAt = strings.TrimSpace(digestLoggedAt)
		digestLogTime, err := time.ParseDateTime(digestLoggedAt)
		if err != nil {
			return "", 0, err
		}
		startedAtLoggedAt = strings.TrimSpace(startedAtLoggedAt)
		startedAtLogTime, err := time.ParseDateTime(startedAtLoggedAt)
		if err != nil {
			return "", 0, err
		}

		// startedAtLoggedAt must be greater than digestLoggedAt,
		// because image pulled and build before it starts serving requests.
		// If startedAtLoggedAt is less than digestLoggedAt, then the container is not running.
		if startedAtLogTime.Before(digestLogTime) {
			return "", 0, nil
		}
		startedAt = startedAtLogTime.Unix()
	}

	return fingerprint, startedAt, nil
}
