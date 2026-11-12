package intoto

import (
	"encoding/json"
	"fmt"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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
// SHA-256 digest: Sigstore bundles among its referrers, as `cosign attest` and
// actions/attest with push-to-registry attach them. Whatever else is attached
// is passed over, but an attestation that cannot be read is an error.
func (i OCIImage) Attestations(digest string) ([]*Attestation, error) {
	subject := "sha256:" + normaliseDigest(digest)
	search := &imageSearch{image: i, subject: subject, seen: map[string]bool{}}

	if err := search.referrers(); err != nil {
		return nil, err
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
		if !strings.HasPrefix(layer.MediaType, bundleMediaTypePrefix) || s.seen[layer.Digest.String()] {
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
		}
		s.found = append(s.found, attestations...)
	}
	return nil
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
