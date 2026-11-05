package azure

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	standinTenant       = "00000000-0000-0000-0000-000000000001"
	standinSubscription = "00000000-0000-0000-0000-000000000002"
	standinGroup        = "standin-group"
	standinToken        = "standin-access-token"
)

// standinRequest is one request the stand-in received, with the host the
// client addressed rather than the one it was redirected to.
type standinRequest struct {
	Method, Host, Path, Authorization string
	Form                              url.Values
}

// azureStandin answers as the Entra ID authority, Resource Manager and Kudu for
// one resource group, whatever host each request was addressed to.
type azureStandin struct {
	srv *httptest.Server
	// sites is the JSON of each site the resource group lists.
	sites []map[string]any
	// zips holds the package Kudu serves at each host.
	zips map[string][]byte

	mu       sync.Mutex
	requests []standinRequest
}

func newAzureStandin(t *testing.T) *azureStandin {
	t.Helper()
	s := &azureStandin{zips: map[string][]byte{}}
	s.srv = httptest.NewTLSServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	return s
}

// host is the stand-in's own address, for a test that points the client at it
// directly.
func (s *azureStandin) host() string {
	return strings.TrimPrefix(s.srv.URL, "https://")
}

// redirectingTransport sends every request to the stand-in, so a client
// configured for a real cloud can be driven against it.
func (s *azureStandin) redirectingTransport() *http.Client {
	var asked []string
	return &http.Client{Transport: hostRewritingTransport{
		addr: s.host(), inner: s.srv.Client().Transport, hostsAsked: &asked,
	}}
}

func (s *azureStandin) received() []standinRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]standinRequest(nil), s.requests...)
}

// tokenRequests are the requests made to the token endpoint.
func (s *azureStandin) tokenRequests() []standinRequest {
	var got []standinRequest
	for _, r := range s.received() {
		if strings.HasSuffix(r.Path, "/oauth2/v2.0/token") {
			got = append(got, r)
		}
	}
	return got
}

func (s *azureStandin) serve(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	s.mu.Lock()
	s.requests = append(s.requests, standinRequest{
		Method: r.Method, Host: r.Host, Path: r.URL.Path,
		Authorization: r.Header.Get("Authorization"), Form: r.PostForm,
	})
	s.mu.Unlock()

	authority := "https://" + r.Host + "/" + standinTenant
	sitesPath := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Web/sites", standinSubscription, standinGroup)

	switch r.URL.Path {
	case "/common/discovery/instance":
		// Entra answers for the authority being asked about, not for itself.
		asked, err := url.Parse(r.URL.Query().Get("authorization_endpoint"))
		if err != nil || asked.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{
			"tenant_discovery_endpoint": "https://" + asked.Host + "/" + standinTenant + "/v2.0/.well-known/openid-configuration",
			"metadata": []map[string]any{{
				"preferred_network": asked.Host, "preferred_cache": asked.Host, "aliases": []string{asked.Host},
			}},
		})
	case "/" + standinTenant + "/v2.0/.well-known/openid-configuration":
		writeJSON(w, map[string]any{
			"token_endpoint":         authority + "/oauth2/v2.0/token",
			"authorization_endpoint": authority + "/oauth2/v2.0/authorize",
			"issuer":                 authority + "/v2.0",
		})
	case "/" + standinTenant + "/oauth2/v2.0/token":
		writeJSON(w, map[string]any{"access_token": standinToken, "token_type": "Bearer", "expires_in": 3600})
	case sitesPath:
		if r.Header.Get("Authorization") != "Bearer "+standinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]any{"value": s.sites})
	case "/api/zip/site/wwwroot/":
		if r.Header.Get("Authorization") != "Bearer "+standinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		pkg, ok := s.zips[r.Host]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write(pkg)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func standinCredentials(cloudName string) AzureStaticCredentials {
	return AzureStaticCredentials{
		TenantId:          standinTenant,
		ClientId:          "00000000-0000-0000-0000-000000000003",
		ClientSecret:      "standin-secret",
		SubscriptionId:    standinSubscription,
		ResourceGroupName: standinGroup,
		Cloud:             cloudName,
	}
}

// zipOf is a package holding one file per entry of files.
func zipOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := zw.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}
