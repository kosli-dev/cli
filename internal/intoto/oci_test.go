package intoto

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/containers/image/v5/types"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/registrytest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

const (
	ociRepo    = "acme/widget"
	ociSubject = "sha256:" + artifactDigest
)

// testImage starts a registry and returns it with the image acme/widget:1.2
// in it as a source. The registry demands credentials so that none are looked
// for on the host.
func testImage(t *testing.T, options ...registrytest.Option) (*registrytest.Registry, OCIImage) {
	t.Helper()
	t.Cleanup(digest.SetOciSystemContextOverride(func(sysCtx *types.SystemContext) {
		sysCtx.DockerInsecureSkipTLSVerify = types.OptionalBoolTrue
	}))
	registry := registrytest.New(t, append(options, registrytest.RequiringCredentials("a-user", "a-password"))...)
	name := registry.Host + "/" + ociRepo + ":1.2"
	repository, err := digest.OpenRepository(name, "a-user", "a-password")
	require.NoError(t, err)
	return registry, OCIImage{Name: name, Repository: repository}
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return content
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// messageSignatureBundle is a Sigstore bundle that signs a blob, as `cosign
// sign` attaches to an image, rather than an attestation.
const messageSignatureBundle = `{
	"mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json",
	"verificationMaterial": {"publicKey": {"hint": "a-key"}},
	"messageSignature": {"messageDigest": {"algorithm": "SHA2_256", "digest": "6IlRKRvKCAekHzwBRACb//XGx33BMQ1S6MbWL+7pYKg="}, "signature": "c2lnbmF0dXJl"}
}`

func TestOCIImageReferrers(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options []registrytest.Option
	}{
		{name: "listed by the Referrers API"},
		{name: "listed under the fallback tag", options: []registrytest.Option{registrytest.WithoutReferrersAPI()}},
	} {
		t.Run("finds a Sigstore bundle attached as a referrer "+tc.name, func(t *testing.T) {
			registry, image := testImage(t, tc.options...)
			bundle := fixture(t, "github-provenance.sigstore.json")
			registry.AttachBundle(ociRepo, ociSubject, bundle)

			found, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
			require.NoError(t, err)
			require.True(t, found.Evidence.Verified)
			require.Equal(t, fixtureIdentity, found.Evidence.SignerIdentity)
			require.Equal(t, bundle, found.Raw, "Raw should be the layer, so its checksum is the layer's digest")
			require.Equal(t, sha256Hex(bundle), found.Evidence.RawSHA256)
		})
	}

	t.Run("chooses the attestation of the kind asked for", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachBundle(ociRepo, ociSubject, fixture(t, "github-provenance.sigstore.json"))
		registry.AttachBundle(ociRepo, ociSubject, fixture(t, "spdx-sbom.sigstore.json"))

		found, err := Find(image, KindSBOM, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.NoError(t, err)
		require.Equal(t, "https://spdx.dev/Document/v2.3", found.Statement.GetPredicateType())
	})

	t.Run("passes over what is attached but is not an attestation", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachBundle(ociRepo, ociSubject, []byte(messageSignatureBundle))
		registry.AttachBundle(ociRepo, ociSubject, fixture(t, "github-provenance.sigstore.json"))

		attestations, err := image.Attestations(artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
	})

	t.Run("passes over a manifest listed as a referrer whose subject is another image", func(t *testing.T) {
		registry, image := testImage(t, registrytest.WithoutReferrersAPI())
		other := "sha256:" + strings.Repeat("1", 64)
		registry.AttachBundle(ociRepo, other, fixture(t, "github-provenance.sigstore.json"))
		// The fallback tag is one any client with push access can write.
		listing, _, err := image.Repository.Manifest(strings.Replace(other, ":", "-", 1))
		require.NoError(t, err)
		registry.PushManifest(ociRepo, strings.Replace(ociSubject, ":", "-", 1), ocispec.MediaTypeImageIndex, listing)

		_, err = image.Attestations(artifactDigest)
		require.ErrorContains(t, err, "no attestation is attached")
	})

	t.Run("fails on an attached attestation that cannot be read", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachBundle(ociRepo, ociSubject, []byte(`{"mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json", "dsseEnvelope": {"payloadType": "application/vnd.in-toto+json", "payload": "bm90IGEgc3RhdGVtZW50"}}`))

		_, err := image.Attestations(artifactDigest)
		require.ErrorContains(t, err, "a referrer of the image")
		require.ErrorContains(t, err, "not a valid Sigstore bundle")
	})

	t.Run("fails verification for a referrer signed by another identity, naming it", func(t *testing.T) {
		registry, image := testImage(t)
		wrongSigner := fixture(t, "wrong-signer-provenance.sigstore.json")
		registry.AttachBundle(ociRepo, ociSubject, wrongSigner)

		_, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.ErrorIs(t, err, ErrFailedVerification)
		require.ErrorContains(t, err, "image "+image.Name+": attestation sha256:"+sha256Hex(wrongSigner)+" (a referrer of the image) failed verification")
	})

	t.Run("says what it looked for when nothing is attached", func(t *testing.T) {
		_, image := testImage(t)

		_, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.EqualError(t, err, "image "+image.Name+": no attestation is attached to sha256:"+artifactDigest+
			": looked among its referrers, under cosign's tag sha256-"+artifactDigest+".att, and for buildx attestation manifests in the image index")
	})
}
