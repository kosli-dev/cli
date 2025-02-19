package azure

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	smithyTime "github.com/aws/smithy-go/time"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/server"
)

type AzureStaticCredentials struct {
	TenantId          string
	ClientId          string
	ClientSecret      string
	SubscriptionId    string
	ResourceGroupName string
	DownloadLogsAsZip bool
	DigestsSource     string
}

type AzureClient struct {
	Credentials       AzureStaticCredentials
	AppServiceFactory *armappservice.ClientFactory
}

// AppData represents the harvested Azure service app and function app data
type AppData struct {
	AppName       string            `json:"app_name"`
	AppKind       string            `json:"app_kind"`
	DigestsSource string            `json:"digests_source"`
	Digests       map[string]string `json:"digests"`
	StartedAt     int64             `json:"creationTimestamp"`
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
	logger.Debug("Found apps:")
	for _, app := range appsInfo {
		logger.Debug("  app Name=%s", *app.Name)
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
	notDocker := len(linuxFxVersion) != 2 || linuxFxVersion[0] != "DOCKER"
	if notDocker {
		return azureClient.fingerprintZipService(app, logger)
	} else {
		return azureClient.fingerprintDockerService(app, logger, linuxFxVersion[1])
	}
}

// getBearerToken gets a bearer token
func (azureClient *AzureClient) getBearerToken() (string, error) {
	oauthURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", azureClient.Credentials.TenantId)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", azureClient.Credentials.ClientId)
	data.Set("client_secret", azureClient.Credentials.ClientSecret)
	data.Set("resource", "https://management.azure.com/")

	req, err := http.NewRequest("POST", oauthURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var oauthResp map[string]interface{}
	err = json.Unmarshal(body, &oauthResp)
	if err != nil {
		return "", err
	}
	accessToken := oauthResp["access_token"].(string)
	return accessToken, nil
}

// downloadAppPackage downloads the zip package of a non-docker web app
func downloadAppPackage(appName, bearerToken, destination string) error {
	kuduZipURL := fmt.Sprintf("https://%s.scm.azurewebsites.net/api/zip/site/wwwroot/", appName)
	req, err := http.NewRequest("GET", kuduZipURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download package for app [%s]: %s", appName, resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (azureClient *AzureClient) fingerprintZipService(app *armappservice.Site, logger *logger.Logger) (AppData, error) {
	// get bearer token
	token, err := azureClient.getBearerToken()
	if err != nil {
		return AppData{}, err
	}
	// download package
	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return AppData{}, err
	}
	defer os.RemoveAll(tmpDir)

	packagePath := filepath.Join(tmpDir, *app.Name+".zip")
	err = downloadAppPackage(*app.Name, token, packagePath)
	if err != nil {
		return AppData{}, err
	}

	// unzip the downloaded package
	destDir := filepath.Join(tmpDir, "extracted")
	err = unzip(packagePath, destDir)
	if err != nil {
		return AppData{}, fmt.Errorf("failed to unzip downloaded package for app [%s]: %v", *app.Name, err)
	}

	//  fingerprint the downloaded and unzipped package
	ps := &server.PathsSpec{
		Version: 1,
		Artifacts: map[string]server.ArtifactPathSpec{
			*app.Name: {
				Path: destDir,
			},
		},
	}

	artifacts, err := server.CreatePathsArtifactsData(ps, logger)
	if err != nil {
		return AppData{}, err
	}

	// webAppsClient := azureClient.AppServiceFactory.NewWebAppsClient()
	// deploymentsPager := webAppsClient.NewListDeploymentsPager(resourceGroupName, *app.Name, &armappservice.WebAppsClientListDeploymentsOptions{})
	// var deploymentsInfo []*armappservice.Deployment
	// ctx := context.Background()
	// for deploymentsPager.More() {
	// 	response, err := deploymentsPager.NextPage(ctx)
	// 	if err != nil {
	// 		return AppData{}, err
	// 	}
	// 	deploymentsInfo = append(deploymentsInfo, response.Value...)
	// }
	// var startedAt int64
	// var deploymentTime *time.Time
	// for _, deploymentInfo := range deploymentsInfo {
	// 	if *deploymentInfo.Properties.Active {
	// 		deploymentTime = deploymentInfo.Properties.StartTime
	// 	}
	// }
	// if deploymentTime != nil {
	// 	startedAt = deploymentTime.Unix()
	// }
	return AppData{*app.Name, *app.Kind, "kosli-cli", artifacts[0].Digests, 0}, nil
}

// unzip extracts a zip archive to a specified destination directory.
func unzip(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		filePath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			// Create directories
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		// Ensure the directory for the file exists
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		// Open the destination file
		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// Open the source file within the ZIP archive
		zipFile, err := f.Open()
		if err != nil {
			return err
		}

		// Copy the file contents
		_, err = io.Copy(destFile, zipFile)

		// Close the open files
		destFile.Close()
		zipFile.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func (azureClient *AzureClient) fingerprintDockerService(app *armappservice.Site, logger *logger.Logger, imageName string) (AppData, error) {
	var fingerprint string
	var startedAt int64
	var fingerprintSource string
	var err error

	if azureClient.Credentials.DigestsSource == "acr" {
		fingerprintSource = "acr"
		fingerprint, err = azureClient.GetImageFingerprintFromRegistry(imageName, logger)
		// Handle exception when image is not found in the registry but is found in the environment
		if err != nil {
			return AppData{}, err
		}
	} else {
		fingerprintSource = "logs"
		logs, err := azureClient.GetDockerLogsForApp(*app.Name, logger)
		if err != nil {
			return AppData{}, err
		}
		fingerprint, startedAt, err = exractImageFingerprintAndStartedTimestampFromLogs(logs, *app.Name)
		if err != nil {
			return AppData{}, err
		}
	}

	logger.Debug("For app %s found: image=%s, fingerprint=%s, startedAt=%d", *app.Name, imageName, fingerprint, startedAt)

	return AppData{*app.Name, *app.Kind, fingerprintSource, map[string]string{imageName: fingerprint}, startedAt}, nil
}

func (azureClient *AzureClient) GetImageFingerprintFromRegistry(imageName string, logger *logger.Logger) (fingerprint string, err error) {
	registryUrl, repoName, tag := parseImageName(imageName)

	credentials, err := azidentity.NewClientSecretCredential(azureClient.Credentials.TenantId,
		azureClient.Credentials.ClientId, azureClient.Credentials.ClientSecret, nil)
	if err != nil {
		return "", err
	}

	AcrClient, err := azcontainerregistry.NewClient(registryUrl, credentials, nil)
	if err != nil {
		return "", err
	}

	manifestRes, err := AcrClient.GetManifest(context.TODO(), repoName, tag,
		&azcontainerregistry.ClientGetManifestOptions{Accept: to.Ptr("application/vnd.docker.distribution.manifest.v2+json")})
	if err != nil {
		return "", err
	}

	manifestPropsRes, err := AcrClient.GetManifestProperties(context.TODO(), repoName, *manifestRes.DockerContentDigest, nil)
	if err != nil {
		return "", err
	}

	fingerprint = strings.TrimPrefix(*manifestPropsRes.Manifest.Digest, "sha256:")

	logger.Debug("For image '%s' got fingerprint '%s' from ACR", imageName, fingerprint)

	return fingerprint, nil
}

func parseImageName(imageName string) (registryUrl, repoName, tag string) {
	// Parse the image name to extract the repository name and tag
	// Example: tookyregistry.azurecr.io/tooky/sha256:latest
	splitFullImageName := strings.SplitN(imageName, "/", 2)
	if len(splitFullImageName) != 2 {
		return "", "", ""
	}

	registryUrl = fmt.Sprintf("https://%s", splitFullImageName[0])

	if strings.Contains(splitFullImageName[1], "@sha256:") {
		// Example: tookyregistry.azurecr.io/tooky@sha256:cb29a6..7
		imageNameAndTag := strings.SplitN(splitFullImageName[1], "@", 2)
		repoName = imageNameAndTag[0]
		tag = imageNameAndTag[1]
	} else if strings.Contains(splitFullImageName[1], ":") {
		imageNameAndTag := strings.SplitN(splitFullImageName[1], ":", 2)
		repoName = imageNameAndTag[0]
		tag = imageNameAndTag[1]
	} else {
		repoName = splitFullImageName[1]
		tag = "latest"
	}

	return registryUrl, repoName, tag
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

func (azureClient *AzureClient) GetDockerLogsForApp(appServiceName string, logger *logger.Logger) (logs []byte, error error) {
	appsClient := azureClient.AppServiceFactory.NewWebAppsClient()

	ctx := context.Background()

	if azureClient.Credentials.DownloadLogsAsZip {
		response, err := appsClient.GetContainerLogsZip(ctx, azureClient.Credentials.ResourceGroupName, appServiceName, nil)
		if err != nil {
			return nil, err
		}
		logger.Debug("Got logs for app service: ", appServiceName)
		if response.Body != nil {
			defer response.Body.Close()
		}
		logger.Debug("Reading logs for app service: ", appServiceName)
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		zipFileName := fmt.Sprintf("%s-logs.zip", appServiceName)
		// TODO: write body to a file
		logger.Debug("Writing logs for app service: ", appServiceName, " to file: ", zipFileName)
		err = os.WriteFile("zipFileName", body, 0o644)
		if err != nil {
			return nil, err
		}
		// TODO: read zip file and return logs
		return nil, nil
	} else {
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
		digestLogTime, err := smithyTime.ParseDateTime(digestLoggedAt)
		if err != nil {
			return "", 0, err
		}
		startedAtLoggedAt = strings.TrimSpace(startedAtLoggedAt)
		startedAtLogTime, err := smithyTime.ParseDateTime(startedAtLoggedAt)
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
