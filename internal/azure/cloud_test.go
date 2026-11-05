package azure

import (
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	"github.com/stretchr/testify/require"
)

func TestCloudConfigurationSelectsTheNamedCloud(t *testing.T) {
	for _, tc := range []struct {
		name string
		want cloud.Configuration
	}{
		{"", cloud.AzurePublic},
		{"public", cloud.AzurePublic},
		{"china", cloud.AzureChina},
		{"usgov", cloud.AzureGovernment},
	} {
		t.Run(tc.name, func(t *testing.T) {
			creds := AzureStaticCredentials{Cloud: tc.name}
			got, err := creds.CloudConfiguration()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestACRClientOptionsCarryTheCloudAndKeepTheTransport(t *testing.T) {
	transport := &http.Client{}
	base := &azcontainerregistry.ClientOptions{ClientOptions: azcore.ClientOptions{Transport: transport}}

	got := acrOptionsForCloud(cloud.AzureChina, base)

	require.Equal(t, cloud.AzureChina, got.Cloud)
	require.Same(t, transport, got.Transport)
	require.Equal(t, cloud.Configuration{}, base.Cloud, "the caller's options must not be modified")

	require.Equal(t, cloud.AzureGovernment, acrOptionsForCloud(cloud.AzureGovernment, nil).Cloud)
}

func TestCloudConfigurationRejectsAnUnknownCloud(t *testing.T) {
	creds := AzureStaticCredentials{Cloud: "germany"}
	_, err := creds.CloudConfiguration()
	require.EqualError(t, err, "invalid value [germany] for --azure-cloud flag. Valid values are 'public', 'china' and 'usgov'")
}

func TestCloudConfigurationOverridesOnlyTheGivenEndpoints(t *testing.T) {
	creds := AzureStaticCredentials{
		Cloud:              "china",
		AuthorityHost:      "https://login.example.test/",
		ResourceManagerURL: "https://management.example.test",
	}

	got, err := creds.CloudConfiguration()

	require.NoError(t, err)
	require.Equal(t, "https://login.example.test/", got.ActiveDirectoryAuthorityHost)
	require.Equal(t, cloud.ServiceConfiguration{
		Endpoint: "https://management.example.test",
		Audience: cloud.AzureChina.Services[cloud.ResourceManager].Audience,
	}, got.Services[cloud.ResourceManager])
	require.Equal(t, cloud.AzureChina.Services[azcontainerregistry.ServiceName], got.Services[azcontainerregistry.ServiceName])
	require.Equal(t, "https://management.chinacloudapi.cn", cloud.AzureChina.Services[cloud.ResourceManager].Endpoint,
		"the SDK's own configuration must not be modified")
	require.Equal(t, "https://login.chinacloudapi.cn/", cloud.AzureChina.ActiveDirectoryAuthorityHost)
}

func TestCloudConfigurationOverridesEachEndpointAlone(t *testing.T) {
	authorityOnly := AzureStaticCredentials{AuthorityHost: "https://login.example.test/"}
	got, err := authorityOnly.CloudConfiguration()
	require.NoError(t, err)
	require.Equal(t, cloud.AzurePublic.Services[cloud.ResourceManager], got.Services[cloud.ResourceManager])

	resourceManagerOnly := AzureStaticCredentials{ResourceManagerURL: "https://management.example.test"}
	got, err = resourceManagerOnly.CloudConfiguration()
	require.NoError(t, err)
	require.Equal(t, cloud.AzurePublic.ActiveDirectoryAuthorityHost, got.ActiveDirectoryAuthorityHost)
}

func TestCloudConfigurationRejectsAnOverrideThatIsNotAnHTTPSURL(t *testing.T) {
	for _, value := range []string{"http://login.example.test/", "login.example.test", "/oauth2", "https://", "://"} {
		t.Run(value, func(t *testing.T) {
			_, err := (&AzureStaticCredentials{AuthorityHost: value}).CloudConfiguration()
			require.EqualError(t, err, "invalid value ["+value+"] for --azure-authority-host flag. It must be an absolute https URL")

			_, err = (&AzureStaticCredentials{ResourceManagerURL: value}).CloudConfiguration()
			require.EqualError(t, err, "invalid value ["+value+"] for --azure-resource-manager-url flag. It must be an absolute https URL")
		})
	}
}

func TestACRReusesTheClientsCredential(t *testing.T) {
	standin := newAzureStandin(t)
	creds := standinCredentials("public")
	client, err := creds.newAzureClient(standin.redirectingTransport())
	require.NoError(t, err)

	credential, err := client.acrCredential(cloud.AzurePublic, nil)

	require.NoError(t, err)
	require.Same(t, client.credential, credential, "a credential per image would skip the token cache")
}
