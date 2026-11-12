package intoto

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
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

	t.Run("fails when the registry cannot be read, rather than finding nothing attached", func(t *testing.T) {
		_, image := testImage(t)
		refused, err := digest.OpenRepository(image.Name, "a-user", "not-the-password")
		require.NoError(t, err)
		image.Repository = refused

		_, err = image.Attestations(artifactDigest)
		require.ErrorContains(t, err, "failed to read")
	})

	t.Run("says what it looked for when nothing is attached", func(t *testing.T) {
		_, image := testImage(t)

		_, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.EqualError(t, err, "image "+image.Name+": no attestation is attached to sha256:"+artifactDigest+
			": looked among its referrers, under cosign's tag sha256-"+artifactDigest+".att, and for buildx attestation manifests in the image index")
	})
}

// unreadableCosignTag is a repository holding nothing about the image but a
// cosign tag that cannot be read.
type unreadableCosignTag struct{ err error }

func (r unreadableCosignTag) Manifest(reference string) ([]byte, string, error) {
	if strings.HasSuffix(reference, ".att") {
		return nil, "", r.err
	}
	return nil, "", fs.ErrNotExist
}
func (r unreadableCosignTag) Blob(string) ([]byte, error) { return nil, fs.ErrNotExist }
func (r unreadableCosignTag) Referrers(string) ([]ocispec.Descriptor, error) {
	return nil, nil
}

func TestOCIImageCosignTag(t *testing.T) {
	t.Run("fails when the tag cannot be read, rather than finding nothing attached", func(t *testing.T) {
		image := OCIImage{Name: "registry.example.com/acme/widget:1.2", Repository: unreadableCosignTag{err: errors.New("registry unavailable")}}

		_, err := image.Attestations(artifactDigest)
		require.ErrorContains(t, err, "registry unavailable")
	})

	t.Run("verifies an attestation under cosign's tag against the pinned identity", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachCosignAttestations(ociRepo, ociSubject, fixture(t, "github-provenance.sigstore.json"))

		found, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.NoError(t, err)
		require.True(t, found.Evidence.Verified)
		require.Equal(t, fixtureIdentity, found.Evidence.SignerIdentity)
		require.Equal(t, fixtureIssuer, found.Evidence.SignerIssuer)
		require.NotNil(t, found.Evidence.LogIndex)
		require.False(t, found.Evidence.IntegratedTime.IsZero())
		require.False(t, found.Evidence.SignedTime.IsZero())

		envelope, _ := registrytest.CosignLayer(t, fixture(t, "github-provenance.sigstore.json"))
		require.Equal(t, envelope, found.Raw, "Raw should be the layer, so its checksum is the layer's digest")
		require.Equal(t, "attestation sha256:"+sha256Hex(envelope)+" (under cosign's tag sha256-"+artifactDigest+".att)", found.Describe())
	})

	t.Run("chooses among the attestations the tag holds", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachCosignAttestations(ociRepo, ociSubject,
			fixture(t, "github-provenance.sigstore.json"), fixture(t, "cyclonedx-sbom.sigstore.json"))

		found, err := Find(image, KindSBOM, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.NoError(t, err)
		require.Equal(t, "https://cyclonedx.org/bom", found.Statement.GetPredicateType())
	})

	t.Run("verifies an attestation signed with a key", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachCosignAttestations(ociRepo, ociSubject, fixture(t, "key-signed-provenance.dsse.json"))

		found, err := Find(image, KindProvenance, artifactDigest,
			NewVerifier(Policy{KeyPath: filepath.Join("testdata", "key.pub")}, fixtureTrust(t)))
		require.NoError(t, err)
		require.True(t, found.Evidence.Verified)
	})

	t.Run("fails verification for one signed by another identity", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachCosignAttestations(ociRepo, ociSubject, fixture(t, "wrong-signer-provenance.sigstore.json"))

		_, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.ErrorIs(t, err, ErrFailedVerification)
		require.ErrorContains(t, err, fixtureOtherIdentity)
	})

	t.Run("fails verification for one changed after signing", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachCosignAttestations(ociRepo, ociSubject, fixture(t, "tampered-provenance.sigstore.json"))

		_, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.ErrorIs(t, err, ErrFailedVerification)
	})

	t.Run("reads cosign's layers attached as a referrer rather than under the tag", func(t *testing.T) {
		registry, image := testImage(t)
		registry.AttachCosignReferrer(ociRepo, ociSubject, fixture(t, "github-provenance.sigstore.json"))

		found, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.NoError(t, err)
		require.True(t, found.Evidence.Verified)
	})
}

func TestCosignVerificationMaterial(t *testing.T) {
	cosignAttestation := func(t *testing.T, change func(annotations map[string]string)) *Attestation {
		envelope, annotations := registrytest.CosignLayer(t, fixture(t, "github-provenance.sigstore.json"))
		change(annotations)
		attestations, err := Read(envelope)
		require.NoError(t, err)
		attestations[0].addCosignMaterial(annotations)
		return attestations[0]
	}
	verify := func(t *testing.T, policy Policy, a *Attestation) error {
		_, err := NewVerifier(policy, fixtureTrust(t)).Verify(a, artifactDigest)
		return err
	}

	t.Run("cannot verify a certificate's signature logged as an entry it cannot read, saying which", func(t *testing.T) {
		a := cosignAttestation(t, func(annotations map[string]string) {
			annotations[cosignBundleAnnotation] = `{"SignedEntryTimestamp": "c2V0", "Payload": {"body": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaW50b3RvIiwic3BlYyI6e319", "integratedTime": 1, "logIndex": 1, "logID": "00"}}`
		})

		err := verify(t, pinnedSigner, a)
		require.ErrorIs(t, err, ErrCannotVerify)
		require.ErrorContains(t, err, "the transparency log entry cosign stored with it cannot be read")
	})

	t.Run("cannot verify against a certificate annotation that holds none", func(t *testing.T) {
		a := cosignAttestation(t, func(annotations map[string]string) {
			annotations[cosignCertificateAnnotation] = "not a certificate"
		})

		err := verify(t, pinnedSigner, a)
		require.ErrorIs(t, err, ErrCannotVerify)
		require.ErrorContains(t, err, "the certificate cosign stored with it is not a PEM certificate")
	})

	t.Run("fails verification without the log entry and timestamp a certificate's signature needs", func(t *testing.T) {
		a := cosignAttestation(t, func(annotations map[string]string) {
			delete(annotations, cosignBundleAnnotation)
			delete(annotations, cosignTimestampAnnotation)
		})

		require.ErrorIs(t, verify(t, pinnedSigner, a), ErrFailedVerification)
	})

	t.Run("says a key-signed envelope needs --key when an identity is pinned", func(t *testing.T) {
		attestations, err := Read(fixture(t, "key-signed-provenance.dsse.json"))
		require.NoError(t, err)
		attestations[0].addCosignMaterial(map[string]string{"dev.cosignproject.cosign/signature": ""})

		err = verify(t, pinnedSigner, attestations[0])
		require.ErrorIs(t, err, ErrCannotVerify)
		require.ErrorContains(t, err, "--key")
	})
}

func TestOCIImageBuildxAttestations(t *testing.T) {
	allowUnsigned := Policy{AllowUnsigned: true}
	statements := func(t *testing.T) registrytest.BuildxAttestations {
		return registrytest.BuildxAttestations{Statements: [][]byte{
			statementFixture(t, "buildkit-provenance.json"), statementFixture(t, "spdx-sbom.json"),
		}}
	}

	t.Run("finds an unsigned statement in the image index the tag names", func(t *testing.T) {
		registry, image := testImage(t)
		registry.PushBuildxIndex(ociRepo, "1.2", registry.PushImage(ociRepo), statements(t))

		found, err := Find(image, KindProvenance, artifactDigest, NewVerifier(allowUnsigned, fixtureTrust(t)))
		require.NoError(t, err)
		require.False(t, found.Evidence.Verified)
		require.Equal(t, statementFixture(t, "buildkit-provenance.json"), found.Raw)
		require.Equal(t, "attestation sha256:"+sha256Hex(found.Raw)+" (in the image index)", found.Describe())
	})

	t.Run("finds it in the image index a digest names", func(t *testing.T) {
		registry, image := testImage(t)
		index := registry.PushBuildxIndex(ociRepo, "1.2", registry.PushImage(ociRepo), statements(t))
		image.Name = registry.Host + "/" + ociRepo + "@" + index.Digest.String()

		found, err := Find(image, KindSBOM, artifactDigest, NewVerifier(allowUnsigned, fixtureTrust(t)))
		require.NoError(t, err)
		require.Equal(t, "https://spdx.dev/Document/v2.3", found.Statement.GetPredicateType())
	})

	t.Run("refuses an unsigned statement unless they are allowed", func(t *testing.T) {
		registry, image := testImage(t)
		registry.PushBuildxIndex(ociRepo, "1.2", registry.PushImage(ociRepo), statements(t))

		_, err := Find(image, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.ErrorContains(t, err, "is not signed. Pass --allow-unsigned")
	})

	t.Run("names the platform images the statements are about when the artifact is not one of them", func(t *testing.T) {
		registry, image := testImage(t)
		index := registry.PushBuildxIndex(ociRepo, "1.2", registry.PushImage(ociRepo), statements(t))

		_, err := Find(image, KindProvenance, index.Digest.Encoded(), NewVerifier(allowUnsigned, fixtureTrust(t)))
		require.ErrorContains(t, err, "no provenance attestation lists the artifact (sha256:"+index.Digest.Encoded()+") as its subject")
		require.ErrorContains(t, err, "pkg:docker/acme/widget@1.2.3?platform=linux%2Famd64 (sha256:"+artifactDigest+")")
	})

	t.Run("returns once a statement that is both in the index and a referrer of the platform image", func(t *testing.T) {
		registry, image := testImage(t)
		platform := registry.PushImage(ociRepo)
		statement := []byte(strings.ReplaceAll(string(statementFixture(t, "buildkit-provenance.json")), artifactDigest, platform.Digest.Encoded()))
		registry.PushBuildxIndex(ociRepo, "1.2", platform, registrytest.BuildxAttestations{Statements: [][]byte{statement}, OCIArtifact: true})

		attestations, err := image.Attestations(platform.Digest.Encoded())
		require.NoError(t, err)
		require.Len(t, attestations, 1)
	})

	t.Run("finds nothing in a tag that names a single image", func(t *testing.T) {
		registry, image := testImage(t)
		pushed := registry.PushImage(ociRepo)
		content, _, err := image.Repository.Manifest(pushed.Digest.String())
		require.NoError(t, err)
		registry.PushManifest(ociRepo, "1.2", ocispec.MediaTypeImageManifest, content)

		_, err = image.Attestations(pushed.Digest.Encoded())
		require.ErrorContains(t, err, "no attestation is attached")
	})

	t.Run("refuses a name that is not an image reference", func(t *testing.T) {
		_, image := testImage(t)
		image.Name = "Not A Reference"

		_, err := image.Attestations(artifactDigest)
		require.ErrorContains(t, err, "is not an image reference")
	})
}
