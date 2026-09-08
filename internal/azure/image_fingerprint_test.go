package azure

import (
	"errors"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestParseImageReference(t *testing.T) {
	validSha256 := strings.Repeat("a", 64)

	for _, tc := range []struct {
		name        string
		imageName   string
		domain      string
		path        string
		canonical   string
		pinnedSha   string
		wantErrText string
	}{
		{
			name:      "acr image with a tag",
			imageName: "myregistry.azurecr.io/myrepo/myapp:1.0",
			domain:    "myregistry.azurecr.io",
			path:      "myrepo/myapp",
			canonical: "myregistry.azurecr.io/myrepo/myapp:1.0",
		},
		{
			name:      "acr image pinned to a digest",
			imageName: "myregistry.azurecr.io/myapp@sha256:" + validSha256,
			domain:    "myregistry.azurecr.io",
			path:      "myapp",
			canonical: "myregistry.azurecr.io/myapp@sha256:" + validSha256,
			pinnedSha: validSha256,
		},
		{
			name:      "image without a tag defaults to latest",
			imageName: "myregistry.azurecr.io/myapp",
			domain:    "myregistry.azurecr.io",
			path:      "myapp",
			canonical: "myregistry.azurecr.io/myapp:latest",
		},
		{
			name:      "third party registry",
			imageName: "ghcr.io/owner/app:v2",
			domain:    "ghcr.io",
			path:      "owner/app",
			canonical: "ghcr.io/owner/app:v2",
		},
		{
			name:      "host with a port",
			imageName: "registry.example.com:8443/app:v1",
			domain:    "registry.example.com:8443",
			path:      "app",
			canonical: "registry.example.com:8443/app:v1",
		},
		{
			name:      "docker hub short form is normalised",
			imageName: "nginx:latest",
			domain:    "docker.io",
			path:      "library/nginx",
			canonical: "docker.io/library/nginx:latest",
		},
		{
			name:      "docker hub user image is normalised",
			imageName: "myuser/myimage:tag",
			domain:    "docker.io",
			path:      "myuser/myimage",
			canonical: "docker.io/myuser/myimage:tag",
		},
		// Regression: a registry component that a suffix check reads as ACR but a
		// URL parser resolves to a different host must be rejected outright, or the
		// Azure credential is handed to that other host.
		{
			name:        "acr host smuggled into userinfo with a port is rejected",
			imageName:   "myregistry.azurecr.io:443@attacker.example/repo:tag",
			wantErrText: "failed to parse the image name",
		},
		{
			name:        "acr host smuggled into userinfo without a numeric port is rejected",
			imageName:   "myregistry.azurecr.io:x@attacker.example/repo:tag",
			wantErrText: "failed to parse the image name",
		},
		{
			name:        "acr host smuggled into userinfo with no port is rejected",
			imageName:   "myregistry.azurecr.io@attacker.example/repo:tag",
			wantErrText: "failed to parse the image name",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			domain, path, canonical, pinned, err := parseImageReference(tc.imageName)
			if tc.wantErrText != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrText)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.domain, domain)
			require.Equal(t, tc.path, path)
			require.Equal(t, tc.canonical, canonical)
			require.Equal(t, tc.pinnedSha, pinned)
		})
	}
}

// TestGetImageFingerprintRejectsSmuggledACRHost is the regression test for the
// bypass: the credential must not reach a host that only looks like ACR.
func TestGetImageFingerprintRejectsSmuggledACRHost(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId:      "00000000-0000-0000-0000-000000000000",
		ClientId:      "00000000-0000-0000-0000-000000000000",
		ClientSecret:  "not-a-real-secret",
		DigestsSource: "acr",
	}}

	for _, imageName := range []string{
		"myregistry.azurecr.io:443@attacker.example/repo:tag",
		"myregistry.azurecr.io:x@attacker.example/repo:tag",
		"myregistry.azurecr.io@attacker.example/repo:tag",
	} {
		t.Run(imageName, func(t *testing.T) {
			got := stubAnonymousFingerprint(t, strings.Repeat("a", 64), nil)

			_, err := client.GetImageFingerprint(imageName, logger.NewStandardLogger())
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to parse the image name")
			require.Empty(t, *got, "must not resolve at all")
		})
	}
}

// TestGetImageFingerprintHoldsTheRegistryToAPinnedDigest stops a registry
// reporting a digest other than the one the reference pinned.
func TestGetImageFingerprintHoldsTheRegistryToAPinnedDigest(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{DigestsSource: "acr"}}
	pinned := strings.Repeat("a", 64)
	other := strings.Repeat("b", 64)

	stubAnonymousFingerprint(t, other, nil)
	_, err := client.GetImageFingerprint("ghcr.io/owner/app@sha256:"+pinned, logger.NewStandardLogger())
	require.Error(t, err)
	require.Contains(t, err.Error(), "is pinned to digest sha256:"+pinned)

	stubAnonymousFingerprint(t, pinned, nil)
	fingerprint, err := client.GetImageFingerprint("ghcr.io/owner/app@sha256:"+pinned, logger.NewStandardLogger())
	require.NoError(t, err)
	require.Equal(t, pinned, fingerprint)
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

// TestClassifyImageReferenceKeepsAzureCredentialForACROnly is the regression
// test for the credential-forwarding vulnerability: the Azure credential is
// attached only on the fingerprintFromACR path, so any host that is not an ACR
// login server must classify as something else.
func TestClassifyImageReferenceKeepsAzureCredentialForACROnly(t *testing.T) {
	for _, tc := range []struct {
		name         string
		registryHost string
		want         imageFingerprintSource
	}{
		{
			name:         "acr host authenticates to acr",
			registryHost: "myregistry.azurecr.io",
			want:         fingerprintFromACR,
		},
		{
			name:         "acr host written as an fqdn still authenticates to acr",
			registryHost: "myregistry.azurecr.io.",
			want:         fingerprintFromACR,
		},
		{
			name:         "attacker controlled host is resolved anonymously",
			registryHost: "attacker.example",
			want:         fingerprintFromAnonymousRegistry,
		},
		{
			name:         "acr lookalike host is resolved anonymously",
			registryHost: "azurecr.io.attacker.example",
			want:         fingerprintFromAnonymousRegistry,
		},
		{
			name:         "third party registry is resolved anonymously",
			registryHost: "ghcr.io",
			want:         fingerprintFromAnonymousRegistry,
		},
		{
			name:         "instance metadata address is resolved anonymously",
			registryHost: "169.254.169.254",
			want:         fingerprintFromAnonymousRegistry,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, classifyImageReference(tc.registryHost))
		})
	}
}

// stubAnonymousFingerprint replaces the anonymous resolver for the duration of a
// test and records the reference it was handed.
func stubAnonymousFingerprint(t *testing.T, fingerprint string, err error) *string {
	t.Helper()
	original := anonymousFingerprint
	var got string
	anonymousFingerprint = func(imageName string) (string, error) {
		got = imageName
		return fingerprint, err
	}
	t.Cleanup(func() { anonymousFingerprint = original })
	return &got
}

// TestGetImageFingerprintRoutesNonACRHostsAnonymously covers the dispatch in
// GetImageFingerprint, not just the classifier, so that swapping the arms would
// fail a test. It also asserts the exact reference handed to the resolver, which
// a wiring bug (passing repoName instead of the full image name) would break.
func TestGetImageFingerprintRoutesNonACRHostsAnonymously(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId:      "00000000-0000-0000-0000-000000000000",
		ClientId:      "00000000-0000-0000-0000-000000000000",
		ClientSecret:  "not-a-real-secret",
		DigestsSource: "acr",
	}}
	validSha256 := strings.Repeat("a", 64)

	for _, tc := range []struct {
		name      string
		imageName string
		wantRef   string
	}{
		{
			name:      "third party registry",
			imageName: "ghcr.io/owner/app:v2",
			wantRef:   "ghcr.io/owner/app:v2",
		},
		{
			name:      "attacker controlled host",
			imageName: "attacker.example/repo:latest",
			wantRef:   "attacker.example/repo:latest",
		},
		{
			name:      "docker hub short form is normalised before resolving",
			imageName: "nginx:latest",
			wantRef:   "docker.io/library/nginx:latest",
		},
		{
			name:      "docker hub user image is normalised before resolving",
			imageName: "myuser/myimage:tag",
			wantRef:   "docker.io/myuser/myimage:tag",
		},
		{
			name:      "digest pinned reference on a non acr host",
			imageName: "ghcr.io/owner/app@sha256:" + validSha256,
			wantRef:   "ghcr.io/owner/app@sha256:" + validSha256,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := stubAnonymousFingerprint(t, validSha256, nil)

			fingerprint, err := client.GetImageFingerprint(tc.imageName, logger.NewStandardLogger())
			require.NoError(t, err)
			require.Equal(t, validSha256, fingerprint)
			require.Equal(t, tc.wantRef, *got, "the full image reference must reach the anonymous resolver")
		})
	}
}

// TestGetImageFingerprintUsesACRForACRHost asserts the other arm of the
// dispatch: an ACR host must not be resolved anonymously. An empty tenant id
// makes the Azure credential fail to construct, so the ACR arm returns before
// any network request and the test needs no registry.
func TestGetImageFingerprintUsesACRForACRHost(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId:      "",
		ClientId:      "00000000-0000-0000-0000-000000000000",
		ClientSecret:  "not-a-real-secret",
		DigestsSource: "acr",
	}}

	got := stubAnonymousFingerprint(t, strings.Repeat("a", 64), nil)

	_, err := client.GetImageFingerprint("myregistry.azurecr.io/app:v1", logger.NewStandardLogger())
	require.Error(t, err)
	require.Empty(t, *got, "an ACR host must not be resolved anonymously")
	require.NotContains(t, err.Error(), "--digests-source logs",
		"the error must come from the ACR arm, not the anonymous one")
}

// TestAnonymousImageFingerprintWrapsTheUnderlyingError keeps the error
// wrapped so callers can inspect it, and keeps the actionable hint.
func TestAnonymousImageFingerprintWrapsTheUnderlyingError(t *testing.T) {
	sentinel := errors.New("registry unreachable")
	stubAnonymousFingerprint(t, "", sentinel)

	_, err := anonymousImageFingerprint("ghcr.io/owner/app:v2", "ghcr.io", logger.NewStandardLogger())
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Contains(t, err.Error(), "--digests-source logs")
}
