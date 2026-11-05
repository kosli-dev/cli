package azure

import (
	"context"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestKuduTokenIsForTheChosenCloudsResourceManager(t *testing.T) {
	for _, tc := range []struct {
		cloud, wantAuthority, wantScope string
	}{
		{"public", "login.microsoftonline.com", "https://management.core.windows.net//.default"},
		{"china", "login.chinacloudapi.cn", "https://management.core.chinacloudapi.cn//.default"},
		{"usgov", "login.microsoftonline.us", "https://management.core.usgovcloudapi.net//.default"},
	} {
		t.Run(tc.cloud, func(t *testing.T) {
			standin := newAzureStandin(t)
			creds := standinCredentials(tc.cloud)
			client, err := creds.newAzureClient(standin.redirectingTransport())
			require.NoError(t, err)

			token, err := client.kuduToken(context.Background())

			require.NoError(t, err)
			require.Equal(t, standinToken, token)
			tokenRequests := standin.tokenRequests()
			require.Len(t, tokenRequests, 1)
			require.Equal(t, tc.wantAuthority, tokenRequests[0].Host)
			// MSAL appends its own OIDC scopes.
			require.Contains(t, strings.Fields(tokenRequests[0].Form.Get("scope")), tc.wantScope)
			require.Equal(t, "client_credentials", tokenRequests[0].Form.Get("grant_type"))
		})
	}
}

func siteWithHostNames(name string, hostNames map[string]armappservice.HostType) *armappservice.Site {
	kind := "app"
	site := &armappservice.Site{Name: &name, Kind: &kind, Properties: &armappservice.SiteProperties{}}
	for host, hostType := range hostNames {
		site.Properties.HostNameSSLStates = append(site.Properties.HostNameSSLStates,
			&armappservice.HostNameSSLState{Name: to.Ptr(host), HostType: to.Ptr(hostType)})
	}
	return site
}

func TestKuduHostIsTheSitesRepositoryHostName(t *testing.T) {
	site := siteWithHostNames("myapp", map[string]armappservice.HostType{
		"myapp.chinacloudsites.cn":     armappservice.HostTypeStandard,
		"myapp.scm.chinacloudsites.cn": armappservice.HostTypeRepository,
	})

	host, err := kuduHost(site)

	require.NoError(t, err)
	require.Equal(t, "myapp.scm.chinacloudsites.cn", host)
}

func TestKuduHostFailsWithoutARepositoryHostName(t *testing.T) {
	for name, site := range map[string]*armappservice.Site{
		"only a standard host name": siteWithHostNames("myapp", map[string]armappservice.HostType{
			"myapp.azurewebsites.net": armappservice.HostTypeStandard,
		}),
		"no host names": siteWithHostNames("myapp", nil),
		"no properties": {Name: to.Ptr("myapp")},
		"an empty name": siteWithHostNames("myapp", map[string]armappservice.HostType{"": armappservice.HostTypeRepository}),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := kuduHost(site)
			require.EqualError(t, err, "app [myapp] lists no Repository host name, so its deployed package cannot be downloaded from Kudu")
		})
	}
}

func TestFingerprintZipServiceDownloadsFromTheSitesKuduHost(t *testing.T) {
	standin := newAzureStandin(t)
	standin.zips["myapp.scm.chinacloudsites.cn"] = zipOf(t, map[string]string{"index.js": "console.log('hi')"})
	creds := standinCredentials("china")
	client, err := creds.newAzureClient(standin.redirectingTransport())
	require.NoError(t, err)
	site := siteWithHostNames("myapp", map[string]armappservice.HostType{
		"myapp.scm.chinacloudsites.cn": armappservice.HostTypeRepository,
	})

	appData, err := client.fingerprintZipService(site, logger.NewStandardLogger())

	require.NoError(t, err)
	require.Regexp(t, `^[0-9a-f]{64}$`, appData.Digests["myapp"])
	var kuduHosts []string
	for _, r := range standin.received() {
		if r.Path == "/api/zip/site/wwwroot/" {
			kuduHosts = append(kuduHosts, r.Host)
			require.Equal(t, "Bearer "+standinToken, r.Authorization)
		}
	}
	require.Equal(t, []string{"myapp.scm.chinacloudsites.cn"}, kuduHosts)
}

// TestGetAzureAppsDataRunsAgainstAStandinNamedByTheOverrides is the path a
// stand-in takes in another process: nothing is redirected, so every request
// reaches the stand-in only because the overrides name it.
func TestGetAzureAppsDataRunsAgainstAStandinNamedByTheOverrides(t *testing.T) {
	standin := newAzureStandin(t)
	standin.zips[standin.host()] = zipOf(t, map[string]string{"index.js": "console.log('hi')"})
	standin.sites = []map[string]any{{
		"name": "myapp",
		"kind": "app",
		"properties": map[string]any{
			"state":             "Running",
			"siteConfig":        map[string]any{"linuxFxVersion": ""},
			"hostNameSslStates": []map[string]any{{"name": standin.host(), "hostType": "Repository"}},
		},
	}}
	creds := standinCredentials("public")
	creds.AuthorityHost = standin.srv.URL
	creds.ResourceManagerURL = standin.srv.URL

	appsData, err := creds.getAzureAppsData(nil, standin.srv.Client(), logger.NewStandardLogger())

	require.NoError(t, err)
	require.Len(t, appsData, 1)
	require.Equal(t, "myapp", appsData[0].AppName)
	require.Regexp(t, `^[0-9a-f]{64}$`, appsData[0].Digests["myapp"])
	for _, r := range standin.received() {
		require.Equal(t, standin.host(), r.Host)
	}
	require.NotEmpty(t, standin.tokenRequests())
}
