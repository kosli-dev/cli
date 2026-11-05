package azure

import (
	"context"
	"strings"
	"testing"

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
