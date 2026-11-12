package digest

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/containers/image/v5/types"
	"github.com/kosli-dev/cli/internal/registrytest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

const (
	testRepo     = "acme/widget"
	testManifest = `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json","config":{"mediaType":"application/vnd.oci.empty.v1+json","digest":"sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a","size":2},"layers":[]}`
)

// allowPlainHTTP lets lookups reach the test registry, which has no TLS,
// through the seam the fingerprint tests use.
func allowPlainHTTP(t *testing.T) {
	t.Helper()
	t.Cleanup(SetOciSystemContextOverride(func(sysCtx *types.SystemContext) {
		sysCtx.DockerInsecureSkipTLSVerify = types.OptionalBoolTrue
	}))
}

func openTestRepository(t *testing.T, registry *registrytest.Registry) *Repository {
	t.Helper()
	repository, err := openRepository(registry.Host+"/"+testRepo+":1.2", noCredentials, "", "")
	require.NoError(t, err)
	return repository
}

func TestRepositoryManifest(t *testing.T) {
	allowPlainHTTP(t)
	registry := registrytest.New(t)
	pushed := registry.PushManifest(testRepo, "1.2", ocispec.MediaTypeImageManifest, []byte(testManifest))
	repository := openTestRepository(t, registry)

	t.Run("returns the manifest a tag names, with its media type", func(t *testing.T) {
		content, mediaType, err := repository.Manifest("1.2")
		require.NoError(t, err)
		require.Equal(t, testManifest, string(content))
		require.Equal(t, ocispec.MediaTypeImageManifest, mediaType)
	})

	t.Run("returns the manifest a digest names", func(t *testing.T) {
		content, _, err := repository.Manifest(pushed.Digest.String())
		require.NoError(t, err)
		require.Equal(t, testManifest, string(content))
	})

	t.Run("reports a manifest the repository does not hold as not existing", func(t *testing.T) {
		_, _, err := repository.Manifest("9.9")
		require.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("refuses content that does not hash to the digest asked for", func(t *testing.T) {
		registry.CorruptManifest(testRepo, pushed.Digest.String())
		_, _, err := repository.Manifest(pushed.Digest.String())
		require.Error(t, err)
		require.NotErrorIs(t, err, fs.ErrNotExist)
	})
}

func TestRepositoryBlob(t *testing.T) {
	allowPlainHTTP(t)
	registry := registrytest.New(t)
	pushed := registry.PushBlob(testRepo, "application/vnd.in-toto+json", []byte(`{"an": "attestation"}`))
	repository := openTestRepository(t, registry)

	t.Run("returns the blob a digest names", func(t *testing.T) {
		content, err := repository.Blob(pushed.Digest.String())
		require.NoError(t, err)
		require.Equal(t, `{"an": "attestation"}`, string(content))
	})

	t.Run("refuses a blob larger than the limit", func(t *testing.T) {
		large := registry.PushBlob(testRepo, "application/octet-stream", []byte(strings.Repeat("a", 64)))
		limited := openTestRepository(t, registry)
		limited.maxBlobSize = 63
		_, err := limited.Blob(large.Digest.String())
		require.ErrorContains(t, err, "larger than")
	})

	t.Run("refuses content that does not hash to the digest asked for", func(t *testing.T) {
		registry.CorruptBlob(testRepo, pushed.Digest.String())
		_, err := repository.Blob(pushed.Digest.String())
		require.Error(t, err)
	})
}

func TestRepositoryReferrers(t *testing.T) {
	allowPlainHTTP(t)
	bundle := []byte(`{"mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json"}`)

	for _, tc := range []struct {
		name    string
		options []registrytest.Option
	}{
		{name: "from the Referrers API"},
		{name: "from the fallback tag of a registry without the Referrers API", options: []registrytest.Option{registrytest.WithoutReferrersAPI()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registry := registrytest.New(t, tc.options...)
			image := registry.PushManifest(testRepo, "1.2", ocispec.MediaTypeImageManifest, []byte(testManifest))
			referrer := registry.AttachBundle(testRepo, image.Digest.String(), bundle)
			repository := openTestRepository(t, registry)

			referrers, err := repository.Referrers(image.Digest.String())
			require.NoError(t, err)
			require.Len(t, referrers, 1)
			require.Equal(t, referrer.Digest, referrers[0].Digest)

			t.Run("and none for a manifest nothing refers to", func(t *testing.T) {
				referrers, err := repository.Referrers("sha256:" + strings.Repeat("0", 64))
				require.NoError(t, err)
				require.Empty(t, referrers)
			})
		})
	}
}

func TestRepositoryCredentials(t *testing.T) {
	allowPlainHTTP(t)
	registry := registrytest.New(t, registrytest.RequiringCredentials("a-user", "a-password"))
	registry.PushManifest(testRepo, "1.2", ocispec.MediaTypeImageManifest, []byte(testManifest))
	image := registry.Host + "/" + testRepo + ":1.2"

	t.Run("presents the caller's credentials", func(t *testing.T) {
		repository, err := OpenRepository(image, "a-user", "a-password")
		require.NoError(t, err)
		_, _, err = repository.Manifest("1.2")
		require.NoError(t, err)
	})

	t.Run("presents nothing without credentials to present", func(t *testing.T) {
		repository, err := openRepository(image, noCredentials, "", "")
		require.NoError(t, err)
		_, _, err = repository.Manifest("1.2")
		require.Error(t, err)
		require.NotErrorIs(t, err, fs.ErrNotExist)
	})
}

// Without the override the registry's plain HTTP is refused, as it is for
// fingerprinting: credentials must not travel unencrypted.
func TestRepositoryRefusesPlainHTTPWithoutTheOverride(t *testing.T) {
	registry := registrytest.New(t)
	registry.PushManifest(testRepo, "1.2", ocispec.MediaTypeImageManifest, []byte(testManifest))

	repository, err := openRepository(registry.Host+"/"+testRepo+":1.2", noCredentials, "", "")
	require.NoError(t, err)
	_, _, err = repository.Manifest("1.2")
	require.ErrorContains(t, err, "plain HTTP")
}

func TestOpenRepositoryRejectsAnInvalidImageName(t *testing.T) {
	_, err := OpenRepository("Not A Reference", "", "")
	require.ErrorContains(t, err, "failed to parse image reference")
}
