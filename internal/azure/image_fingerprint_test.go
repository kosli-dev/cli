package azure

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"

	"github.com/kosli-dev/cli/internal/digest"
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
				domain:    "myregistry.azurecr.io",
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
				domain:    "myregistry.azurecr.io",
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
				domain:    "ghcr.io",
				reference: "ghcr.io/owner/app@sha256:" + sha,
				repoPath:  "owner/app", tagOrDigest: "sha256:" + sha,
				pinnedFingerprint: sha,
			},
		},
		{
			name:      "image without a tag defaults to latest",
			imageName: "myregistry.azurecr.io/myapp",
			want: fingerprintPlan{
				domain:    "myregistry.azurecr.io",
				reference: "myregistry.azurecr.io/myapp:latest",
				repoPath:  "myapp", tagOrDigest: "latest",
			},
		},
		{
			name:      "acr host with a port authenticates to acr",
			imageName: "myregistry.azurecr.io:443/myapp:v1",
			want: fingerprintPlan{
				domain:    "myregistry.azurecr.io:443",
				reference: "myregistry.azurecr.io:443/myapp:v1",
				repoPath:  "myapp", tagOrDigest: "v1",
			},
		},
		{
			name:      "third party registry resolves anonymously",
			imageName: "ghcr.io/owner/app:v2",
			want: fingerprintPlan{
				domain:    "ghcr.io",
				reference: "ghcr.io/owner/app:v2", repoPath: "owner/app", tagOrDigest: "v2",
			},
		},
		{
			name:      "attacker controlled host resolves anonymously",
			imageName: "attacker.example/repo:latest",
			want: fingerprintPlan{
				domain:    "attacker.example",
				reference: "attacker.example/repo:latest", repoPath: "repo", tagOrDigest: "latest",
			},
		},
		{
			name:      "acr lookalike host resolves anonymously",
			imageName: "azurecr.io.attacker.example/repo:latest",
			want: fingerprintPlan{
				domain:    "azurecr.io.attacker.example",
				reference: "azurecr.io.attacker.example/repo:latest", repoPath: "repo", tagOrDigest: "latest",
			},
		},
		{
			name:      "docker hub short form is normalised",
			imageName: "nginx:latest",
			want: fingerprintPlan{
				domain:    "docker.io",
				reference: "docker.io/library/nginx:latest", repoPath: "library/nginx", tagOrDigest: "latest",
			},
		},
		{
			name:      "docker hub user image is normalised",
			imageName: "myuser/myimage:tag",
			want: fingerprintPlan{
				domain:    "docker.io",
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

	// A near miss, so a comparison of only part of the digest cannot pass.
	stubAnonymousFingerprint(t, strings.Repeat("a", 63)+"b", nil)
	_, err = client.GetImageFingerprint("ghcr.io/owner/app@sha256:"+pinned, logger.NewStandardLogger())
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

// hostRewritingTransport sends every request to addr regardless of the host in
// the URL, so a reference naming a real ACR login server can be resolved against
// a fake registry.
// It records the host it was asked for first, so a test can still assert which
// registry the client was pointed at even though the request is redirected.
type hostRewritingTransport struct {
	addr       string
	inner      http.RoundTripper
	hostsAsked *[]string
}

func (t hostRewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	*t.hostsAsked = append(*t.hostsAsked, req.URL.Host)
	rewritten := req.Clone(req.Context())
	rewritten.URL.Host = t.addr
	return t.inner.RoundTrip(rewritten)
}

// fakeACR stands up a registry that answers the manifest request with the given
// digest header and status, and records the paths it was asked for.
func fakeACR(t *testing.T, contentDigest string, status int) (*azcontainerregistry.ClientOptions, *[]string, *[]string) {
	t.Helper()
	var paths []string
	var hostsAsked []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// EscapedPath is what actually went on the wire; URL.Path is decoded.
		paths = append(paths, r.URL.EscapedPath())
		if contentDigest != "" {
			w.Header().Set("Docker-Content-Digest", contentDigest)
		}
		w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"schemaVersion":2}`))
	}))
	t.Cleanup(srv.Close)

	parsed, err := url.Parse(srv.URL)
	require.NoError(t, err)
	options := &azcontainerregistry.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Transport: &http.Client{Transport: hostRewritingTransport{
				addr: parsed.Host, inner: srv.Client().Transport, hostsAsked: &hostsAsked,
			}},
		},
	}
	return options, &paths, &hostsAsked
}

func acrTestClient(t *testing.T, options *azcontainerregistry.ClientOptions) *AzureClient {
	t.Helper()
	return &AzureClient{
		Credentials: AzureStaticCredentials{
			TenantId:     "00000000-0000-0000-0000-000000000000",
			ClientId:     "00000000-0000-0000-0000-000000000000",
			ClientSecret: "not-a-real-secret",
		},
		acrClientOptions: options,
	}
}

// TestGetImageFingerprintDrivesTheACRArmEndToEnd resolves an ACR reference all
// the way through GetImageFingerprint, which is what proves the parsed domain,
// repo path and tag actually reach the registry client. Without it, replacing
// plan.domain inside the arm with a hand-rolled split of the image name goes
// unnoticed.
func TestGetImageFingerprintDrivesTheACRArmEndToEnd(t *testing.T) {
	want := strings.Repeat("a", 64)
	options, paths, hostsAsked := fakeACR(t, "sha256:"+want, http.StatusOK)
	client := acrTestClient(t, options)

	fingerprint, err := client.GetImageFingerprint("myregistry.azurecr.io/team/app:v1", logger.NewStandardLogger())

	require.NoError(t, err)
	require.Equal(t, want, fingerprint)
	require.Contains(t, *paths, "/v2/team%2Fapp/manifests/v1",
		"the parsed repo path and tag must reach the registry, in that order")
	require.Contains(t, *hostsAsked, "myregistry.azurecr.io",
		"the client must be pointed at the domain the parser reported")
}

// TestGetImageFingerprintACRArmRequestsThePinnedDigest is the same for a pinned
// reference: the digest must reach the registry with its algorithm prefix.
func TestGetImageFingerprintACRArmRequestsThePinnedDigest(t *testing.T) {
	pinned := strings.Repeat("a", 64)
	options, paths, hostsAsked := fakeACR(t, "sha256:"+pinned, http.StatusOK)
	client := acrTestClient(t, options)

	fingerprint, err := client.GetImageFingerprint("myregistry.azurecr.io/app@sha256:"+pinned, logger.NewStandardLogger())

	require.NoError(t, err)
	require.Equal(t, pinned, fingerprint)
	require.Contains(t, *paths, "/v2/app/manifests/sha256:"+pinned,
		"a bare hex digest would be read as a tag by the registry")
	require.Contains(t, *hostsAsked, "myregistry.azurecr.io",
		"the client must be pointed at the domain the parser reported")
}

// TestGetImageFingerprintACRArmHoldsTheRegistryToAPinnedDigest exercises the
// cross-check on the credential-bearing arm, and with a digest differing in one
// character so a partial comparison cannot pass.
func TestGetImageFingerprintACRArmHoldsTheRegistryToAPinnedDigest(t *testing.T) {
	pinned := strings.Repeat("a", 64)
	nearMiss := strings.Repeat("a", 63) + "b"
	options, _, _ := fakeACR(t, "sha256:"+nearMiss, http.StatusOK)
	client := acrTestClient(t, options)

	_, err := client.GetImageFingerprint("myregistry.azurecr.io/app@sha256:"+pinned, logger.NewStandardLogger())

	require.Error(t, err)
	require.Contains(t, err.Error(), "is pinned to digest sha256:"+pinned)
	require.Contains(t, err.Error(), "reported sha256:"+nearMiss)
}

// TestGetImageFingerprintACRArmReportsRegistryErrors keeps the registry's own
// failure rather than degrading to the missing-header message.
func TestGetImageFingerprintACRArmReportsRegistryErrors(t *testing.T) {
	options, _, _ := fakeACR(t, "", http.StatusNotFound)
	client := acrTestClient(t, options)

	_, err := client.GetImageFingerprint("myregistry.azurecr.io/app:v1", logger.NewStandardLogger())

	require.Error(t, err)
	require.NotContains(t, err.Error(), "no digest returned",
		"a 404 must surface as the registry error, not as a missing digest header")
}

// TestGetImageFingerprintACRArmRejectsUnusableDigests covers the ACR arm's own
// error branches through the full path, which need a registry and so have no
// coverage otherwise.
func TestGetImageFingerprintACRArmRejectsUnusableDigests(t *testing.T) {
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
			options, _, _ := fakeACR(t, tc.contentDigest, http.StatusOK)
			client := acrTestClient(t, options)

			fingerprint, err := client.GetImageFingerprint("myregistry.azurecr.io/app:v1", logger.NewStandardLogger())

			require.Error(t, err)
			require.Empty(t, fingerprint)
			require.Contains(t, err.Error(), tc.wantErrText)
		})
	}
}

// TestAnonymousFingerprintIsTheCredentialFreeResolver pins what the variable
// points at in production. Every other test replaces it, so swapping it for the
// credential-discovering OciSha256 would otherwise go unnoticed.
func TestAnonymousFingerprintIsTheCredentialFreeResolver(t *testing.T) {
	require.Equal(t,
		reflect.ValueOf(digest.OciSha256Anonymous).Pointer(),
		reflect.ValueOf(anonymousFingerprint).Pointer(),
		"anonymousFingerprint must be digest.OciSha256Anonymous, not a resolver that discovers host credentials")
}

// TestFingerprintDockerServiceUsesTheACRSource covers the only production caller
// of GetImageFingerprint. Without it, inverting the digests-source condition, or
// replacing the resolver call with a constant, goes unnoticed.
func TestFingerprintDockerServiceUsesTheACRSource(t *testing.T) {
	want := strings.Repeat("a", 64)
	options, paths, _ := fakeACR(t, "sha256:"+want, http.StatusOK)
	client := acrTestClient(t, options)
	client.Credentials.DigestsSource = "acr"

	appName, appKind := "payments-api", "app"
	imageName := "myregistry.azurecr.io/team/app:v1"

	appData, err := client.fingerprintDockerService(
		&armappservice.Site{Name: &appName, Kind: &appKind}, logger.NewStandardLogger(), imageName)

	require.NoError(t, err)
	require.Equal(t, AppData{
		AppName:       appName,
		AppKind:       appKind,
		DigestsSource: "acr",
		Digests:       map[string]string{imageName: want},
		StartedAt:     0,
	}, appData)
	require.NotEmpty(t, *paths, "the acr source must actually contact the registry")
}

// TestFingerprintDockerServicePropagatesResolverErrors keeps the resolver's error
// rather than reporting an app with no fingerprint.
func TestFingerprintDockerServicePropagatesResolverErrors(t *testing.T) {
	options, _, _ := fakeACR(t, "sha512:"+strings.Repeat("c", 128), http.StatusOK)
	client := acrTestClient(t, options)
	client.Credentials.DigestsSource = "acr"

	appName, appKind := "payments-api", "app"

	_, err := client.fingerprintDockerService(
		&armappservice.Site{Name: &appName, Kind: &appKind}, logger.NewStandardLogger(),
		"myregistry.azurecr.io/team/app:v1")

	require.Error(t, err)
	require.Contains(t, err.Error(), "algorithm is sha512")
}
