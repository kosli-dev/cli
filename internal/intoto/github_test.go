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
// download URL, both, or neither.
type stored struct {
	fixture string
	inline  bool
	blob    bool
	// gone lists a download URL the blob store no longer answers.
	gone bool
	// listedAs is the predicate type GitHub filters the attestation by, for a
	// fixture that is not an attestation and so states none.
	listedAs string
}

func inline(fixture string) stored { return stored{fixture: fixture, inline: true} }
func blob(fixture string) stored   { return stored{fixture: fixture, blob: true} }
func gone(fixture string) stored   { return stored{fixture: fixture, blob: true, gone: true} }

// notABundle is listed among the provenance attestations, though what GitHub
// returns for it is not an attestation.
func notABundle() stored {
	return stored{fixture: "trusted_root.json", inline: true, listedAs: slsaProvenanceV1}
}

const (
	listsProvenanceV1  = "per_page=100&predicate_type=https%3A%2F%2Fslsa.dev%2Fprovenance%2Fv1"
	listsProvenanceV02 = "per_page=100&predicate_type=https%3A%2F%2Fslsa.dev%2Fprovenance%2Fv0.2"
	listsSPDX          = "per_page=100&predicate_type=https%3A%2F%2Fspdx.dev%2FDocument"
	listsCycloneDX     = "per_page=100&predicate_type=https%3A%2F%2Fcyclonedx.org%2Fbom"
	listsEverything    = "per_page=100"
)

// fakeGitHub serves the attestations API for acme/widget, one page of the
// listing per element of pages, and the blob store bundles are downloaded from.
//
// It filters by predicate_type as github.com was seen to in November 2026,
// which GitHub's documentation does not describe: the filter matches an
// attestation whose predicate type starts with it, and a filter matching
// nothing is answered with 404 rather than an empty listing.
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

	if name, ok := strings.CutPrefix(r.URL.Path, "/blobs/"); ok {
		f.blobAuth = append(f.blobAuth, r.Header.Get("Authorization"))
		if f.blobFailures > 0 {
			f.blobFailures--
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		fixture, held := f.blobFixture(name)
		if !held {
			w.WriteHeader(http.StatusNotFound)
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

	pages := f.pages
	if filter := r.URL.Query().Get("predicate_type"); filter != "" {
		pages = f.matching(filter)
		if len(pages) == 0 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"message": "Not Found"}`)
			return
		}
	}

	page := 0
	if after := r.URL.Query().Get("after"); after != "" {
		_, err := fmt.Sscanf(after, "cursor-%d", &page)
		require.NoError(f.t, err)
	}
	if page+1 < len(pages) {
		w.Header().Set("Link", fmt.Sprintf(`<%s%s?per_page=100&after=cursor-%d>; rel="next"`, f.server.URL, githubListPath, page+1))
	}

	type listed struct {
		Bundle       json.RawMessage `json:"bundle,omitempty"`
		BundleURL    string          `json:"bundle_url,omitempty"`
		RepositoryID int64           `json:"repository_id"`
	}
	attestations := []listed{}
	if page < len(pages) {
		for _, s := range pages[page] {
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

// matching is the pages of the listing left when it is filtered by a predicate
// type.
func (f *fakeGitHub) matching(filter string) [][]stored {
	var pages [][]stored
	for _, page := range f.pages {
		var matched []stored
		for _, s := range page {
			if strings.HasPrefix(f.predicateType(s), filter) {
				matched = append(matched, s)
			}
		}
		if len(matched) > 0 {
			pages = append(pages, matched)
		}
	}
	return pages
}

// predicateType is the type GitHub filters an attestation by: the one its
// statement states.
func (f *fakeGitHub) predicateType(s stored) string {
	if s.listedAs != "" {
		return s.listedAs
	}
	var bundle struct {
		DSSEEnvelope struct {
			Payload []byte `json:"payload"`
		} `json:"dsseEnvelope"`
	}
	var statement struct {
		PredicateType string `json:"predicateType"`
	}
	if json.Unmarshal(f.bundle(s.fixture), &bundle) != nil || json.Unmarshal(bundle.DSSEEnvelope.Payload, &statement) != nil {
		return ""
	}
	return statement.PredicateType
}

// blobFixture is the fixture the blob store holds under name, so a download
// reads only what a test stored rather than whatever path the request names.
func (f *fakeGitHub) blobFixture(name string) (string, bool) {
	for _, page := range f.pages {
		for _, s := range page {
			if s.blob && !s.gone && s.fixture == name {
				return s.fixture, true
			}
		}
	}
	return "", false
}

func (f *fakeGitHub) bundle(fixture string) []byte {
	content, err := os.ReadFile(filepath.Join("testdata", fixture))
	require.NoError(f.t, err)
	return compact(f.t, content)
}

func TestGitHubRepositoryAttestations(t *testing.T) {
	t.Run("lists the artifact's attestations of the kind wanted with the token and reads the bundles inline", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("spdx-sbom.sigstore.json"), inline("github-provenance.sigstore.json"), inline("slsa-v0.2-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 2)
		require.Equal(t, "https://slsa.dev/provenance/v1", attestations[0].Statement.GetPredicateType())
		require.Equal(t, "https://slsa.dev/provenance/v0.2", attestations[1].Statement.GetPredicateType())
		require.True(t, attestations[0].Signed())
		require.Equal(t, []string{"Bearer a-token", "Bearer a-token"}, fake.listAuth)
		require.Equal(t, []string{listsProvenanceV1, listsProvenanceV02}, fake.listQuery)
	})

	t.Run("lists SBOMs by the predicate type each format starts with, whatever version follows", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("cyclonedx-sbom.sigstore.json"), inline("github-provenance.sigstore.json"), inline("spdx-sbom.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindSBOM, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 2)
		require.Equal(t, "https://spdx.dev/Document/v2.3", attestations[0].Statement.GetPredicateType())
		require.Equal(t, "https://cyclonedx.org/bom", attestations[1].Statement.GetPredicateType())
		require.Equal(t, []string{listsSPDX, listsCycloneDX}, fake.listQuery)
	})

	t.Run("downloads no bundle of another kind", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json"), blob("rerun-provenance.sigstore.json"), blob("spdx-sbom.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindSBOM, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Len(t, fake.blobAuth, 1)
	})

	t.Run("lists everything GitHub holds when it holds none of the kind, so that a caller can say what it holds instead", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("spdx-sbom.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Equal(t, "https://spdx.dev/Document/v2.3", attestations[0].Statement.GetPredicateType())
		require.Equal(t, []string{listsProvenanceV1, listsProvenanceV02, listsEverything}, fake.listQuery)
	})

	t.Run("keeps each bundle's bytes, numbers them, and says where GitHub holds them", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), inline("rerun-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Equal(t, fake.bundle("github-provenance.sigstore.json"), attestations[0].Raw)
		require.Equal(t, 1, attestations[0].Position)
		require.Equal(t, 2, attestations[1].Position)
		require.Equal(t, fake.server.URL+githubListPath, attestations[0].URL)
	})

	t.Run("does not number the only attestation", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Equal(t, 0, attestations[0].Position)
	})

	t.Run("accepts a digest with its algorithm prefix", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, "sha256:"+strings.ToUpper(artifactDigest))
		require.NoError(t, err)
		require.Len(t, attestations, 1)
	})

	t.Run("follows the listing through every page", func(t *testing.T) {
		fake := newFakeGitHub(t,
			[]stored{inline("github-provenance.sigstore.json")},
			[]stored{inline("rerun-provenance.sigstore.json")},
			[]stored{inline("branch-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 3)
		require.Equal(t, []string{listsProvenanceV1, "after=cursor-1&" + listsProvenanceV1, "after=cursor-2&" + listsProvenanceV1, listsProvenanceV02}, fake.listQuery)
	})

	t.Run("downloads a bundle GitHub lists by URL, without sending the token there", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Equal(t, fake.bundle("github-provenance.sigstore.json"), attestations[0].Raw)
		require.Equal(t, []string{""}, fake.blobAuth)
	})

	t.Run("uses the inline bundle when GitHub gives both", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{{fixture: "github-provenance.sigstore.json", inline: true, blob: true}})

		_, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Empty(t, fake.blobAuth)
	})

	t.Run("cannot read a downloaded bundle that is not snappy-compressed, as the GitHub CLI cannot", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobBody = func(bundle []byte) []byte { return bundle }

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.EqualError(t, attestations[0].Unreadable, "the attestation: its bundle could not be downloaded from "+fake.server.URL+"/blobs/github-provenance.sigstore.json: it is not snappy-compressed")
	})

	t.Run("cannot read a downloaded bundle larger than the limit, naming the limit", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobBody = func([]byte) []byte { return make([]byte, 32<<20+1) }

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.ErrorContains(t, attestations[0].Unreadable, "/blobs/github-provenance.sigstore.json: it is larger than the 32 MiB a bundle may be")
	})

	t.Run("cannot read a downloaded bundle that would decompress to more than the limit", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobBody = func([]byte) []byte { return snappy.Encode(nil, make([]byte, 32<<20+1)) }

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.ErrorContains(t, attestations[0].Unreadable, "/blobs/github-provenance.sigstore.json: it would decompress to more than the 32 MiB a bundle may be")
	})

	t.Run("retries a download the blob store fails", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobFailures = 2

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Len(t, fake.blobAuth, 3)
	})

	t.Run("gives up on a download that keeps failing, without printing the URL's signature", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{blob("github-provenance.sigstore.json")})
		fake.blobFailures = 10

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.ErrorContains(t, attestations[0].Unreadable, "its bundle could not be downloaded from "+fake.server.URL+"/blobs/github-provenance.sigstore.json: ")
		require.NotContains(t, attestations[0].Unreadable.Error(), "a-secret")
		require.Len(t, fake.blobAuth, 4)
	})

	t.Run("keeps an attestation whose bundle it cannot download, saying why, beside the ones it can", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{gone("rerun-provenance.sigstore.json"), blob("github-provenance.sigstore.json")})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 2)
		require.EqualError(t, attestations[0].Unreadable, "attestation 1: its bundle could not be downloaded from "+fake.server.URL+"/blobs/rerun-provenance.sigstore.json: 404 Not Found")
		require.NoError(t, attestations[1].Unreadable)
		require.Equal(t, fake.bundle("github-provenance.sigstore.json"), attestations[1].Raw)
	})

	t.Run("keeps an attestation GitHub lists with neither a bundle nor a URL, beside the ones it can read", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), {fixture: "rerun-provenance.sigstore.json"}})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, attestations, 2)
		require.NoError(t, attestations[0].Unreadable)
		require.EqualError(t, attestations[1].Unreadable, "attestation 2: GitHub lists neither a bundle nor a URL to download one from")
	})

	t.Run("says when GitHub holds no attestation for the artifact", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{})

		_, err := fake.repository().Attestations(KindProvenance, artifactDigest)
		require.EqualError(t, err, "it holds no attestation for the artifact (sha256:"+artifactDigest+")")
	})

	t.Run("names both causes when GitHub answers the listing with not found", func(t *testing.T) {
		fake := newFakeGitHub(t)
		repository := fake.repository()
		repository.Name = "gadget"

		_, err := repository.Attestations(KindProvenance, artifactDigest)
		require.ErrorContains(t, err, "failed to list its attestations: ")
		require.ErrorContains(t, err, "404 Not Found")
		require.ErrorContains(t, err, "Either GitHub holds no attestation for the artifact, or the token cannot read the repository and its attestations (the attestations: read permission)")
	})

	t.Run("keeps an attestation whose bundle it cannot read, saying why, beside the ones it can", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{inline("github-provenance.sigstore.json"), notABundle()})

		attestations, err := fake.repository().Attestations(KindProvenance, artifactDigest)
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
		fake := newFakeGitHub(t, []stored{notABundle(), inline("github-provenance.sigstore.json")})

		found, err := find(t, fake.repository(), KindProvenance)
		require.NoError(t, err)
		require.Equal(t, 2, found.Position)
		require.True(t, found.Evidence.Verified)
		require.Len(t, found.Rejected, 1)
		require.ErrorContains(t, found.Rejected[0], "attestation 1: not a Sigstore bundle, a DSSE envelope or an in-toto statement")
	})

	t.Run("passes over an attestation whose bundle it cannot download when another verifies, and reports it", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{gone("wrong-signer-provenance.sigstore.json"), blob("github-provenance.sigstore.json")})

		found, err := find(t, fake.repository(), KindProvenance)
		require.NoError(t, err)
		require.Equal(t, 2, found.Position)
		require.True(t, found.Evidence.Verified)
		require.Len(t, found.Rejected, 1)
		require.ErrorContains(t, found.Rejected[0], "attestation 1: its bundle could not be downloaded from ")
	})

	t.Run("names the attestation whose bundle it cannot download when none verifies", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{gone("github-provenance.sigstore.json"), inline("wrong-signer-provenance.sigstore.json")})

		_, err := find(t, fake.repository(), KindProvenance)
		require.ErrorIs(t, err, ErrFailedVerification)
		require.ErrorContains(t, err, "GitHub repository acme/widget: 2 attestations could not be accepted: attestation 1: its bundle could not be downloaded from "+fake.server.URL+"/blobs/github-provenance.sigstore.json: 404 Not Found; attestation 2 failed verification: ")
	})

	t.Run("names every attestation it cannot read or verify when none verifies", func(t *testing.T) {
		fake := newFakeGitHub(t, []stored{notABundle(), inline("wrong-signer-provenance.sigstore.json")})

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

func TestGitHubDefaultSigner(t *testing.T) {
	repository := GitHubRepository{Owner: "acme", Name: "widget.js", RunURI: "https://github.com/acme/widget.js/actions/runs/77/attempts/1"}
	identityRegexp := func(t *testing.T) string {
		policy, err := repository.DefaultSigner()
		require.NoError(t, err)
		return policy.CertificateIdentityRegexp
	}

	t.Run("pins the GitHub Actions issuer and the repository's own workflows", func(t *testing.T) {
		policy, err := repository.DefaultSigner()
		require.NoError(t, err)
		require.Equal(t, "https://token.actions.githubusercontent.com", policy.CertificateOIDCIssuer)
		require.Regexp(t, policy.CertificateIdentityRegexp, "https://github.com/acme/widget.js/.github/workflows/build.yml@refs/heads/main")
		require.NoError(t, policy.Validate())
	})

	t.Run("is tied to the run the caller is part of, as the identity matches a workflow on any branch or pull request", func(t *testing.T) {
		policy, err := repository.DefaultSigner()
		require.NoError(t, err)
		require.Regexp(t, policy.CertificateIdentityRegexp, "https://github.com/acme/widget.js/.github/workflows/build.yml@refs/heads/feature-x")
		require.Regexp(t, policy.CertificateIdentityRegexp, "https://github.com/acme/widget.js/.github/workflows/build.yml@refs/pull/12/merge")
		require.Equal(t, repository.RunURI, policy.RunInvocationURI)
	})

	t.Run("does not exist for a caller that is part of no run", func(t *testing.T) {
		_, err := GitHubRepository{Owner: "acme", Name: "widget.js"}.DefaultSigner()
		require.EqualError(t, err, "GitHub repository acme/widget.js has no default signer outside a GitHub Actions run to tie it to")
	})

	t.Run("matches the repository's name in any case, as GitHub does", func(t *testing.T) {
		require.Regexp(t, identityRegexp(t), "https://github.com/Acme/Widget.JS/.github/workflows/build.yml@refs/heads/main")
	})

	for name, identity := range map[string]string{
		"a workflow in another repository":               "https://github.com/acme/other/.github/workflows/build.yml@refs/heads/main",
		"a repository whose name only starts the same":   "https://github.com/acme/widget.js-fork/.github/workflows/build.yml@refs/heads/main",
		"a repository whose name differs where a dot is": "https://github.com/acme/widgetxjs/.github/workflows/build.yml@refs/heads/main",
		"the repository's name later in another URL":     "https://example.com/https://github.com/acme/widget.js/.github/workflows/build.yml",
	} {
		t.Run("does not match "+name, func(t *testing.T) {
			require.NotRegexp(t, identityRegexp(t), identity)
		})
	}
}

func TestFindInGitHubRepositoryUnderTheDefaultSigner(t *testing.T) {
	// Newest last: the build on main, then what a branch and a pull request
	// attested about the same artifact afterwards.
	held := []stored{inline("github-provenance.sigstore.json"), inline("branch-provenance.sigstore.json"), inline("pull-request-provenance.sigstore.json")}
	find := func(t *testing.T, runURI string) (*Found, error) {
		repository := newFakeGitHub(t, held).repository()
		repository.RunURI = runURI
		signer, err := repository.DefaultSigner()
		require.NoError(t, err)
		return Find(repository, KindProvenance, artifactDigest, NewVerifier(signer, fixtureTrust(t)))
	}

	t.Run("records what the run it is part of attested, not the newer attestations other runs made", func(t *testing.T) {
		found, err := find(t, fixtureRunURI)
		require.NoError(t, err)
		require.Equal(t, fixtureIdentity, found.Evidence.SignerIdentity)
		require.Equal(t, fixtureRunURI, found.Evidence.RunInvocationURI)
		require.Empty(t, found.Rejected)
	})

	t.Run("records what an earlier attempt of the run attested", func(t *testing.T) {
		found, err := find(t, fixtureRerunURI)
		require.NoError(t, err)
		require.Equal(t, fixtureRunURI, found.Evidence.RunInvocationURI)
	})

	t.Run("records nothing from another run, such as a deployment, though the repository's own workflows signed all three", func(t *testing.T) {
		_, err := find(t, "https://github.com/acme/widget/actions/runs/77/attempts/1")
		require.ErrorIs(t, err, ErrOtherRun)
		require.ErrorContains(t, err, "GitHub repository acme/widget: 3 attestations could not be accepted: attestation 1 was made by another run: "+fixtureRunURI+", signed as "+fixtureIdentity+"; attestation 2 was made by another run: ")
	})
}
