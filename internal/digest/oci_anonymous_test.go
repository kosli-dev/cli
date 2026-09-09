package digest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/containers/image/v5/docker"

	"github.com/containers/image/v5/types"
	godigest "github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

// fakeRegistry answers the manifest HEAD with a chosen Docker-Content-Digest.
func fakeRegistry(t *testing.T, contentDigest string) string {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Docker-Content-Digest", contentDigest)
		w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	parsed, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return parsed.Host
}

// insecureAnonymousLookup mirrors OciSha256Anonymous against a fake TLS
// registry. Production does not skip TLS verification, so the flag is set here
// rather than in credentialContext.
func insecureAnonymousLookup(artifactName string) (string, error) {
	sysCtx := credentialContext(noCredentials, "", "")
	sysCtx.DockerInsecureSkipTLSVerify = types.OptionalBoolTrue

	ref, err := docker.ParseReference("//" + artifactName)
	if err != nil {
		return "", err
	}
	remoteDigest, err := docker.GetDigest(context.Background(), sysCtx, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get digest: %w", err)
	}
	return Sha256Fingerprint(remoteDigest)
}

// TestOciSha256RejectsNonSha256RegistryDigest covers a registry answering with
// an algorithm other than sha256. go-digest accepts sha384 and sha512, so
// without an explicit check the digest string cannot be split as assumed.
func TestOciSha256RejectsNonSha256RegistryDigest(t *testing.T) {
	for _, tc := range []struct {
		algorithm     string
		contentDigest string
	}{
		{algorithm: "sha384", contentDigest: "sha384:" + strings.Repeat("b", 96)},
		{algorithm: "sha512", contentDigest: "sha512:" + strings.Repeat("c", 128)},
	} {
		t.Run(tc.algorithm, func(t *testing.T) {
			host := fakeRegistry(t, tc.contentDigest)

			fingerprint, err := insecureAnonymousLookup(host + "/repo:tag")

			require.Error(t, err)
			require.Empty(t, fingerprint)
			require.Contains(t, err.Error(), "Kosli fingerprints are sha256")
			require.Contains(t, err.Error(), tc.algorithm)
		})
	}
}

func TestOciSha256ReturnsTheSha256Fingerprint(t *testing.T) {
	want := strings.Repeat("a", 64)
	host := fakeRegistry(t, "sha256:"+want)

	fingerprint, err := insecureAnonymousLookup(host + "/repo:tag")

	require.NoError(t, err)
	require.Equal(t, want, fingerprint)
}

// TestCredentialContext pins the credential decision itself, which is the
// load-bearing property of the anonymous lookup: containers/image falls back to
// credential discovery from auth files and helpers whenever DockerAuthConfig is
// nil, so the anonymous source must produce a non-nil empty one.
func TestCredentialContext(t *testing.T) {
	for _, tc := range []struct {
		name           string
		source         credentialSource
		username       string
		password       string
		wantAuthConfig bool
		wantUsername   string
		wantPassword   string
	}{
		{
			name:           "no credentials presents an empty config, not discovery",
			source:         noCredentials,
			wantAuthConfig: true,
		},
		{
			name:           "no credentials ignores any credentials passed alongside it",
			source:         noCredentials,
			username:       "user",
			password:       "pass",
			wantAuthConfig: true,
		},
		{
			name:           "caller credentials are presented",
			source:         callerOrHostCredentials,
			username:       "user",
			password:       "pass",
			wantAuthConfig: true,
			wantUsername:   "user",
			wantPassword:   "pass",
		},
		{
			name:           "caller credentials, password only",
			source:         callerOrHostCredentials,
			password:       "pass",
			wantAuthConfig: true,
			wantPassword:   "pass",
		},
		{
			// A nil config is what enables discovery, and this path wants it.
			name:           "no caller credentials leaves discovery enabled",
			source:         callerOrHostCredentials,
			wantAuthConfig: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sysCtx := credentialContext(tc.source, tc.username, tc.password)

			if !tc.wantAuthConfig {
				require.Nil(t, sysCtx.DockerAuthConfig, "a nil config enables credential discovery")
				return
			}
			require.NotNil(t, sysCtx.DockerAuthConfig, "a nil config would enable credential discovery")
			require.Equal(t, tc.wantUsername, sysCtx.DockerAuthConfig.Username)
			require.Equal(t, tc.wantPassword, sysCtx.DockerAuthConfig.Password)
			require.Empty(t, sysCtx.DockerAuthConfig.IdentityToken)
		})
	}
}

// TestNoCredentialsIsTheZeroValue means a lookup that fails to state its
// credential source presents nothing rather than the host's credentials.
func TestNoCredentialsIsTheZeroValue(t *testing.T) {
	var unset credentialSource
	require.Equal(t, noCredentials, unset)
	require.NotNil(t, credentialContext(unset, "", "").DockerAuthConfig)
}

func TestSha256FingerprintFromDigest(t *testing.T) {
	validHex := strings.Repeat("a", 64)

	for _, tc := range []struct {
		name        string
		digest      string
		want        string
		wantErrText string
	}{
		{name: "sha256 digest yields its hex", digest: "sha256:" + validHex, want: validHex},
		{name: "sha384 is rejected", digest: "sha384:" + strings.Repeat("b", 96), wantErrText: "algorithm is sha384"},
		{name: "sha512 is rejected", digest: "sha512:" + strings.Repeat("c", 128), wantErrText: "algorithm is sha512"},
		{name: "empty is rejected", digest: "", wantErrText: "unparseable digest"},
		{name: "bare hex without an algorithm is rejected", digest: validHex, wantErrText: "unparseable digest"},
		{name: "non hex is rejected", digest: "sha256:" + strings.Repeat("z", 64), wantErrText: "unparseable digest"},
		{name: "uppercase hex is rejected", digest: "sha256:" + strings.Repeat("A", 64), wantErrText: "unparseable digest"},
		{name: "wrong length is rejected", digest: "sha256:" + strings.Repeat("a", 63), wantErrText: "unparseable digest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Sha256FingerprintFromDigest(tc.digest)
			if tc.wantErrText != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrText)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// TestSha256FingerprintValidatesWhatItIsHanded covers the exported typed entry
// point. godigest.Digest is a string type, so an unvalidated value must not be
// split into a fingerprint.
func TestSha256FingerprintValidatesWhatItIsHanded(t *testing.T) {
	for _, tc := range []struct{ name, digest string }{
		{name: "non hex encoded portion", digest: "sha256:hello"},
		{name: "empty encoded portion", digest: "sha256:"},
		{name: "uppercase encoded portion", digest: "sha256:" + strings.Repeat("A", 64)},
		{name: "wrong length", digest: "sha256:abc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Sha256Fingerprint(godigest.Digest(tc.digest))
			require.Error(t, err)
			require.Empty(t, got)
			require.Contains(t, err.Error(), "invalid digest")
		})
	}

	valid := strings.Repeat("a", 64)
	got, err := Sha256Fingerprint(godigest.Digest("sha256:" + valid))
	require.NoError(t, err)
	require.Equal(t, valid, got)
}
