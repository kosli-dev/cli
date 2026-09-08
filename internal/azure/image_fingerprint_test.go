package azure

import (
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestParseImageName(t *testing.T) {
	for _, tc := range []struct {
		name        string
		imageName   string
		registryUrl string
		repoName    string
		tag         string
	}{
		{
			name:        "acr image with a tag",
			imageName:   "myregistry.azurecr.io/myrepo/myapp:1.0",
			registryUrl: "https://myregistry.azurecr.io",
			repoName:    "myrepo/myapp",
			tag:         "1.0",
		},
		{
			name:        "acr image pinned to a digest",
			imageName:   "myregistry.azurecr.io/myapp@sha256:" + strings.Repeat("a", 64),
			registryUrl: "https://myregistry.azurecr.io",
			repoName:    "myapp",
			tag:         "sha256:" + strings.Repeat("a", 64),
		},
		{
			name:        "image without a tag defaults to latest",
			imageName:   "myregistry.azurecr.io/myapp",
			registryUrl: "https://myregistry.azurecr.io",
			repoName:    "myapp",
			tag:         "latest",
		},
		{
			name:        "third party registry",
			imageName:   "ghcr.io/owner/app:v2",
			registryUrl: "https://ghcr.io",
			repoName:    "owner/app",
			tag:         "v2",
		},
		{
			name:        "host with a port",
			imageName:   "registry.example.com:8443/app:v1",
			registryUrl: "https://registry.example.com:8443",
			repoName:    "app",
			tag:         "v1",
		},
		{
			name:        "no registry host means nothing is parsed",
			imageName:   "nginx:latest",
			registryUrl: "",
			repoName:    "",
			tag:         "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registryUrl, repoName, tag := parseImageName(tc.imageName)
			require.Equal(t, tc.registryUrl, registryUrl)
			require.Equal(t, tc.repoName, repoName)
			require.Equal(t, tc.tag, tag)
		})
	}
}

func TestIsACRLoginServer(t *testing.T) {
	for _, tc := range []struct {
		host string
		want bool
	}{
		{host: "myregistry.azurecr.io", want: true},
		{host: "MyRegistry.AzureCR.IO", want: true},
		{host: "myregistry.azurecr.cn", want: true},
		{host: "myregistry.azurecr.us", want: true},
		{host: "myregistry.azurecr.io:443", want: true},
		// A suffix on its own names no registry.
		{host: "azurecr.io", want: false},
		{host: ".azurecr.io", want: false},
		// The suffix must be the end of the host, not a label inside it.
		{host: "azurecr.io.attacker.example", want: false},
		{host: "myregistry.azurecr.io.attacker.example", want: false},
		// Nor may it merely appear in a longer label.
		{host: "notazurecr.io", want: false},
		{host: "ghcr.io", want: false},
		{host: "registry-1.docker.io", want: false},
		{host: "mcr.microsoft.com", want: false},
		{host: "169.254.169.254", want: false},
		{host: "attacker.example:8443", want: false},
		{host: "", want: false},
	} {
		t.Run(tc.host, func(t *testing.T) {
			require.Equal(t, tc.want, isACRLoginServer(tc.host))
		})
	}
}

func TestPinnedDigest(t *testing.T) {
	validSha256 := strings.Repeat("a", 64)

	fingerprint, pinned := pinnedDigest("sha256:" + validSha256)
	require.True(t, pinned)
	require.Equal(t, validSha256, fingerprint)

	for _, tag := range []string{
		"latest",
		"1.0",
		"sha256:tooshort",
		"sha256:" + strings.Repeat("z", 64), // 64 chars but not hex
		"sha256:",
		"",
	} {
		t.Run(tag, func(t *testing.T) {
			_, pinned := pinnedDigest(tag)
			require.False(t, pinned)
		})
	}
}

// TestClassifyImageReferenceKeepsAzureCredentialForACROnly is the regression
// test for the credential-forwarding vulnerability: the Azure credential is
// attached only on the fingerprintFromACR path, so any host that is not an ACR
// login server must classify as something else.
func TestClassifyImageReferenceKeepsAzureCredentialForACROnly(t *testing.T) {
	validSha256 := strings.Repeat("a", 64)

	for _, tc := range []struct {
		name         string
		registryHost string
		tag          string
		want         imageFingerprintSource
	}{
		{
			name:         "acr host with a tag authenticates to acr",
			registryHost: "myregistry.azurecr.io",
			tag:          "1.0",
			want:         fingerprintFromACR,
		},
		{
			name:         "attacker controlled host is resolved anonymously",
			registryHost: "attacker.example",
			tag:          "1.0",
			want:         fingerprintFromAnonymousRegistry,
		},
		{
			name:         "acr lookalike host is resolved anonymously",
			registryHost: "azurecr.io.attacker.example",
			tag:          "1.0",
			want:         fingerprintFromAnonymousRegistry,
		},
		{
			name:         "third party registry is resolved anonymously",
			registryHost: "ghcr.io",
			tag:          "v2",
			want:         fingerprintFromAnonymousRegistry,
		},
		{
			name:         "instance metadata address is resolved anonymously",
			registryHost: "169.254.169.254",
			tag:          "latest",
			want:         fingerprintFromAnonymousRegistry,
		},
		{
			name:         "a pinned digest needs no registry at all",
			registryHost: "myregistry.azurecr.io",
			tag:          "sha256:" + validSha256,
			want:         fingerprintFromPinnedDigest,
		},
		{
			name:         "a pinned digest on an attacker host needs no registry either",
			registryHost: "attacker.example",
			tag:          "sha256:" + validSha256,
			want:         fingerprintFromPinnedDigest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, classifyImageReference(tc.registryHost, tc.tag))
		})
	}
}

// TestGetImageFingerprintDoesNotAuthenticateToNonACRHost asserts the whole
// resolver, not just the classifier: a non-ACR host must not reach Azure
// authentication. The Azure credentials here are deliberately invalid, so an
// attempt to use them would surface as an Azure credential error.
func TestGetImageFingerprintDoesNotAuthenticateToNonACRHost(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId:      "00000000-0000-0000-0000-000000000000",
		ClientId:      "00000000-0000-0000-0000-000000000000",
		ClientSecret:  "not-a-real-secret",
		DigestsSource: "acr",
	}}

	// A closed local port, so the lookup fails immediately and reaches no registry.
	_, err := client.GetImageFingerprint("127.0.0.1:1/owner/app:v1", logger.NewStandardLogger())
	require.Error(t, err)
	require.Contains(t, err.Error(), "Azure credentials are only sent to Azure Container Registry")
}

func TestGetImageFingerprintRejectsImageWithoutRegistryHost(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{}}
	_, err := client.GetImageFingerprint("nginx:latest", logger.NewStandardLogger())
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not name a registry host")
}
