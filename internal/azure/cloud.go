package azure

import (
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
)

var azureClouds = map[string]cloud.Configuration{
	"public": cloud.AzurePublic,
	"china":  cloud.AzureChina,
	"usgov":  cloud.AzureGovernment,
}

// CloudConfiguration is the Azure cloud whose authority and Resource Manager
// the credentials are used against: the named cloud, with AuthorityHost and
// ResourceManagerURL replacing its endpoints where they are set.
func (staticCreds *AzureStaticCredentials) CloudConfiguration() (cloud.Configuration, error) {
	name := staticCreds.cloudName()
	base, ok := azureClouds[name]
	if !ok {
		return cloud.Configuration{}, fmt.Errorf("invalid value [%s] for --azure-cloud flag. Valid values are 'public', 'china' and 'usgov'", name)
	}

	// Services is shared with the SDK's package-level configuration.
	cfg := cloud.Configuration{
		ActiveDirectoryAuthorityHost: base.ActiveDirectoryAuthorityHost,
		Services:                     maps.Clone(base.Services),
	}

	if staticCreds.AuthorityHost != "" {
		if err := requireHTTPSURL("azure-authority-host", staticCreds.AuthorityHost); err != nil {
			return cloud.Configuration{}, err
		}
		cfg.ActiveDirectoryAuthorityHost = staticCreds.AuthorityHost
	}

	if staticCreds.ResourceManagerURL != "" {
		if err := requireHTTPSURL("azure-resource-manager-url", staticCreds.ResourceManagerURL); err != nil {
			return cloud.Configuration{}, err
		}
		// The audience stays the base cloud's: a token for a private Resource
		// Manager is still requested from the authority under that name.
		resourceManager := cfg.Services[cloud.ResourceManager]
		resourceManager.Endpoint = staticCreds.ResourceManagerURL
		cfg.Services[cloud.ResourceManager] = resourceManager
	}

	return cfg, nil
}

func (staticCreds *AzureStaticCredentials) cloudName() string {
	if staticCreds.Cloud == "" {
		return "public"
	}
	return staticCreds.Cloud
}

// kuduResources is the token resource Kudu is sent per cloud: the one this
// command has always used in the public cloud, and the one the Azure CLI uses
// in the others.
var kuduResources = map[string]string{
	"public": "https://management.azure.com/",
	"china":  "https://management.core.chinacloudapi.cn/",
	"usgov":  "https://management.core.usgovcloudapi.net/",
}

// appServiceDomains is the App Service DNS domain per cloud.
var appServiceDomains = map[string]string{
	"public": "azurewebsites.net",
	"china":  "chinacloudsites.cn",
	"usgov":  "azurewebsites.us",
}

// checkKuduHost refuses a host other than the app's own SCM host in the
// cloud's App Service domain, in its classic form "<app>.scm.<domain>" or its
// unique default hostname form "<app>-<hash>.scm.<region>.<domain>".
func checkKuduHost(cloudName, appName, host string) error {
	domain := regexp.QuoteMeta(appServiceDomains[cloudName])
	app := regexp.QuoteMeta(strings.ToLower(appName))
	classic := regexp.MustCompile(`^` + app + `\.scm\.` + domain + `$`)
	unique := regexp.MustCompile(`^` + app + `-[a-z0-9]+\.scm\.[a-z0-9-]+\.` + domain + `$`)
	h := strings.ToLower(host)
	if domain != "" && (classic.MatchString(h) || unique.MatchString(h)) {
		return nil
	}
	return fmt.Errorf("app [%s] lists Kudu host [%s], which is not its own App Service SCM host in the [%s] Azure cloud, so its package is not downloaded and no token is sent there. Set --azure-resource-manager-url to trust the host names your Resource Manager lists", appName, host, cloudName)
}

// requireHTTPSURL refuses anything but an absolute https URL, which the SDK
// demands of an authority and of any endpoint it sends a bearer token to.
func requireHTTPSURL(flag, value string) error {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("invalid value [%s] for --%s flag. It must be an absolute https URL", value, flag)
	}
	return nil
}

// acrOptionsForCloud copies base, which may be nil, with cfg as its cloud. The
// registry's token audience is looked up in that cloud's services.
func acrOptionsForCloud(cfg cloud.Configuration, base *azcontainerregistry.ClientOptions) *azcontainerregistry.ClientOptions {
	options := azcontainerregistry.ClientOptions{}
	if base != nil {
		options = *base
	}
	options.Cloud = cfg
	return &options
}

func (staticCreds *AzureStaticCredentials) newCredential(cfg cloud.Configuration, transport policy.Transporter) (*azidentity.ClientSecretCredential, error) {
	if staticCreds.cloudName() == "public" && staticCreds.AuthorityHost == "" {
		// Given no cloud, azidentity uses AZURE_AUTHORITY_HOST when it is set.
		cfg = cloud.Configuration{}
	}
	return azidentity.NewClientSecretCredential(staticCreds.TenantId, staticCreds.ClientId, staticCreds.ClientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: clientOptions(cfg, transport),
			// Instance discovery asks the public cloud's authority about the
			// overriding one, which a stand-in or disconnected cloud cannot reach.
			DisableInstanceDiscovery: staticCreds.AuthorityHost != "",
		})
}

// clientOptions points an SDK client at cfg. transport is nil in production.
func clientOptions(cfg cloud.Configuration, transport policy.Transporter) azcore.ClientOptions {
	return azcore.ClientOptions{Cloud: cfg, Transport: transport}
}
