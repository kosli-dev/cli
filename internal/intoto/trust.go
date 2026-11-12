package intoto

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

const (
	publicGoodIssuerOrg = "sigstore.dev"
	githubIssuerOrg     = "GitHub, Inc."
	githubTUFMirror     = "https://tuf-repo.github.com"
)

// githubTUFRoot is the root of trust for GitHub's TUF repository, as the GitHub
// CLI embeds it. TUF fetches every later root version from the mirror and
// checks each against the one before, so an old root here is still a sound
// starting point.
//
//go:embed github-tuf-root.json
var githubTUFRoot []byte

// TrustRoots supplies the Sigstore trust roots signatures are checked against.
type TrustRoots struct {
	fixed *root.TrustedRoot

	publicGood cachedRoot
	github     cachedRoot
}

type cachedRoot struct {
	once  sync.Once
	fetch func() (*root.TrustedRoot, error)
	root  *root.TrustedRoot
	err   error
}

func (c *cachedRoot) get() (*root.TrustedRoot, error) {
	c.once.Do(func() { c.root, c.err = c.fetch() })
	return c.root, c.err
}

// SigstoreTrustRoots fetches trust roots over TUF the first time a signature
// needs one: Sigstore's public-good instance, and GitHub's own, which signs
// attestations for private repositories. Nothing is fetched for unsigned
// attestations or for ones signed with a key and carrying no timestamps.
func SigstoreTrustRoots() *TrustRoots {
	return &TrustRoots{
		publicGood: cachedRoot{fetch: func() (*root.TrustedRoot, error) {
			return root.FetchTrustedRootWithOptions(tuf.DefaultOptions())
		}},
		github: cachedRoot{fetch: fetchGitHubTrustedRoot},
	}
}

// FixedTrustRoots checks every signature against one trusted root, which must
// then hold every authority a signature relies on.
func FixedTrustRoots(trustedRoot *root.TrustedRoot) *TrustRoots {
	return &TrustRoots{fixed: trustedRoot}
}

func fetchGitHubTrustedRoot() (*root.TrustedRoot, error) {
	opts := tuf.DefaultOptions()
	opts.Root = githubTUFRoot
	opts.RepositoryBaseURL = githubTUFMirror
	client, err := tuf.New(opts)
	if err != nil {
		return nil, err
	}
	content, err := client.GetTarget("trusted_root.json")
	if err != nil {
		return nil, err
	}
	return root.NewTrustedRootFromJSON(content)
}

// forCertificate returns the trust root for a signature whose Fulcio
// certificate was issued by issuerOrg, and the evidence that instance's
// signatures must carry: public-good signatures are logged in Rekor, while
// GitHub's are timestamped by its own authority instead.
func (t *TrustRoots) forCertificate(issuerOrg string) (root.TrustedMaterial, []verify.VerifierOption, error) {
	if t.fixed != nil {
		return t.fixed, fixedRootOptions(t.fixed), nil
	}
	switch issuerOrg {
	case publicGoodIssuerOrg:
		trustedRoot, err := t.publicGood.get()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch Sigstore's public-good trust root: %w", err)
		}
		return trustedRoot, []verify.VerifierOption{
			verify.WithSignedCertificateTimestamps(1),
			verify.WithTransparencyLog(1),
			verify.WithObserverTimestamps(1),
		}, nil
	case githubIssuerOrg:
		trustedRoot, err := t.github.get()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch GitHub's Sigstore trust root: %w", err)
		}
		return trustedRoot, []verify.VerifierOption{verify.WithSignedTimestamps(1)}, nil
	}
	return nil, nil, fmt.Errorf("the signing certificate was issued by %q, which is neither Sigstore's public-good instance nor GitHub's", issuerOrg)
}

// forKey returns the trust root for the transparency log entries and
// timestamps carried by a signature made with a key.
func (t *TrustRoots) forKey() (*root.TrustedRoot, error) {
	if t.fixed != nil {
		return t.fixed, nil
	}
	trustedRoot, err := t.publicGood.get()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Sigstore's public-good trust root: %w", err)
	}
	return trustedRoot, nil
}

// fixedRootOptions asks for the evidence a fixed root can check: a log entry
// when it names a log, and a trusted time from the log or a timestamp
// authority either way.
func fixedRootOptions(trustedRoot *root.TrustedRoot) []verify.VerifierOption {
	opts := []verify.VerifierOption{verify.WithObserverTimestamps(1)}
	if len(trustedRoot.RekorLogs()) > 0 {
		opts = append(opts, verify.WithTransparencyLog(1))
	}
	return opts
}
