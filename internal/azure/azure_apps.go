package azure

import (
	"archive/zip"
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
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/distribution/reference"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/kosli-dev/cli/internal/utils"
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
	// acrClientOptions is nil in production. Tests set it so the ACR arm can be
	// driven against a fake registry without package-level state.
	acrClientOptions *azcontainerregistry.ClientOptions
	// dockerLogsForApp is a test seam; nil falls back to GetDockerLogsForApp.
	dockerLogsForApp func(appServiceName string, logger *logger.Logger) ([]byte, error)
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

func warnAboutDigestsSource(digestsSource string, logger *logger.Logger) {
	if digestsSource != "logs" {
		return
	}
	logger.Warn("--digests-source logs reads each app's image digest from its docker log, which the running container can also write to; a compromised container may misreport its image. Prefer --digests-source acr where the registry allows it")
}

func (staticCreds *AzureStaticCredentials) GetAzureAppsData(logger *logger.Logger) (appsData []*AppData, err error) {
	warnAboutDigestsSource(staticCreds.DigestsSource, logger)

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
				// One app's error cancels the run, so say which app it was. Wrapped
				// here rather than at each return so every path is covered once.
				err = fmt.Errorf("app [%s]: %w", *app.Name, err)
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
		return AppData{}, fmt.Errorf("failed to unzip the downloaded package: %w", err)
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

	// Resolved path to the entry name that produced it, so a collision can name
	// both sides.
	extracted := make(map[string]string, len(r.File))

	for _, f := range r.File {
		// The entry name comes from the deployed package, which anyone able to
		// deploy the app controls, so it must not be able to leave destDir.
		filePath, err := utils.ContainedPath(destDir, f.Name)
		if errors.Is(err, utils.ErrNamesNoFile) && f.FileInfo().IsDir() {
			// A "./" entry names destDir itself; there is nothing to create.
			continue
		}
		if err != nil {
			// One app's error cancels the whole run, snapshot azure has no exclude
			// flag, and .kosli_ignore is only read after extraction, so the only
			// remedy is a changed package.
			return fmt.Errorf("zip entry %w; the package cannot be extracted safely, so no app in the environment is reported until this app is redeployed without that entry", err)
		}

		if err := extractZipEntry(f, filePath, extracted, logger); err != nil {
			return fmt.Errorf("zip entry [%s]: %w", f.Name, err)
		}
	}
	return nil
}

// extractZipEntry writes one entry to filePath, which the caller has already
// checked stays inside the destination directory, and records it in extracted.
func extractZipEntry(f *zip.File, filePath string, extracted map[string]string, logger *logger.Logger) error {
	isDir := f.FileInfo().IsDir()

	// Legal in a zip, impossible on disk: one name as both a file and a
	// directory. Checked up front so the message names this entry rather than
	// whichever filesystem call happens to fail, and fails the same way on
	// every platform.
	if existing, statErr := os.Lstat(filePath); statErr == nil && existing.IsDir() != isDir {
		if existing.IsDir() {
			return errors.New("was already extracted as a directory")
		}
		return errors.New("was already extracted as a file")
	}

	dir := filePath
	if !isDir {
		dir = filepath.Dir(filePath)
	}
	err := os.MkdirAll(dir, 0o700)
	if errors.Is(err, syscall.ENOTDIR) {
		// A file "a" and an entry under "a/".
		return errors.New("one of its parent directories was already extracted as a file")
	}
	if err != nil {
		return err
	}
	if isDir {
		return nil
	}

	// The tree is deleted once fingerprinted and the fingerprint never reads a
	// mode, so a constant is safe: it keeps a zero-mode entry readable and
	// drops setuid, setgid and sticky, which os.OpenFile would honour.
	// Writing every entry through OpenFile, never os.Symlink, is what keeps a
	// symlink entry from becoming a real symlink. Name containment does not
	// survive one, since a later entry or the fingerprinter would follow it
	// out of the destination.
	// Legal in a zip, and containment collapses "x", "/x" and "./x" onto one
	// path, but overwriting would fingerprint the package without the first
	// entry. This fails the same way on every platform.
	if first, dup := extracted[filePath]; dup {
		return fmt.Errorf("resolves to the same local path as entry [%s]", first)
	}
	destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		// No recorded entry resolves here, so the filesystem itself equates two
		// names: case, or Unicode form, on macOS and Windows. Moving the
		// snapshot is the operator's only remedy.
		return errors.New("collides with an earlier entry whose name this filesystem treats as the same, such as one differing only in case; run the snapshot on a case-sensitive filesystem")
	}
	if err != nil {
		return err
	}
	extracted[filePath] = f.Name

	zipFile, err := f.Open()
	if err != nil {
		if closeErr := destFile.Close(); closeErr != nil {
			logger.Warn("failed to close destination file %s: %v", filePath, closeErr)
		}
		return err
	}

	_, err = io.Copy(destFile, zipFile)

	// A write error can surface only at Close, and an entry truncated that
	// way would be fingerprinted as if it were complete.
	if closeErr := destFile.Close(); closeErr != nil {
		logger.Warn("failed to close destination file %s: %v", filePath, closeErr)
		if err == nil {
			err = closeErr
		}
	}
	if closeErr := zipFile.Close(); closeErr != nil {
		logger.Warn("failed to close zip file: %v", closeErr)
	}

	return err
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
		fetchLogs := azureClient.dockerLogsForApp
		if fetchLogs == nil {
			fetchLogs = azureClient.GetDockerLogsForApp
		}
		logs, err := fetchLogs(*app.Name, logger)
		if err != nil {
			return AppData{}, err
		}
		fingerprint, startedAt = extractImageFingerprintAndStartedTimestampFromLogs(logs, *app.Name, imageName, logger)
		if fingerprint == "" {
			logger.Warn("no platform-written image digest found in the docker log for app [%s]; it is reported without a fingerprint. The container may not have started within the log window, or the platform may log its start differently for this app kind; --digests-source acr reads the digest from the registry instead", *app.Name)
		}
	}

	logger.Debug("For app %s found: image=%s, fingerprint=%s, startedAt=%d", *app.Name, imageName, fingerprint, startedAt)

	return AppData{*app.Name, *app.Kind, fingerprintSource, map[string]string{imageName: fingerprint}, startedAt}, nil
}

// acrLoginServerSuffixes are the Azure Container Registry login-server suffixes
// for the public, China and US Government clouds. The Azure SDK publishes only
// the token audience per cloud, not the login-server suffix.
var acrLoginServerSuffixes = []string{".azurecr.io", ".azurecr.cn", ".azurecr.us"}

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
// Only an Azure Container Registry login server gets the Azure credential, so
// the domain this reports is what keeps that credential away from a registry
// named in an app's own configuration.
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
	if isACRLoginServer(plan.domain) {
		fingerprint, err = azureClient.acrImageFingerprint(plan, azureClient.acrClientOptions, logger)
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
		return "", fmt.Errorf("image [%s] is pinned to digest sha256:%s but [%s] reported sha256:%s", plan.reference, plan.pinnedFingerprint, plan.domain, fingerprint)
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
	if manifestRes.ManifestData != nil {
		defer func() {
			if err := manifestRes.ManifestData.Close(); err != nil {
				logger.Warn("failed to close the manifest response for image %s: %v", plan.reference, err)
			}
		}()
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

	azureClient := &AzureClient{
		Credentials:       *staticCreds,
		AppServiceFactory: appserviceFactory,
	}
	azureClient.dockerLogsForApp = azureClient.GetDockerLogsForApp
	return azureClient, nil
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

// platformLinePrefix opens every line the App Service platform writes to the
// docker log. Container stdout shares the stream but is timestamped by Docker
// at nanosecond precision with no level or dash, so this shape is what tells a
// platform line from a container line written to look like one.
//
// Submatches are read by name (see group), so the prefix may gain groups.
const platformLinePrefix = `^(?P<ts>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z) [A-Z]+\s+-\s+`

var (
	platformDigestLine = regexp.MustCompile(platformLinePrefix + `Digest: sha256:(?P<digest>[0-9a-f]{64})\s*$`)
	// The "docker run" line is the start marker: the container's output can only
	// appear after it, and unlike "Starting container for site" it names the
	// site and the image.
	platformRunLine   = regexp.MustCompile(platformLinePrefix + `docker run (?P<args>.*)$`)
	platformReadyLine = regexp.MustCompile(platformLinePrefix +
		`Container \S+ for site (?P<site>\S+) initialized successfully and is ready to serve requests\.\s*$`)
)

// group is the named submatch of m, or "" for a name re lacks; a bad name must
// not panic the snapshot.
func group(re *regexp.Regexp, m [][]byte, name string) string {
	i := re.SubexpIndex(name)
	if i < 0 || i >= len(m) {
		return ""
	}
	return string(m[i])
}

// runLineIsForSite reports whether a "docker run" argument list is the site's,
// by WEBSITE_SITE_NAME or by the container name. App names are hostnames, so
// case is ignored.
func runLineIsForSite(args []string, appName string) bool {
	for i, arg := range args {
		if strings.EqualFold(arg, "WEBSITE_SITE_NAME="+appName) {
			return true
		}
		name, ok := strings.CutPrefix(arg, "--name=")
		if !ok && arg == "--name" && i+1 < len(args) {
			name, ok = args[i+1], true
		}
		if ok && containerNameIsForSite(name, appName) {
			return true
		}
	}
	return false
}

// containerNameIsForSite matches "<site>_<instance>_<hash>". The instance digit
// is required so a slot's "<site>__<slot>_..." does not pass.
func containerNameIsForSite(name, appName string) bool {
	rest, ok := strings.CutPrefix(strings.ToLower(name), strings.ToLower(appName)+"_")
	return ok && rest != "" && rest[0] >= '0' && rest[0] <= '9'
}

// runLineDigest is the digest a "docker run" argument list pins the configured
// repository to, or "" when it ran a tag or is ambiguous. The line is logged
// unquoted and split on whitespace, so an app setting value or startup command
// can also yield a token naming the repository. Every such token is a claim
// about which image ran, a tag as much as a digest, and they must all agree.
func runLineDigest(args []string, repository string) string {
	var found string
	for _, arg := range args {
		named, err := reference.ParseNormalizedNamed(arg)
		if err != nil || canonicalRepository(named) != repository {
			continue
		}
		var hex string
		if digested, ok := named.(reference.Digested); ok {
			if hex, err = digest.Sha256Fingerprint(digested.Digest()); err != nil {
				hex = ""
			}
		}
		if hex == "" || (found != "" && found != hex) {
			return ""
		}
		found = hex
	}
	return found
}

// canonicalRepository is a reference's repository with the Docker Hub short form
// expanded, tag and digest dropped, and the host lowercased, so the app's
// configuration and the platform's "docker run" line compare equal.
func canonicalRepository(named reference.Named) string {
	named = reference.TrimNamed(named)
	return strings.ToLower(reference.Domain(named)) + "/" + reference.Path(named)
}

// extractImageFingerprintAndStartedTimestampFromLogs reads the digest of the
// container running for appName from the app's docker log. imageName is the
// reference the app's configuration names.
//
// The digest is the platform's own record of the site's last start, its
// "docker run" line: the digest it pins the configured image to or, when it ran
// a tag, the last platform "Digest:" line before it. Anything after the start
// is the container's own output and cannot name the fingerprint
// (kosli-dev/server#6881). The latest start wins because one log window can
// hold several deployments.
//
// startedAt is when the platform reported that container ready, 0 if it has not
// yet. A start after a ready report is reported only when it names the digest
// that was ready, since that is a scale-out or restart and keeps the earlier
// ready time. A different digest is a replacement in flight, and a ready
// container whose own start precedes the window has an unknown digest; in both
// cases nothing is reported until the new container is ready.
//
// No log content can fail the snapshot: lines have no length cap, and a line of
// platform shape but invalid content is skipped.
func extractImageFingerprintAndStartedTimestampFromLogs(logs []byte, appName, imageName string, logger *logger.Logger) (fingerprint string, startedAt int64) {
	var repository string
	if named, err := reference.ParseNormalizedNamed(imageName); err == nil {
		repository = canonicalRepository(named)
	}

	var pulledDigest, startedDigest, readyDigest string
	var startSeen, initializedSeen, initializedAfterLastStart bool

	for line := range bytes.Lines(logs) {
		line = bytes.TrimRight(line, "\r\n")
		if m := platformDigestLine.FindSubmatch(line); m != nil {
			pulledDigest = group(platformDigestLine, m, "digest")
		} else if m := platformRunLine.FindSubmatch(line); m != nil {
			args := strings.Fields(group(platformRunLine, m, "args"))
			if !runLineIsForSite(args, appName) {
				logger.Debug("skipping a docker run line that does not name site %s: %s", appName, line)
				continue
			}
			startedDigest = pulledDigest
			if runDigest := runLineDigest(args, repository); runDigest != "" {
				startedDigest = runDigest
			}
			startSeen = true
			initializedAfterLastStart = false
		} else if m := platformReadyLine.FindSubmatch(line); m != nil && strings.EqualFold(group(platformReadyLine, m, "site"), appName) {
			readyAt, err := time.Parse(time.RFC3339Nano, group(platformReadyLine, m, "ts"))
			if err != nil {
				continue
			}
			initializedSeen = true
			if !startSeen {
				continue
			}
			startedAt = readyAt.Unix()
			readyDigest = startedDigest
			initializedAfterLastStart = true
		}
	}

	if !startSeen || startedDigest == "" {
		return "", 0
	}
	if initializedSeen && !initializedAfterLastStart && startedDigest != readyDigest {
		return "", 0
	}
	return startedDigest, startedAt
}
