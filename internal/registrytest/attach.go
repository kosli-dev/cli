package registrytest

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"

	godigest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	bundleMediaType        = "application/vnd.dev.sigstore.bundle.v0.3+json"
	dsseEnvelopeMediaType  = "application/vnd.dsse.envelope.v1+json"
	inTotoMediaType        = "application/vnd.in-toto+json"
	buildxAttestationType  = "application/vnd.docker.attestation.manifest.v1+json"
	cosignSignature        = "dev.cosignproject.cosign/signature"
	cosignCertificate      = "dev.sigstore.cosign/certificate"
	cosignBundle           = "dev.sigstore.cosign/bundle"
	cosignRFC3161Timestamp = "dev.sigstore.cosign/rfc3161timestamp"
)

// PushImage stores a minimal single-platform image in repo, under its own
// digest, and returns its descriptor.
func (r *Registry) PushImage(repo string) ocispec.Descriptor {
	r.t.Helper()
	manifest := marshal(r.t, ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    r.PushBlob(repo, ocispec.MediaTypeImageConfig, []byte(`{"architecture":"amd64","os":"linux"}`)),
		Layers:    []ocispec.Descriptor{},
	})
	return r.PushManifest(repo, Digest(manifest).String(), ocispec.MediaTypeImageManifest, manifest)
}

// AttachBundle attaches a Sigstore bundle to the manifest with digest subject
// as `cosign attest` does in its bundle format, and actions/attest with
// push-to-registry: a referrer whose one layer is the bundle.
func (r *Registry) AttachBundle(repo, subject string, bundle []byte) ocispec.Descriptor {
	r.t.Helper()
	layer := r.PushBlob(repo, bundleMediaType, bundle)
	return r.attach(repo, subject, bundleMediaType, []ocispec.Descriptor{layer})
}

// AttachCosignAttestations stores attestations under the image's
// sha256-<digest>.att tag, as `cosign attest` did before its bundle format:
// one DSSE envelope layer per attestation, with the signing certificate and
// the transparency log entry in the layer's annotations. Each attestation is
// given as a Sigstore bundle, or as a DSSE envelope for one signed with a key
// and never logged.
func (r *Registry) AttachCosignAttestations(repo, subject string, attestations ...[]byte) ocispec.Descriptor {
	r.t.Helper()
	manifest := marshal(r.t, ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    r.PushBlob(repo, ocispec.MediaTypeImageConfig, []byte(`{}`)),
		Layers:    r.cosignLayers(repo, attestations),
	})
	return r.PushManifest(repo, strings.Replace(subject, ":", "-", 1)+".att", ocispec.MediaTypeImageManifest, manifest)
}

// AttachCosignReferrer attaches attestations in cosign's layout to the
// manifest with digest subject as a referrer rather than under a tag, as
// cosign's experimental OCI 1.1 mode does.
func (r *Registry) AttachCosignReferrer(repo, subject string, attestations ...[]byte) ocispec.Descriptor {
	r.t.Helper()
	return r.attach(repo, subject, inTotoMediaType, r.cosignLayers(repo, attestations))
}

// BuildxAttestations is what `docker buildx build --provenance --sbom`
// attaches to a platform image.
type BuildxAttestations struct {
	// Statements are the unsigned in-toto statements about the platform image.
	Statements [][]byte
	// OCIArtifact has the attestation manifest name the platform image as its
	// subject, as BuildKit does when it stores attestations as OCI artifacts,
	// which makes it a referrer of the platform image too.
	OCIArtifact bool
}

// PushBuildxIndex stores, under tag, the image index buildx pushes: the
// platform image and the attestation manifest about it.
func (r *Registry) PushBuildxIndex(repo, tag string, platform ocispec.Descriptor, attestations BuildxAttestations) ocispec.Descriptor {
	r.t.Helper()
	layers := make([]ocispec.Descriptor, 0, len(attestations.Statements))
	for _, statement := range attestations.Statements {
		layer := r.PushBlob(repo, inTotoMediaType, statement)
		layer.Annotations = map[string]string{"in-toto.io/predicate-type": predicateType(r.t, statement)}
		layers = append(layers, layer)
	}
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    r.PushBlob(repo, ocispec.MediaTypeEmptyJSON, []byte(`{}`)),
		Layers:    layers,
	}
	if attestations.OCIArtifact {
		manifest.ArtifactType = buildxAttestationType
		manifest.Subject = &ocispec.Descriptor{MediaType: platform.MediaType, Digest: platform.Digest, Size: platform.Size}
	}
	content := marshal(r.t, manifest)
	attestation := r.PushManifest(repo, Digest(content).String(), ocispec.MediaTypeImageManifest, content)
	attestation.Annotations = map[string]string{
		"vnd.docker.reference.type":   "attestation-manifest",
		"vnd.docker.reference.digest": platform.Digest.String(),
	}
	attestation.Platform = &ocispec.Platform{Architecture: "unknown", OS: "unknown"}
	platform.Platform = &ocispec.Platform{Architecture: "amd64", OS: "linux"}

	index := marshal(r.t, ocispec.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageIndex,
		Manifests: []ocispec.Descriptor{platform, attestation},
	})
	return r.PushManifest(repo, tag, ocispec.MediaTypeImageIndex, index)
}

// attach stores a manifest whose subject is the manifest with digest subject.
// A registry without the Referrers API lists it under the subject's fallback
// tag, which clients maintain themselves.
func (r *Registry) attach(repo, subject, artifactType string, layers []ocispec.Descriptor) ocispec.Descriptor {
	r.t.Helper()
	manifest := marshal(r.t, ocispec.Manifest{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    ocispec.MediaTypeImageManifest,
		ArtifactType: artifactType,
		Config:       r.PushBlob(repo, ocispec.MediaTypeEmptyJSON, []byte(`{}`)),
		Layers:       layers,
		Subject:      &ocispec.Descriptor{MediaType: ocispec.MediaTypeImageManifest, Digest: godigest.Digest(subject)},
	})
	referrer := r.PushManifest(repo, Digest(manifest).String(), ocispec.MediaTypeImageManifest, manifest)
	if r.referrersAPI {
		return referrer
	}

	fallbackTag := strings.Replace(subject, ":", "-", 1)
	index := ocispec.Index{Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageIndex}
	if existing, found := r.manifest(repo, fallbackTag); found {
		require.NoError(r.t, json.Unmarshal(existing, &index))
	}
	listed := referrer
	listed.ArtifactType = artifactType
	index.Manifests = append(index.Manifests, listed)
	r.PushManifest(repo, fallbackTag, ocispec.MediaTypeImageIndex, marshal(r.t, index))
	return referrer
}

func (r *Registry) cosignLayers(repo string, attestations [][]byte) []ocispec.Descriptor {
	r.t.Helper()
	layers := make([]ocispec.Descriptor, 0, len(attestations))
	for _, attestation := range attestations {
		envelope, annotations := CosignLayer(r.t, attestation)
		layer := r.PushBlob(repo, dsseEnvelopeMediaType, envelope)
		layer.Annotations = annotations
		layers = append(layers, layer)
	}
	return layers
}

// The shapes cosign marshals into a layer and its annotations.
type (
	cosignEnvelope struct {
		PayloadType string                    `json:"payloadType"`
		Payload     []byte                    `json:"payload"`
		Signatures  []cosignEnvelopeSignature `json:"signatures"`
	}
	cosignEnvelopeSignature struct {
		KeyID string `json:"keyid"`
		Sig   []byte `json:"sig"`
	}
	cosignRekorBundle struct {
		SignedEntryTimestamp []byte
		Payload              cosignRekorPayload
	}
	cosignRekorPayload struct {
		Body           string `json:"body"`
		IntegratedTime int64  `json:"integratedTime"`
		LogIndex       int64  `json:"logIndex"`
		LogID          string `json:"logID"`
	}
	cosignTimestamp struct {
		SignedRFC3161Timestamp []byte
	}
)

// CosignLayer splits an attestation, given as for AttachCosignAttestations,
// into the DSSE envelope cosign stores as a layer and the annotations it
// stores the verification material in.
func CosignLayer(t testing.TB, attestation []byte) ([]byte, map[string]string) {
	t.Helper()
	annotations := map[string]string{cosignSignature: ""}

	var probe struct {
		MediaType string `json:"mediaType"`
	}
	require.NoError(t, json.Unmarshal(attestation, &probe))
	if !strings.HasPrefix(probe.MediaType, "application/vnd.dev.sigstore.bundle") {
		annotations["predicateType"] = predicateType(t, envelopePayload(t, attestation))
		return attestation, annotations
	}

	var bundle protobundle.Bundle
	require.NoError(t, protojson.Unmarshal(attestation, &bundle))
	signed := bundle.GetDsseEnvelope()
	require.NotNil(t, signed, "cosign stores attestations, not signatures over a blob")
	envelope := cosignEnvelope{PayloadType: signed.GetPayloadType(), Payload: signed.GetPayload()}
	for _, signature := range signed.GetSignatures() {
		envelope.Signatures = append(envelope.Signatures, cosignEnvelopeSignature{KeyID: signature.GetKeyid(), Sig: signature.GetSig()})
	}
	annotations["predicateType"] = predicateType(t, signed.GetPayload())

	material := bundle.GetVerificationMaterial()
	if certificate := material.GetCertificate(); certificate != nil {
		annotations[cosignCertificate] = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.GetRawBytes()}))
	}
	if entries := material.GetTlogEntries(); len(entries) > 0 {
		entry := entries[0]
		annotations[cosignBundle] = string(marshal(t, cosignRekorBundle{
			SignedEntryTimestamp: entry.GetInclusionPromise().GetSignedEntryTimestamp(),
			Payload: cosignRekorPayload{
				Body:           base64.StdEncoding.EncodeToString(entry.GetCanonicalizedBody()),
				IntegratedTime: entry.GetIntegratedTime(),
				LogIndex:       entry.GetLogIndex(),
				LogID:          hex.EncodeToString(entry.GetLogId().GetKeyId()),
			},
		}))
	}
	if timestamps := material.GetTimestampVerificationData().GetRfc3161Timestamps(); len(timestamps) > 0 {
		annotations[cosignRFC3161Timestamp] = string(marshal(t, cosignTimestamp{SignedRFC3161Timestamp: timestamps[0].GetSignedTimestamp()}))
	}
	return marshal(t, envelope), annotations
}

func envelopePayload(t testing.TB, envelope []byte) []byte {
	t.Helper()
	var parsed struct {
		Payload []byte `json:"payload"`
	}
	require.NoError(t, json.Unmarshal(envelope, &parsed))
	return parsed.Payload
}

func predicateType(t testing.TB, statement []byte) string {
	t.Helper()
	var parsed struct {
		PredicateType string `json:"predicateType"`
	}
	require.NoError(t, json.Unmarshal(statement, &parsed))
	return parsed.PredicateType
}
