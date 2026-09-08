package digest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containers/image/v5/types"
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
	return strings.TrimPrefix(srv.URL, "https://")
}

func insecureAnonymousContext() *types.SystemContext {
	return &types.SystemContext{
		DockerAuthConfig:            &types.DockerAuthConfig{},
		DockerInsecureSkipTLSVerify: types.OptionalBoolTrue,
	}
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

			fingerprint, err := ociSha256(host+"/repo:tag", insecureAnonymousContext())

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

	fingerprint, err := ociSha256(host+"/repo:tag", insecureAnonymousContext())

	require.NoError(t, err)
	require.Equal(t, want, fingerprint)
}

// TestOciSha256AnonymousPresentsNoStoredCredential guards the credential
// boundary: a nil DockerAuthConfig makes containers/image fall back to
// credential discovery, so the anonymous helper must set a non-nil empty one.
func TestOciSha256AnonymousPresentsNoStoredCredential(t *testing.T) {
	sysCtx := anonymousSystemContext()
	require.NotNil(t, sysCtx.DockerAuthConfig, "a nil DockerAuthConfig falls back to credential discovery")
	require.Empty(t, sysCtx.DockerAuthConfig.Username)
	require.Empty(t, sysCtx.DockerAuthConfig.Password)
	require.Empty(t, sysCtx.DockerAuthConfig.IdentityToken)
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
