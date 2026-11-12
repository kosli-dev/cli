package intoto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/klauspost/compress/snappy"
	"github.com/stretchr/testify/require"
)

var _ Selector = GitHubRepository{}

const githubListPath = "/api/v3/repos/acme/widget/attestations/sha256:" + artifactDigest

// stored is one attestation as GitHub's API lists it: the bundle inline, at a
// download URL, or both.
type stored struct {
	fixture string
	inline  bool
	blob    bool
}

func inline(fixture string) stored { return stored{fixture: fixture, inline: true} }
func blob(fixture string) stored   { return stored{fixture: fixture, blob: true} }

// fakeGitHub serves the attestations API for acme/widget, one page of the
// listing per element of pages, and the blob store bundles are downloaded from.
type fakeGitHub struct {
	t      *testing.T
	server *httptest.Server
	pages  [][]stored
	// blobBody replaces how a downloaded bundle is encoded, which is
	// snappy-compressed JSON unless set.
	blobBody func(bundle []byte) []byte
	// blobFailures is how many downloads fail with a 503 before one succeeds.
	blobFailures int

	mu        sync.Mutex
	listAuth  []string
	listQuery []string
	blobAuth  []string
}

func newFakeGitHub(t *testing.T, pages ...[]stored) *fakeGitHub {
	t.Helper()
	fake := &fakeGitHub{t: t, pages: pages}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(fake.server.Close)

	saved := bundleRetryWait
	bundleRetryWait = time.Millisecond
	t.Cleanup(func() { bundleRetryWait = saved })
	return fake
}

func (f *fakeGitHub) repository() GitHubRepository {
	return GitHubRepository{Owner: "acme", Name: "widget", Token: "a-token", BaseURL: f.server.URL + "/"}
}

func (f *fakeGitHub) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if fixture, ok := strings.CutPrefix(r.URL.Path, "/blobs/"); ok {
		f.blobAuth = append(f.blobAuth, r.Header.Get("Authorization"))
		if f.blobFailures > 0 {
			f.blobFailures--
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		bundle := f.bundle(fixture)
		if f.blobBody != nil {
			_, _ = w.Write(f.blobBody(bundle))
			return
		}
		_, _ = w.Write(snappy.Encode(nil, bundle))
		return
	}

	if r.URL.Path != githubListPath {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message": "Not Found"}`)
		return
	}
	f.listAuth = append(f.listAuth, r.Header.Get("Authorization"))
	f.listQuery = append(f.listQuery, r.URL.RawQuery)

	page := 0
	if after := r.URL.Query().Get("after"); after != "" {
		_, err := fmt.Sscanf(after, "cursor-%d", &page)
		require.NoError(f.t, err)
	}
	if page+1 < len(f.pages) {
		w.Header().Set("Link", fmt.Sprintf(`<%s%s?per_page=100&after=cursor-%d>; rel="next"`, f.server.URL, githubListPath, page+1))
	}

	type listed struct {
		Bundle       json.RawMessage `json:"bundle,omitempty"`
		BundleURL    string          `json:"bundle_url,omitempty"`
		RepositoryID int64           `json:"repository_id"`
	}
	attestations := []listed{}
	if page < len(f.pages) {
		for _, s := range f.pages[page] {
			entry := listed{RepositoryID: 123456789}
			if s.inline {
				entry.Bundle = f.bundle(s.fixture)
			}
			if s.blob {
				entry.BundleURL = f.server.URL + "/blobs/" + s.fixture + "?sig=a-secret"
			}
			attestations = append(attestations, entry)
		}
	}
	require.NoError(f.t, json.NewEncoder(w).Encode(map[string]any{"attestations": attestations}))
}

func (f *fakeGitHub) bundle(fixture string) []byte {
	content, err := os.ReadFile(filepath.Join("testdata", fixture))
	require.NoError(f.t, err)
	return compact(f.t, content)
}

func TestGitHubRepositoryAttestations(t *testing.T) {
	t.Run("lists the artifact's attestations with the token and reads the bundles inline", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), inline("spdx-sbom.sigstore.json")})

		attestations, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 2)
		require.Equal(t, "https://slsa.dev/provenance/v1", attestations[0].Statement.GetPredicateType())
		require.Equal(t, "https://spdx.dev/Document/v2.3", attestations[1].Statement.GetPredicateType())
		require.True(t, attestations[0].Signed())
		require.Equal(t, []string{"Bearer a-token"}, fake.listAuth)
		require.Equal(t, []string{"per_page=100"}, fake.listQuery)
	})

	t.Run("keeps each bundle's bytes, numbers them, and says where GitHub holds them", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), inline("spdx-sbom.sigstore.json")})

		attestations, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Equal(t, fake.bundle("github-provenance.sigstore.json"), attestations[0].Raw)
		require.Equal(t, 1, attestations[0].Position)
		require.Equal(t, 2, attestations[1].Position)
		require.Equal(t, fake.server.URL+githubListPath, attestations[0].URL)
	})

	t.Run("does not number the only attestation", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Equal(t, 0, attestations[0].Position)
	})

	t.Run("accepts a digest with its algorithm prefix", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations("sha256:" + strings.ToUpper(artifactDigest))
		require.NoError(t, err)
		require.Len(t, attestations, 1)
	})

	t.Run("follows the listing through every page", func(t *testing.T) {
		fake := newFakeGitHub(t,
			[]stored{inline("cyclonedx-sbom.sigstore.json")},
			[]stored{inline("github-provenance.sigstore.json")},
			[]stored{inline("spdx-sbom.sigstore.json")})

		attestations, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 3)
		require.Equal(t, []string{"per_page=100", "after=cursor-1&per_page=100", "after=cursor-2&per_page=100"}, fake.listQuery)
	})

	t.Run("downloads a bundle GitHub lists by URL, without sending the token there", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Equal(t, fake.bundle("github-provenance.sigstore.json"), attestations[0].Raw)
		require.Equal(t, []string{""}, fake.blobAuth)
	})

	t.Run("uses the inline bundle when GitHub gives both", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{{fixture: "github-provenance.sigstore.json", inline: true, blob: true}})

		_, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Empty(t, fake.blobAuth)
	})

	t.Run("refuses a downloaded bundle that is not snappy-compressed, as the GitHub CLI does", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobBody = func(bundle []byte) []byte { return bundle }

		_, err := fake.repository().Attestations(artifactDigest)
		require.ErrorContains(t, err, "/blobs/github-provenance.sigstore.json: it is not snappy-compressed")
	})

	t.Run("retries a download the blob store fails", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobFailures = 2

		attestations, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Len(t, fake.blobAuth, 3)
	})

	t.Run("gives up on a download that keeps failing, without printing the URL's signature", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobFailures = 10

		_, err := fake.repository().Attestations(artifactDigest)
		require.ErrorContains(t, err, "failed to download the bundle of the attestation from "+fake.server.URL+"/blobs/github-provenance.sigstore.json: ")
		require.NotContains(t, err.Error(), "a-secret")
		require.Len(t, fake.blobAuth, 4)
	})

	t.Run("says when GitHub holds no attestation for the artifact", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{})

		_, err := fake.repository().Attestations(artifactDigest)
		require.EqualError(t, err, "it holds no attestation for the artifact (sha256:"+artifactDigest+")")
	})

	t.Run("names both causes when GitHub answers the listing with not found", func(t *testing.T) {
		fake := newFakeGitHub(t)
		repository := fake.repository()
		repository.Name = "gadget"

		_, err := repository.Attestations(artifactDigest)
		require.ErrorContains(t, err, "failed to list its attestations: ")
		require.ErrorContains(t, err, "404 Not Found")
		require.ErrorContains(t, err, "Either GitHub holds no attestation for the artifact, or the token cannot read the repository and its attestations (the attestations: read permission)")
	})

	t.Run("keeps an attestation whose bundle it cannot read, saying why, beside the ones it can", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), inline("trusted_root.json")})

		attestations, err := fake.repository().Attestations(artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 2)
		require.NoError(t, attestations[0].Unreadable)
		require.EqualError(t, attestations[1].Unreadable, "attestation 2: not a Sigstore bundle, a DSSE envelope or an in-toto statement")
	})

	t.Run("names itself by the repository", func(t *testing.T) {
		require.Equal(t, "GitHub repository acme/widget", GitHubRepository{Owner: "acme", Name: "widget"}.String())
	})
}

func TestFindInGitHubRepository(t *testing.T) {
	find := func(t *testing.T, repository GitHubRepository, kind PredicateKind) (*Found, error) {
		return Find(repository, kind, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
	}

	t.Run("finds and verifies the attestation of the kind asked for", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), blob("spdx-sbom.sigstore.json")})

		found, err := find(t, fake.repository(), KindSBOM)
		require.NoError(t, err)
		require.True(t, found.Evidence.Verified)
		require.Equal(t, "https://spdx.dev/Document/v2.3", found.Statement.GetPredicateType())
	})

	t.Run("records the pinned workflow's attestation though another workflow attested the artifact too", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("wrong-signer-provenance.sigstore.json"), inline("github-provenance.sigstore.json")})

		found, err := find(t, fake.repository(), KindProvenance)
		require.NoError(t, err)
		require.Equal(t, fixtureIdentity, found.Evidence.SignerIdentity)
		require.Len(t, found.Rejected, 1)
	})

	t.Run("prefers the attestation made by the run it is part of", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), inline("rerun-provenance.sigstore.json")})
		repository := fake.repository()
		repository.RunURI = fixtureRunURI

		found, err := find(t, repository, KindProvenance)
		require.NoError(t, err)
		require.Equal(t, fixtureRunURI, found.Evidence.RunInvocationURI)
	})

	t.Run("passes over an attestation it cannot read when another verifies, and reports it", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("trusted_root.json"), inline("github-provenance.sigstore.json")})

		found, err := find(t, fake.repository(), KindProvenance)
		require.NoError(t, err)
		require.Equal(t, 2, found.Position)
		require.True(t, found.Evidence.Verified)
		require.Len(t, found.Rejected, 1)
		require.ErrorContains(t, found.Rejected[0], "attestation 1: not a Sigstore bundle, a DSSE envelope or an in-toto statement")
	})

	t.Run("names every attestation it cannot read or verify when none verifies", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("trusted_root.json"), inline("wrong-signer-provenance.sigstore.json")})

		_, err := find(t, fake.repository(), KindProvenance)
		require.ErrorIs(t, err, ErrFailedVerification)
		require.ErrorContains(t, err, "GitHub repository acme/widget: 2 attestations could not be accepted: attestation 1: not a Sigstore bundle, a DSSE envelope or an in-toto statement; attestation 2 failed verification: ")
	})

	t.Run("names the attestation it cannot read beside the kinds it holds when none is the kind asked for", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("spdx-sbom.sigstore.json"), inline("trusted_root.json")})

		_, err := find(t, fake.repository(), KindProvenance)
		require.EqualError(t, err, "GitHub repository acme/widget: found no provenance attestation (SLSA provenance v1 or v0.2); the predicate types found were: https://spdx.dev/Document/v2.3; attestation 2: not a Sigstore bundle, a DSSE envelope or an in-toto statement")
	})

	t.Run("says why when it can read nothing GitHub holds", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("trusted_root.json")})

		_, err := find(t, fake.repository(), KindProvenance)
		require.EqualError(t, err, "GitHub repository acme/widget: the attestation: not a Sigstore bundle, a DSSE envelope or an in-toto statement")
	})

	t.Run("names the repository and the predicate types it holds when none is the kind asked for", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("spdx-sbom.sigstore.json")})

		_, err := find(t, fake.repository(), KindProvenance)
		require.EqualError(t, err, "GitHub repository acme/widget: found no provenance attestation (SLSA provenance v1 or v0.2); the predicate types found were: https://spdx.dev/Document/v2.3")
	})
}

func TestGitHubWorkflowSigner(t *testing.T) {
	repository := GitHubRepository{Owner: "acme", Name: "widget.js"}

	t.Run("pins the GitHub Actions issuer and the repository's own workflows", func(t *testing.T) {
		policy := repository.WorkflowSigner()
		require.Equal(t, "https://token.actions.githubusercontent.com", policy.CertificateOIDCIssuer)
		require.Regexp(t, policy.CertificateIdentityRegexp, "https://github.com/acme/widget.js/.github/workflows/build.yml@refs/heads/main")
		require.NoError(t, policy.Validate())
	})

	t.Run("matches the repository's name in any case, as GitHub does", func(t *testing.T) {
		require.Regexp(t, repository.WorkflowSigner().CertificateIdentityRegexp, "https://github.com/Acme/Widget.JS/.github/workflows/build.yml@refs/heads/main")
	})

	for name, identity := range map[string]string{
		"a workflow in another repository":               "https://github.com/acme/other/.github/workflows/build.yml@refs/heads/main",
		"a repository whose name only starts the same":   "https://github.com/acme/widget.js-fork/.github/workflows/build.yml@refs/heads/main",
		"a repository whose name differs where a dot is": "https://github.com/acme/widgetxjs/.github/workflows/build.yml@refs/heads/main",
		"the repository's name later in another URL":     "https://example.com/https://github.com/acme/widget.js/.github/workflows/build.yml",
	} {
		t.Run("does not match "+name, func(t *testing.T) {
			require.NotRegexp(t, repository.WorkflowSigner().CertificateIdentityRegexp, identity)
		})
	}
}
