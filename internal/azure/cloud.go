package azure

import (
	"fmt"

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
// the credentials are used against.
func (staticCreds *AzureStaticCredentials) CloudConfiguration() (cloud.Configuration, error) {
	name := staticCreds.Cloud
	if name == "" {
		name = "public"
	}
	cfg, ok := azureClouds[name]
	if !ok {
		return cloud.Configuration{}, fmt.Errorf("invalid value [%s] for --azure-cloud flag. Valid values are 'public', 'china' and 'usgov'", name)
	}
	return cfg, nil
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
	return azidentity.NewClientSecretCredential(staticCreds.TenantId, staticCreds.ClientId, staticCreds.ClientSecret,
		&azidentity.ClientSecretCredentialOptions{ClientOptions: clientOptions(cfg, transport)})
}

// clientOptions points an SDK client at cfg. transport is nil in production.
func clientOptions(cfg cloud.Configuration, transport policy.Transporter) azcore.ClientOptions {
	return azcore.ClientOptions{Cloud: cfg, Transport: transport}
}
