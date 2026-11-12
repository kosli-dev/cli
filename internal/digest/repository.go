package digest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/pkg/docker/config"
	"github.com/containers/image/v5/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	godigest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// defaultMaxBlobSize bounds what Blob holds in memory. Attestations are the
// blobs read this way, and an SBOM of a large image runs to tens of megabytes.
const defaultMaxBlobSize = 128 << 20

// Repository reads the manifests and blobs of the registry repository that
// holds an image, which is where tools attach attestations about it.
//
// containers/image, which fingerprints an image, has no client for the OCI
// Referrers API, so these reads go through go-containerregistry. The
// credentials are still the ones containers/image resolves, and plain HTTP is
// still refused, so a repository is reached exactly as its image is
// fingerprinted.
type Repository struct {
	repo        name.Repository
	puller      *remote.Puller
	maxBlobSize int64
}

// OpenRepository prepares reading the repository that holds the image
// artifactName, presenting the given credentials, or those the host holds
// when none are given: the credentials OciSha256 fingerprints the image with.
// They are only ever presented to that image's registry. Nothing is requested
// until the repository is read.
func OpenRepository(artifactName string, registryUsername string, registryPassword string) (*Repository, error) {
	return openRepository(artifactName, callerOrHostCredentials, registryUsername, registryPassword)
}

func openRepository(artifactName string, source credentialSource, registryUsername, registryPassword string) (*Repository, error) {
	imageName := fmt.Sprintf("//%s", artifactName)
	ref, err := docker.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference for %s: %w", imageName, err)
	}
	named := ref.DockerReference()

	sysCtx := credentialContext(source, registryUsername, registryPassword)
	applyOciSystemContextOverride(sysCtx)
	credentials, err := config.GetCredentialsForRef(sysCtx, named)
	if err != nil {
		return nil, fmt.Errorf("failed to read the registry credentials for %s: %w", named.Name(), err)
	}

	repo, err := name.NewRepository(named.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference for %s: %w", imageName, err)
	}
	puller, err := remote.NewPuller(
		remote.WithAuth(authenticator(credentials)),
		remote.WithTransport(registryTransport(sysCtx)),
	)
	if err != nil {
		return nil, err
	}
	return &Repository{repo: repo, puller: puller, maxBlobSize: defaultMaxBlobSize}, nil
}

func authenticator(credentials types.DockerAuthConfig) authn.Authenticator {
	if credentials == (types.DockerAuthConfig{}) {
		return authn.Anonymous
	}
	return authn.FromConfig(authn.AuthConfig{
		Username:      credentials.Username,
		Password:      credentials.Password,
		IdentityToken: credentials.IdentityToken,
	})
}

// registryTransport refuses plain HTTP, which go-containerregistry would
// otherwise fall back to for a registry on a loopback or private address,
// sending credentials unencrypted. Only a context that skips TLS
// verification allows it, and nothing but the test override builds one.
func registryTransport(sysCtx *types.SystemContext) http.RoundTripper {
	if sysCtx.DockerInsecureSkipTLSVerify == types.OptionalBoolTrue {
		return remote.DefaultTransport
	}
	return tlsOnlyTransport{next: remote.DefaultTransport}
}

type tlsOnlyTransport struct {
	next http.RoundTripper
}

func (t tlsOnlyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" {
		return nil, fmt.Errorf("refusing to reach registry %s over plain HTTP", request.URL.Host)
	}
	return t.next.RoundTrip(request)
}

// Manifest returns the manifest that reference, a tag or a digest, names in
// the repository, and its media type. A manifest fetched by digest is checked
// against it. The error wraps fs.ErrNotExist when the repository holds no
// such manifest.
func (r *Repository) Manifest(reference string) ([]byte, string, error) {
	ref, err := r.reference(reference)
	if err != nil {
		return nil, "", err
	}
	descriptor, err := r.puller.Get(context.Background(), ref)
	if err != nil {
		return nil, "", r.failure("manifest", reference, err)
	}
	return descriptor.Manifest, string(descriptor.MediaType), nil
}

// Blob returns the content of the blob with digest, checked against it.
func (r *Repository) Blob(digest string) ([]byte, error) {
	ref, err := r.digestReference(digest)
	if err != nil {
		return nil, err
	}
	layer, err := r.puller.Layer(context.Background(), ref)
	if err != nil {
		return nil, r.failure("blob", digest, err)
	}
	stream, err := layer.Compressed()
	if err != nil {
		return nil, r.failure("blob", digest, err)
	}
	defer func() { _ = stream.Close() }()

	content, err := io.ReadAll(io.LimitReader(stream, r.maxBlobSize+1))
	if err != nil {
		return nil, r.failure("blob", digest, err)
	}
	if int64(len(content)) > r.maxBlobSize {
		return nil, fmt.Errorf("blob %s in %s is larger than the %d bytes read for an attestation", digest, r.repo, r.maxBlobSize)
	}
	return content, nil
}

// Referrers lists the manifests in the repository whose subject is the
// manifest with digest: from the registry's Referrers API, or from the
// subject's fallback tag when the registry has none. It is empty when
// nothing refers to the manifest.
func (r *Repository) Referrers(digest string) ([]ocispec.Descriptor, error) {
	ref, err := r.digestReference(digest)
	if err != nil {
		return nil, err
	}
	index, err := remote.Referrers(ref, remote.Reuse(r.puller))
	if err != nil {
		return nil, r.failure("referrers of", digest, err)
	}
	manifest, err := index.IndexManifest()
	if err != nil {
		return nil, r.failure("referrers of", digest, err)
	}

	referrers := make([]ocispec.Descriptor, 0, len(manifest.Manifests))
	for _, descriptor := range manifest.Manifests {
		referrers = append(referrers, ocispec.Descriptor{
			MediaType:    string(descriptor.MediaType),
			Digest:       godigest.Digest(descriptor.Digest.String()),
			Size:         descriptor.Size,
			ArtifactType: descriptor.ArtifactType,
			Annotations:  descriptor.Annotations,
		})
	}
	return referrers, nil
}

func (r *Repository) reference(reference string) (name.Reference, error) {
	if strings.Contains(reference, ":") {
		return r.digestReference(reference)
	}
	tag, err := name.NewTag(r.repo.String() + ":" + reference)
	if err != nil {
		return nil, fmt.Errorf("%q is not a tag %s can hold: %w", reference, r.repo, err)
	}
	return tag, nil
}

func (r *Repository) digestReference(digest string) (name.Digest, error) {
	if _, err := godigest.Parse(digest); err != nil {
		return name.Digest{}, fmt.Errorf("%q is not a digest: %w", digest, err)
	}
	return r.repo.Digest(digest), nil
}

func (r *Repository) failure(kind, reference string, err error) error {
	var response *transport.Error
	if errors.As(err, &response) && response.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%s holds no %s %s: %w", r.repo, kind, reference, fs.ErrNotExist)
	}
	return fmt.Errorf("failed to read %s %s from %s: %w", kind, reference, r.repo, err)
}
