package azure

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestPlanImageFingerprint(t *testing.T) {
	sha := strings.Repeat("a", 64)

	for _, tc := range []struct {
		name        string
		imageName   string
		want        fingerprintPlan
		wantErrText string
	}{
		{
			name:      "acr image with a tag authenticates to acr",
			imageName: "myregistry.azurecr.io/myrepo/myapp:1.0",
			want: fingerprintPlan{
				source: fingerprintFromACR, domain: "myregistry.azurecr.io",
				reference: "myregistry.azurecr.io/myrepo/myapp:1.0",
				repoPath:  "myrepo/myapp", tagOrDigest: "1.0",
			},
		},
		{
			// Regression: the digest must keep its algorithm prefix, or ACR reads
			// the bare hex as a tag and returns MANIFEST_UNKNOWN.
			name:      "acr image pinned to a digest keeps the sha256 prefix",
			imageName: "myregistry.azurecr.io/myapp@sha256:" + sha,
			want: fingerprintPlan{
				source: fingerprintFromACR, domain: "myregistry.azurecr.io",
				reference: "myregistry.azurecr.io/myapp@sha256:" + sha,
				repoPath:  "myapp", tagOrDigest: "sha256:" + sha,
				pinnedFingerprint: sha,
			},
		},
		{
			// containers/image refuses a reference holding both, so the tag is
			// dropped and the digest wins.
			name:      "tag and digest together drops the tag",
			imageName: "ghcr.io/owner/app:v1@sha256:" + sha,
			want: fingerprintPlan{
				source: fingerprintFromAnonymousRegistry, domain: "ghcr.io",
				reference: "ghcr.io/owner/app@sha256:" + sha,
				repoPath:  "owner/app", tagOrDigest: "sha256:" + sha,
				pinnedFingerprint: sha,
			},
		},
		{
			name:      "image without a tag defaults to latest",
			imageName: "myregistry.azurecr.io/myapp",
			want: fingerprintPlan{
				source: fingerprintFromACR, domain: "myregistry.azurecr.io",
				reference: "myregistry.azurecr.io/myapp:latest",
				repoPath:  "myapp", tagOrDigest: "latest",
			},
		},
		{
			name:      "acr host with a port authenticates to acr",
			imageName: "myregistry.azurecr.io:443/myapp:v1",
			want: fingerprintPlan{
				source: fingerprintFromACR, domain: "myregistry.azurecr.io:443",
				reference: "myregistry.azurecr.io:443/myapp:v1",
				repoPath:  "myapp", tagOrDigest: "v1",
			},
		},
		{
			name:      "third party registry resolves anonymously",
			imageName: "ghcr.io/owner/app:v2",
			want: fingerprintPlan{
				source: fingerprintFromAnonymousRegistry, domain: "ghcr.io",
				reference: "ghcr.io/owner/app:v2", repoPath: "owner/app", tagOrDigest: "v2",
			},
		},
		{
			name:      "attacker controlled host resolves anonymously",
			imageName: "attacker.example/repo:latest",
			want: fingerprintPlan{
				source: fingerprintFromAnonymousRegistry, domain: "attacker.example",
				reference: "attacker.example/repo:latest", repoPath: "repo", tagOrDigest: "latest",
			},
		},
		{
			name:      "acr lookalike host resolves anonymously",
			imageName: "azurecr.io.attacker.example/repo:latest",
			want: fingerprintPlan{
				source: fingerprintFromAnonymousRegistry, domain: "azurecr.io.attacker.example",
				reference: "azurecr.io.attacker.example/repo:latest", repoPath: "repo", tagOrDigest: "latest",
			},
		},
		{
			name:      "docker hub short form is normalised",
			imageName: "nginx:latest",
			want: fingerprintPlan{
				source: fingerprintFromAnonymousRegistry, domain: "docker.io",
				reference: "docker.io/library/nginx:latest", repoPath: "library/nginx", tagOrDigest: "latest",
			},
		},
		{
			name:      "docker hub user image is normalised",
			imageName: "myuser/myimage:tag",
			want: fingerprintPlan{
				source: fingerprintFromAnonymousRegistry, domain: "docker.io",
				reference: "docker.io/myuser/myimage:tag", repoPath: "myuser/myimage", tagOrDigest: "tag",
			},
		},
		// Regression: a registry component a suffix check reads as ACR but a URL
		// parser resolves elsewhere must be rejected, or the Azure credential is
		// handed to that other host.
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
		{
			name:        "trailing dot fqdn is rejected by the parser",
			imageName:   "myregistry.azurecr.io./myapp:v1",
			wantErrText: "failed to parse the image name",
		},
		{
			name:        "a non sha256 pin cannot produce a kosli fingerprint",
			imageName:   "ghcr.io/owner/app@sha512:" + strings.Repeat("c", 128),
			wantErrText: "pinned to a digest Kosli cannot use",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := planImageFingerprint(tc.imageName)
			if tc.wantErrText != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrText)
				require.Equal(t, fingerprintPlan{}, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
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
		// The suffix must end the host, not sit inside it.
		{host: "azurecr.io.attacker.example", want: false},
		{host: "myregistry.azurecr.io.attacker.example", want: false},
		{host: "notazurecr.io", want: false},
		{host: "ghcr.io", want: false},
		{host: "registry-1.docker.io", want: false},
		{host: "mcr.microsoft.com", want: false},
		{host: "169.254.169.254", want: false},
		{host: "attacker.example:8443", want: false},
		// Domains reference.Domain can produce for a local registry.
		{host: "localhost:5000", want: false},
		{host: "[::1]:5000", want: false},
		{host: "[2001:db8::1]", want: false},
		{host: "", want: false},
	} {
		t.Run(tc.host, func(t *testing.T) {
			require.Equal(t, tc.want, isACRLoginServer(tc.host))
		})
	}
}

// stubAnonymousFingerprint replaces the anonymous resolver for one test and
// records the reference it was handed.
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

// TestGetImageFingerprintRoutesNonACRHostsAnonymously covers the dispatch and
// asserts the exact reference handed to the resolver, which a wiring bug would
// break.
func TestGetImageFingerprintRoutesNonACRHostsAnonymously(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId: "00000000-0000-0000-0000-000000000000", DigestsSource: "acr",
	}}
	sha := strings.Repeat("a", 64)

	for _, tc := range []struct{ name, imageName, wantRef string }{
		{"third party registry", "ghcr.io/owner/app:v2", "ghcr.io/owner/app:v2"},
		{"attacker controlled host", "attacker.example/repo:latest", "attacker.example/repo:latest"},
		{"docker hub short form", "nginx:latest", "docker.io/library/nginx:latest"},
		{"docker hub user image", "myuser/myimage:tag", "docker.io/myuser/myimage:tag"},
		{"pinned on a non acr host", "ghcr.io/owner/app@sha256:" + sha, "ghcr.io/owner/app@sha256:" + sha},
		{"tag and digest drops the tag", "ghcr.io/owner/app:v1@sha256:" + sha, "ghcr.io/owner/app@sha256:" + sha},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := stubAnonymousFingerprint(t, sha, nil)

			fingerprint, err := client.GetImageFingerprint(tc.imageName, logger.NewStandardLogger())
			require.NoError(t, err)
			require.Equal(t, sha, fingerprint)
			require.Equal(t, tc.wantRef, *got, "the canonical reference must reach the resolver")
		})
	}
}

// TestGetImageFingerprintUsesACRForACRHost asserts the other arm. An empty
// tenant id makes the Azure credential fail to construct, so the ACR arm returns
// before any network request.
func TestGetImageFingerprintUsesACRForACRHost(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{TenantId: "", DigestsSource: "acr"}}

	got := stubAnonymousFingerprint(t, strings.Repeat("a", 64), nil)

	_, err := client.GetImageFingerprint("myregistry.azurecr.io/app:v1", logger.NewStandardLogger())
	require.Error(t, err)
	require.Empty(t, *got, "an ACR host must not be resolved anonymously")
	require.NotContains(t, err.Error(), "--digests-source logs",
		"the error must come from the ACR arm, not the anonymous one")
}

func TestGetImageFingerprintRejectsSmuggledACRHost(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId: "00000000-0000-0000-0000-000000000000", DigestsSource: "acr",
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
// reporting a digest other than the one the reference pinned. Neither resolver
// checks this, so it is checked here.
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

func TestAnonymousImageFingerprintWrapsTheUnderlyingError(t *testing.T) {
	sentinel := errors.New("registry unreachable")
	stubAnonymousFingerprint(t, "", sentinel)

	plan := fingerprintPlan{reference: "ghcr.io/owner/app:v2", domain: "ghcr.io"}
	_, err := anonymousImageFingerprint(plan, logger.NewStandardLogger())
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Contains(t, err.Error(), "--digests-source logs")
}

// fakeACR answers the manifest request with a chosen Docker-Content-Digest, or
// omits the header entirely when contentDigest is empty.
func fakeACR(t *testing.T, contentDigest string) (fingerprintPlan, *azcontainerregistry.ClientOptions) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contentDigest != "" {
			w.Header().Set("Docker-Content-Digest", contentDigest)
		}
		w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"schemaVersion":2}`))
	}))
	t.Cleanup(srv.Close)

	plan := fingerprintPlan{
		source: fingerprintFromACR, domain: strings.TrimPrefix(srv.URL, "https://"),
		reference: "fake/app:v1", repoPath: "app", tagOrDigest: "v1",
	}
	options := &azcontainerregistry.ClientOptions{
		ClientOptions: azcore.ClientOptions{Transport: srv.Client()},
	}
	return plan, options
}

// TestACRImageFingerprintRejectsUnusableDigests covers the ACR arm's own error
// branches, which have no coverage otherwise because the client talks to a
// registry. The digest rule itself lives in internal/digest; this asserts the
// arm is actually wired to it.
func TestACRImageFingerprintRejectsUnusableDigests(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId:     "00000000-0000-0000-0000-000000000000",
		ClientId:     "00000000-0000-0000-0000-000000000000",
		ClientSecret: "not-a-real-secret",
	}}

	for _, tc := range []struct {
		name          string
		contentDigest string
		wantErrText   string
	}{
		{name: "sha512 digest", contentDigest: "sha512:" + strings.Repeat("c", 128), wantErrText: "algorithm is sha512"},
		{name: "sha384 digest", contentDigest: "sha384:" + strings.Repeat("b", 96), wantErrText: "algorithm is sha384"},
		{name: "unparseable digest", contentDigest: "not-a-digest", wantErrText: "unparseable digest"},
		{name: "missing digest header", contentDigest: "", wantErrText: "no digest returned"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan, options := fakeACR(t, tc.contentDigest)

			fingerprint, err := client.acrImageFingerprint(plan, options, logger.NewStandardLogger())

			require.Error(t, err)
			require.Empty(t, fingerprint)
			require.Contains(t, err.Error(), tc.wantErrText)
		})
	}
}

func TestACRImageFingerprintReturnsTheSha256Hex(t *testing.T) {
	client := &AzureClient{Credentials: AzureStaticCredentials{
		TenantId:     "00000000-0000-0000-0000-000000000000",
		ClientId:     "00000000-0000-0000-0000-000000000000",
		ClientSecret: "not-a-real-secret",
	}}
	want := strings.Repeat("a", 64)
	plan, options := fakeACR(t, "sha256:"+want)

	fingerprint, err := client.acrImageFingerprint(plan, options, logger.NewStandardLogger())

	require.NoError(t, err)
	require.Equal(t, want, fingerprint)
}
