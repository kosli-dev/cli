package intoto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/distribution/reference"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// How buildx marks an attestation manifest in an image index.
const (
	buildxReferenceTypeAnnotation = "vnd.docker.reference.type"
	buildxAttestationManifest     = "attestation-manifest"
	dockerManifestListMediaType   = "application/vnd.docker.distribution.manifest.list.v2+json"
)

// ImageRepository reads the registry repository that holds an image.
type ImageRepository interface {
	// Manifest returns the manifest that reference, a tag or a digest, names,
	// and its media type. The error wraps fs.ErrNotExist when the repository
	// holds no such manifest.
	Manifest(reference string) (content []byte, mediaType string, err error)
	// Blob returns the content of the blob with digest.
	Blob(digest string) ([]byte, error)
	// Referrers lists the manifests whose subject is the manifest with
	// digest. It is empty when nothing refers to the manifest.
	Referrers(digest string) ([]ocispec.Descriptor, error)
}

// OCIImage is an image in an OCI registry, where the tools that build and
// sign it attach attestations about it.
type OCIImage struct {
	// Name is the image's reference, such as ghcr.io/acme/widget:1.2.
	Name string
	// Repository reads the repository that holds the image.
	Repository ImageRepository
}

func (i OCIImage) String() string {
	return "image " + i.Name
}

// Attestations returns the attestations attached to the image manifest with
// SHA-256 digest:
//
//   - among its referrers, as `cosign attest` and actions/attest with
//     push-to-registry attach them;
//   - under the tag sha256-<digest>.att, as cosign attached them before its
//     bundle format;
//   - in the attestation manifests of an image index, as `docker buildx build
//     --provenance --sbom` stores them. The index is the manifest with digest
//     itself, or the one the image's name refers to.
//
// Whatever else is attached is passed over, but an attestation that cannot be
// read is an error.
func (i OCIImage) Attestations(digest string) ([]*Attestation, error) {
	subject := "sha256:" + normaliseDigest(digest)
	search := &imageSearch{image: i, subject: subject, seen: map[string]bool{}}

	for _, look := range []func() error{search.index, search.cosignTag, search.referrers} {
		if err := look(); err != nil {
			return nil, err
		}
	}
	if len(search.found) == 0 {
		return nil, fmt.Errorf("no attestation is attached to %s: looked among its referrers, under cosign's tag %s, and for buildx attestation manifests in the image index",
			subject, cosignAttestationTag(subject))
	}
	return search.found, nil
}

// imageSearch gathers the attestations attached to one image manifest.
type imageSearch struct {
	image   OCIImage
	subject string
	found   []*Attestation
	// seen holds the digests of the layers already read, since one
	// attestation can be reached by more than one route.
	seen map[string]bool
}

// index reads the attestation manifests buildx puts in an image index beside
// the platform images they are about. The statements name a platform image,
// not the index, so the index the image's name refers to is read as well as
// the manifest searched for.
func (s *imageSearch) index() error {
	named, err := s.image.reference()
	if err != nil {
		return err
	}
	references := []string{s.subject}
	if named != s.subject {
		references = append(references, named)
	}
	for _, reference := range references {
		content, mediaType, err := s.image.Repository.Manifest(reference)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if mediaType != ocispec.MediaTypeImageIndex && mediaType != dockerManifestListMediaType {
			continue
		}
		var index ocispec.Index
		if err := json.Unmarshal(content, &index); err != nil {
			return fmt.Errorf("image index %s is not valid JSON: %w", reference, err)
		}
		for _, descriptor := range index.Manifests {
			if descriptor.Annotations[buildxReferenceTypeAnnotation] != buildxAttestationManifest {
				continue
			}
			manifest, err := s.manifest(descriptor.Digest.String())
			if err != nil {
				return err
			}
			if manifest == nil {
				continue
			}
			if err := s.layers(manifest, "in the image index"); err != nil {
				return err
			}
		}
	}
	return nil
}

// reference is the tag or digest the image's name refers to.
func (i OCIImage) reference() (string, error) {
	named, err := reference.ParseNormalizedNamed(i.Name)
	if err != nil {
		return "", fmt.Errorf("%q is not an image reference: %w", i.Name, err)
	}
	if digested, ok := named.(reference.Digested); ok {
		return digested.Digest().String(), nil
	}
	if tagged, ok := reference.TagNameOnly(named).(reference.Tagged); ok {
		return tagged.Tag(), nil
	}
	return "", fmt.Errorf("%q names neither a tag nor a digest", i.Name)
}

func (s *imageSearch) cosignTag() error {
	tag := cosignAttestationTag(s.subject)
	manifest, err := s.manifest(tag)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil || manifest == nil {
		return err
	}
	return s.layers(manifest, "under cosign's tag "+tag)
}

func (s *imageSearch) referrers() error {
	referrers, err := s.image.Repository.Referrers(s.subject)
	if err != nil {
		return err
	}
	for _, referrer := range referrers {
		manifest, err := s.manifest(referrer.Digest.String())
		if err != nil {
			return err
		}
		// A registry without the Referrers API lists referrers under a tag
		// that any client with push access writes, so the listing is only a
		// hint: the manifest itself must name the image.
		if manifest == nil || manifest.Subject == nil || manifest.Subject.Digest.String() != s.subject {
			continue
		}
		if err := s.layers(manifest, "a referrer of the image"); err != nil {
			return err
		}
	}
	return nil
}

// manifest returns the image manifest that reference names, or nil when it
// names something else, such as an image index.
func (s *imageSearch) manifest(reference string) (*ocispec.Manifest, error) {
	content, mediaType, err := s.image.Repository.Manifest(reference)
	if err != nil {
		return nil, err
	}
	if mediaType != ocispec.MediaTypeImageManifest {
		return nil, nil
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("manifest %s is not valid JSON: %w", reference, err)
	}
	return &manifest, nil
}

// layers reads the attestations among a manifest's layers. attached says how
// the manifest is attached to the image, for naming an attestation in a
// message.
func (s *imageSearch) layers(manifest *ocispec.Manifest, attached string) error {
	for _, layer := range manifest.Layers {
		if !holdsAttestations(layer.MediaType) || s.seen[layer.Digest.String()] {
			continue
		}
		s.seen[layer.Digest.String()] = true

		content, err := s.image.Repository.Blob(layer.Digest.String())
		if err != nil {
			return err
		}
		if !holdsInTotoStatement(content) {
			continue
		}
		origin := fmt.Sprintf("%s (%s)", layer.Digest, attached)
		attestations, err := Read(content)
		if err != nil {
			return fmt.Errorf("attestation %s: %w", origin, err)
		}
		for _, attestation := range attestations {
			attestation.Origin = origin
			if layer.MediaType == cosignEnvelopeMediaType {
				attestation.addCosignMaterial(layer.Annotations)
			}
		}
		s.found = append(s.found, attestations...)
	}
	return nil
}

// holdsAttestations reports whether a layer of this media type is one a tool
// stores an attestation in: a Sigstore bundle, the DSSE envelope of cosign's
// earlier layout, or the bare statement buildx writes.
func holdsAttestations(mediaType string) bool {
	return strings.HasPrefix(mediaType, bundleMediaTypePrefix) || mediaType == cosignEnvelopeMediaType || mediaType == inTotoPayloadType
}

// holdsInTotoStatement reports whether a layer's content is an in-toto
// attestation rather than something stored the same way: a Sigstore bundle
// that signs a blob, as `cosign sign` attaches, or a DSSE envelope with
// another payload. Content it cannot tell apart is left for Read to refuse.
func holdsInTotoStatement(content []byte) bool {
	var probe struct {
		PayloadType      string          `json:"payloadType"`
		MessageSignature json.RawMessage `json:"messageSignature"`
		DSSEEnvelope     *struct {
			PayloadType string `json:"payloadType"`
		} `json:"dsseEnvelope"`
	}
	if json.Unmarshal(content, &probe) != nil {
		return true
	}
	switch {
	case probe.DSSEEnvelope != nil:
		return probe.DSSEEnvelope.PayloadType == inTotoPayloadType
	case probe.MessageSignature != nil:
		return false
	case probe.PayloadType != "":
		return probe.PayloadType == inTotoPayloadType
	}
	return true
}

func cosignAttestationTag(subject string) string {
	return strings.Replace(subject, ":", "-", 1) + ".att"
}
