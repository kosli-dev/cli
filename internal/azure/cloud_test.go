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
