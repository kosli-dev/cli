package azure

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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
	"github.com/distribution/reference"
	"github.com/kosli-dev/cli/internal/digest"
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

// These are for handling temporary 503 errors so that they do not fail the whole command
var ErrAppUnavailable = errors.New("app is unavailable (503)")

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
func (azureClient *AzureClient) getBearerToken(logger *logger.Logger) (string, error) {
	oauthURL, err := url.JoinPath("https://login.microsoftonline.com", azureClient.Credentials.TenantId, "oauth2/token")
	if err != nil {
		return "", err
	}

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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log warning for cleanup error
			logger.Warn("failed to close response body: %v", err)
		}
	}()

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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log warning for cleanup error
			fmt.Printf("warning: failed to close response body: %v\n", err)
		}
	}()
	if resp.StatusCode == http.StatusServiceUnavailable {
		return ErrAppUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download package for app [%s]: %s", appName, resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			// Log warning for cleanup error
			fmt.Printf("warning: failed to close file %s: %v\n", destination, err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (azureClient *AzureClient) fingerprintZipService(app *armappservice.Site, logger *logger.Logger) (AppData, error) {
	// get bearer token
	token, err := azureClient.getBearerToken(logger)
	if err != nil {
		return AppData{}, err
	}
	// download package
	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return AppData{}, err
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			logger.Warn("failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	packagePath := filepath.Join(tmpDir, *app.Name+".zip")
	err = downloadAppPackage(*app.Name, token, packagePath)
	if err != nil {
		if err == ErrAppUnavailable {
			logger.Debug("app %s is unavailable (503), skipping", *app.Name)
			return AppData{}, nil
		}
		return AppData{}, err
	}

	// unzip the downloaded package
	destDir := filepath.Join(tmpDir, "extracted")
	err = unzip(packagePath, destDir, logger)
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
func unzip(zipFile, destDir string, logger *logger.Logger) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			// Log warning for cleanup error
			logger.Warn("failed to close zip reader: %v", err)
		}
	}()

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
		if closeErr := destFile.Close(); closeErr != nil {
			// Log warning for cleanup error
			logger.Warn("failed to close destination file %s: %v", filePath, closeErr)
		}
		if closeErr := zipFile.Close(); closeErr != nil {
			// Log warning for cleanup error
			logger.Warn("failed to close zip file: %v", closeErr)
		}

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
		fingerprint, err = azureClient.GetImageFingerprint(imageName, logger)
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
		fingerprint, startedAt, err = extractImageFingerprintAndStartedTimestampFromLogs(logs, *app.Name)
		if err != nil {
			return AppData{}, err
		}
	}

	logger.Debug("For app %s found: image=%s, fingerprint=%s, startedAt=%d", *app.Name, imageName, fingerprint, startedAt)

	return AppData{*app.Name, *app.Kind, fingerprintSource, map[string]string{imageName: fingerprint}, startedAt}, nil
}

// acrLoginServerSuffixes are the Azure Container Registry login-server suffixes
// for the public, China and US Government clouds. The Azure SDK publishes only
// the token audience per cloud, not the login-server suffix.
var acrLoginServerSuffixes = []string{".azurecr.io", ".azurecr.cn", ".azurecr.us"}

// imageFingerprintSource is how an image reference is resolved to a fingerprint.
type imageFingerprintSource int

const (
	// fingerprintFromACR reads the fingerprint from Azure Container Registry,
	// authenticated with the Azure credential.
	fingerprintFromACR imageFingerprintSource = iota
	// fingerprintFromAnonymousRegistry reads it from any other registry with no
	// credential attached.
	fingerprintFromAnonymousRegistry
)

// isACRLoginServer reports whether domain is an Azure Container Registry login
// server, matching on a whole label so that "azurecr.io.example.com" is not one.
func isACRLoginServer(domain string) bool {
	h := strings.ToLower(domain)
	if hostWithoutPort, _, err := net.SplitHostPort(h); err == nil {
		h = hostWithoutPort
	}
	for _, suffix := range acrLoginServerSuffixes {
		if len(h) > len(suffix) && strings.HasSuffix(h, suffix) {
			return true
		}
	}
	return false
}

// fingerprintPlan is how one image reference will be resolved. It is decided
// before anything is contacted, so a test can assert every value that crosses
// the boundary rather than only which resolver ran.
type fingerprintPlan struct {
	source imageFingerprintSource
	// domain is the registry the reference names, as the parser reports it.
	domain string
	// reference is the canonical form handed to a resolver. Classification and
	// resolution use this same value, so they cannot disagree about the host.
	reference string
	// repoPath and tagOrDigest address the manifest on the ACR arm.
	repoPath    string
	tagOrDigest string
	// pinnedFingerprint is the sha256 hex a digest-pinned reference claims, or
	// empty when the reference is not pinned.
	pinnedFingerprint string
}

// planImageFingerprint decides how an App Service image reference is resolved.
//
// The reference is parsed with the same normalising parser the registry clients
// use rather than being split by hand, because a hand-rolled split can be talked
// into disagreeing with the client about which host it named:
// "reg.azurecr.io:443@attacker.example/repo:tag" passes a suffix check on the
// registry component but resolves to attacker.example as a URL. The parser
// rejects it.
//
// Only fingerprintFromACR attaches the Azure credential, so the classification
// here is what keeps that credential away from a registry named in an app's own
// configuration.
func planImageFingerprint(imageName string) (fingerprintPlan, error) {
	named, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return fingerprintPlan{}, fmt.Errorf("failed to parse the image name [%s]: %w", imageName, err)
	}

	var plan fingerprintPlan

	if digested, ok := named.(reference.Digested); ok {
		// A reference pinned to an algorithm Kosli cannot fingerprint can never
		// match, so reject it here rather than after a pointless round trip.
		plan.pinnedFingerprint, err = digest.Sha256Fingerprint(digested.Digest())
		if err != nil {
			return fingerprintPlan{}, fmt.Errorf("image [%s] is pinned to a digest Kosli cannot use: %w", imageName, err)
		}
		// A digest is authoritative when a reference carries both, and
		// containers/image refuses a reference holding a tag and a digest
		// together, so drop the tag.
		named, err = reference.WithDigest(reference.TrimNamed(named), digested.Digest())
		if err != nil {
			return fingerprintPlan{}, fmt.Errorf("failed to normalise the image name [%s]: %w", imageName, err)
		}
		plan.tagOrDigest = digested.Digest().String()
	} else {
		named = reference.TagNameOnly(named)
		tagged, ok := named.(reference.Tagged)
		if !ok {
			return fingerprintPlan{}, fmt.Errorf("image [%s] names neither a tag nor a digest", imageName)
		}
		plan.tagOrDigest = tagged.Tag()
	}

	plan.domain = reference.Domain(named)
	plan.repoPath = reference.Path(named)
	plan.reference = named.String()
	plan.source = fingerprintFromAnonymousRegistry
	if isACRLoginServer(plan.domain) {
		plan.source = fingerprintFromACR
	}

	return plan, nil
}

// GetImageFingerprint resolves the fingerprint of a container image referenced
// by a Web App. The registry comes from the app's own configuration, which
// anyone with write access to that app controls, so the Azure credential is
// attached only for an Azure Container Registry login server.
func (azureClient *AzureClient) GetImageFingerprint(imageName string, logger *logger.Logger) (string, error) {
	plan, err := planImageFingerprint(imageName)
	if err != nil {
		return "", err
	}

	var fingerprint string
	if plan.source == fingerprintFromACR {
		fingerprint, err = azureClient.acrImageFingerprint(plan, nil, logger)
	} else {
		fingerprint, err = anonymousImageFingerprint(plan, logger)
	}
	if err != nil {
		return "", err
	}

	// A pinned reference is a claim about which image is deployed, and neither
	// resolver checks the digest it is given against the one it gets back, so
	// hold the registry to it here.
	if plan.pinnedFingerprint != "" && fingerprint != plan.pinnedFingerprint {
		return "", fmt.Errorf("image [%s] is pinned to digest sha256:%s but [%s] reported sha256:%s", imageName, plan.pinnedFingerprint, plan.domain, fingerprint)
	}

	return fingerprint, nil
}

// acrImageFingerprint reads a fingerprint from Azure Container Registry using
// the Azure credential supplied to Kosli.
// clientOptions is nil in production; tests pass options carrying a transport
// pointed at a fake registry, so the arm is exercised without package-level state.
func (azureClient *AzureClient) acrImageFingerprint(plan fingerprintPlan, clientOptions *azcontainerregistry.ClientOptions, logger *logger.Logger) (string, error) {
	credentials, err := azidentity.NewClientSecretCredential(azureClient.Credentials.TenantId,
		azureClient.Credentials.ClientId, azureClient.Credentials.ClientSecret, nil)
	if err != nil {
		return "", err
	}

	acrClient, err := azcontainerregistry.NewClient("https://"+plan.domain, credentials, clientOptions)
	if err != nil {
		return "", err
	}

	manifestRes, err := acrClient.GetManifest(context.TODO(), plan.repoPath, plan.tagOrDigest,
		&azcontainerregistry.ClientGetManifestOptions{Accept: to.Ptr("application/vnd.docker.distribution.manifest.v2+json")})
	if err != nil {
		return "", err
	}
	if manifestRes.DockerContentDigest == nil {
		return "", fmt.Errorf("no digest returned for image [%s]", plan.reference)
	}

	fingerprint, err := digest.Sha256FingerprintFromDigest(*manifestRes.DockerContentDigest)
	if err != nil {
		return "", fmt.Errorf("registry reported a digest Kosli cannot use for image [%s]: %w", plan.reference, err)
	}

	logger.Debug("For image '%s' got fingerprint '%s' from ACR", plan.reference, fingerprint)

	return fingerprint, nil
}

// anonymousFingerprint resolves a fingerprint with no credential presented. It
// is a variable so tests can assert the reference the resolver is handed.
var anonymousFingerprint = digest.OciSha256Anonymous

// anonymousImageFingerprint reads a fingerprint from a registry outside Azure
// Container Registry, presenting no credential.
func anonymousImageFingerprint(plan fingerprintPlan, logger *logger.Logger) (string, error) {
	fingerprint, err := anonymousFingerprint(plan.reference)
	if err != nil {
		return "", fmt.Errorf("failed to get the fingerprint of image [%s] from [%s]: %w. Azure credentials are only sent to Azure Container Registry; use --digests-source logs for this app", plan.reference, plan.domain, err)
	}

	logger.Debug("For image '%s' got fingerprint '%s' from '%s' with no credentials", plan.reference, fingerprint, plan.domain)

	return fingerprint, nil
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
			defer func() {
				if err := response.Body.Close(); err != nil {
					logger.Warn("failed to close response body: %v", err)
				}
			}()
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
		defer func() {
			if err := response.Body.Close(); err != nil {
				logger.Warn("failed to close response body: %v", err)
			}
		}()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
}

func extractImageFingerprintAndStartedTimestampFromLogs(logs []byte, appName string) (fingerprint string, startedAt int64, error error) {
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
