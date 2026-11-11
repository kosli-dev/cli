package intoto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/stretchr/testify/require"
)

var pinnedSigner = Policy{CertificateIdentity: fixtureIdentity, CertificateOIDCIssuer: fixtureIssuer}

func fixtureTrust(t *testing.T) *TrustRoots {
	t.Helper()
	trustedRoot, err := root.NewTrustedRootFromPath(filepath.Join("testdata", "trusted_root.json"))
	require.NoError(t, err)
	return FixedTrustRoots(trustedRoot)
}

func readOne(t *testing.T, path string) *Attestation {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	attestations, err := Read(content)
	require.NoError(t, err)
	require.Len(t, attestations, 1)
	return attestations[0]
}

func verifyFixture(t *testing.T, policy Policy, name string) (*Evidence, error) {
	t.Helper()
	return NewVerifier(policy, fixtureTrust(t)).Verify(readOne(t, filepath.Join("testdata", name)), artifactDigest)
}

func TestVerifySignedBundle(t *testing.T) {
	t.Run("verifies a bundle signed by the pinned identity and records the evidence", func(t *testing.T) {
		evidence, err := verifyFixture(t, pinnedSigner, "github-provenance.sigstore.json")
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join("testdata", "github-provenance.sigstore.json"))
		require.NoError(t, err)
		sum := sha256.Sum256(content)

		require.True(t, evidence.Verified)
		require.Equal(t, fixtureIdentity, evidence.SignerIdentity)
		require.Equal(t, fixtureIssuer, evidence.SignerIssuer)
		require.NotNil(t, evidence.LogIndex)
		require.Equal(t, int64(0), *evidence.LogIndex)
		require.False(t, evidence.IntegratedTime.IsZero())
		require.False(t, evidence.SignedTime.IsZero())
		require.Equal(t, "https://rekor.localhost/api/v1/log/entries?logIndex=0", evidence.LogEntryURL)
		require.Equal(t, hex.EncodeToString(sum[:]), evidence.RawSHA256)
	})

	t.Run("accepts an identity matched by a regular expression", func(t *testing.T) {
		policy := Policy{CertificateIdentityRegexp: `^https://github\.com/acme/widget/`, CertificateOIDCIssuer: fixtureIssuer}
		_, err := verifyFixture(t, policy, "github-provenance.sigstore.json")
		require.NoError(t, err)
	})

	for _, name := range []string{"buildkit-provenance.sigstore.json", "spdx-sbom.sigstore.json", "cyclonedx-sbom.sigstore.json"} {
		t.Run("verifies "+name, func(t *testing.T) {
			_, err := verifyFixture(t, pinnedSigner, name)
			require.NoError(t, err)
		})
	}

	t.Run("refuses a bundle signed by another identity", func(t *testing.T) {
		_, err := verifyFixture(t, pinnedSigner, "wrong-signer-provenance.sigstore.json")
		require.ErrorContains(t, err, "the attestation failed verification")
		require.ErrorContains(t, err, fixtureOtherIdentity)
	})

	t.Run("refuses a bundle with the wrong OIDC issuer", func(t *testing.T) {
		policy := Policy{CertificateIdentity: fixtureIdentity, CertificateOIDCIssuer: "https://accounts.example.com"}
		_, err := verifyFixture(t, policy, "github-provenance.sigstore.json")
		require.ErrorContains(t, err, "the attestation failed verification")
	})

	t.Run("refuses a bundle whose statement was changed after signing", func(t *testing.T) {
		_, err := verifyFixture(t, pinnedSigner, "tampered-provenance.sigstore.json")
		require.ErrorContains(t, err, "the attestation failed verification")
	})

	t.Run("refuses a signed statement about another artifact", func(t *testing.T) {
		_, err := verifyFixture(t, pinnedSigner, "wrong-subject-provenance.sigstore.json")
		require.EqualError(t, err, "the attestation does not list the artifact (sha256:"+artifactDigest+") as its subject")
	})

	t.Run("refuses a signed bundle when no signer is pinned, naming the flags that pin one", func(t *testing.T) {
		_, err := verifyFixture(t, Policy{}, "github-provenance.sigstore.json")
		require.EqualError(t, err, "the attestation is signed, so its signer must be pinned to verify it. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key")
	})

	t.Run("refuses a certificate signature checked with --key", func(t *testing.T) {
		_, err := verifyFixture(t, Policy{KeyPath: filepath.Join("testdata", "key.pub")}, "github-provenance.sigstore.json")
		require.ErrorContains(t, err, "expected key signature, not certificate")
	})
}

func TestVerifyKeySignedEnvelope(t *testing.T) {
	t.Run("verifies a DSSE envelope with the key that signed it", func(t *testing.T) {
		evidence, err := verifyFixture(t, Policy{KeyPath: filepath.Join("testdata", "key.pub")}, "key-signed-provenance.dsse.json")
		require.NoError(t, err)
		require.True(t, evidence.Verified)
		require.Empty(t, evidence.SignerIdentity)
		require.Nil(t, evidence.LogIndex)
	})

	t.Run("refuses a DSSE envelope checked with another key", func(t *testing.T) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		pem, err := cryptoutils.MarshalPublicKeyToPEM(&key.PublicKey)
		require.NoError(t, err)
		keyPath := filepath.Join(t.TempDir(), "other.pub")
		require.NoError(t, os.WriteFile(keyPath, pem, 0o600))

		_, err = verifyFixture(t, Policy{KeyPath: keyPath}, "key-signed-provenance.dsse.json")
		require.ErrorContains(t, err, "the attestation failed verification")
	})

	t.Run("explains that a DSSE envelope has no certificate to check an identity against", func(t *testing.T) {
		_, err := verifyFixture(t, pinnedSigner, "key-signed-provenance.dsse.json")
		require.EqualError(t, err, "the attestation failed verification: a DSSE envelope carries no certificate to check --certificate-identity against. Supply the Sigstore bundle instead, or verify the envelope with --key")
	})

	t.Run("names a key file that is not a public key", func(t *testing.T) {
		_, err := verifyFixture(t, Policy{KeyPath: filepath.Join("testdata", "artifact.bin")}, "key-signed-provenance.dsse.json")
		require.ErrorContains(t, err, "--key testdata/artifact.bin is not a PEM public key")
	})
}

func TestVerifyUnsigned(t *testing.T) {
	t.Run("refuses an unsigned statement unless unsigned ones are allowed", func(t *testing.T) {
		_, err := verifyFixture(t, Policy{}, filepath.Join("statements", "buildkit-provenance.json"))
		require.EqualError(t, err, "the attestation is not signed. Pass --allow-unsigned to record it as unverified")
	})

	t.Run("accepts an unsigned statement as unverified when allowed", func(t *testing.T) {
		evidence, err := verifyFixture(t, Policy{AllowUnsigned: true}, filepath.Join("statements", "buildkit-provenance.json"))
		require.NoError(t, err)
		require.False(t, evidence.Verified)
		require.True(t, evidence.Time().IsZero())
	})

	t.Run("still checks the subject of an unsigned statement", func(t *testing.T) {
		_, err := verifyFixture(t, Policy{AllowUnsigned: true}, filepath.Join("statements", "wrong-subject-provenance.json"))
		require.ErrorContains(t, err, "does not list the artifact")
	})
}

func TestPolicyValidate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy Policy
		want   string
	}{
		{"identity and identity regexp together", Policy{CertificateIdentity: "a", CertificateIdentityRegexp: "b", CertificateOIDCIssuer: "c"},
			"only one of --certificate-identity, --certificate-identity-regexp is allowed"},
		{"identity without issuer", Policy{CertificateIdentity: "a"},
			"--certificate-oidc-issuer is required with --certificate-identity or --certificate-identity-regexp"},
		{"issuer without identity", Policy{CertificateOIDCIssuer: "c"},
			"--certificate-identity or --certificate-identity-regexp is required with --certificate-oidc-issuer"},
		{"key with identity", Policy{KeyPath: "k", CertificateIdentity: "a", CertificateOIDCIssuer: "c"},
			"--key cannot be combined with --certificate-identity, --certificate-identity-regexp or --certificate-oidc-issuer: a signature is made either with a key or with a certificate"},
		{"invalid regexp", Policy{CertificateIdentityRegexp: "(", CertificateOIDCIssuer: "c"},
			"--certificate-identity-regexp is not a valid regular expression: error parsing regexp: missing closing ): `(`"},
	} {
		t.Run("refuses "+tc.name, func(t *testing.T) {
			require.EqualError(t, tc.policy.Validate(), tc.want)
		})
	}

	t.Run("accepts no policy at all, for unsigned input", func(t *testing.T) {
		require.NoError(t, Policy{}.Validate())
	})
}

// jsonlFile writes the named fixtures, one per line, as an .intoto.jsonl file
// and returns its path. A name under statements/ is an unsigned statement.
func jsonlFile(t *testing.T, names ...string) string {
	t.Helper()
	lines := make([][]byte, len(names))
	for i, name := range names {
		content, err := os.ReadFile(filepath.Join("testdata", name))
		require.NoError(t, err)
		lines[i] = compact(t, content)
	}
	path := filepath.Join(t.TempDir(), "attestations.intoto.jsonl")
	require.NoError(t, os.WriteFile(path, append(bytes.Join(lines, []byte("\n")), '\n'), 0o600))
	return path
}

func TestFind(t *testing.T) {
	find := func(t *testing.T, policy Policy, kind PredicateKind, name string) (*Found, error) {
		return Find(BundleFile{Path: filepath.Join("testdata", name)}, kind, artifactDigest, NewVerifier(policy, fixtureTrust(t)))
	}

	t.Run("finds the attestation of the kind asked for among several", func(t *testing.T) {
		found, err := find(t, pinnedSigner, KindProvenance, "attestations.intoto.jsonl")
		require.NoError(t, err)
		require.Equal(t, 2, found.Position)
		require.True(t, found.Evidence.Verified)
	})

	t.Run("chooses the newest by its logged time, not its place in the file", func(t *testing.T) {
		found, err := find(t, pinnedSigner, KindProvenance, "newest-first.intoto.jsonl")
		require.NoError(t, err)
		require.Equal(t, 1, found.Position)
	})

	t.Run("prefers a verified attestation to an unsigned one", func(t *testing.T) {
		found, err := find(t, Policy{CertificateIdentity: fixtureIdentity, CertificateOIDCIssuer: fixtureIssuer, AllowUnsigned: true}, KindProvenance, "attestations.intoto.jsonl")
		require.NoError(t, err)
		require.True(t, found.Evidence.Verified)
	})

	t.Run("names the predicate types found when none is the kind asked for", func(t *testing.T) {
		_, err := find(t, pinnedSigner, KindProvenance, "vuln-only.intoto.jsonl")
		require.EqualError(t, err, "bundle file testdata/vuln-only.intoto.jsonl: found no provenance attestation (SLSA provenance v1 or v0.2); the predicate types found were: https://cosign.sigstore.dev/attestation/vuln/v1")
	})

	t.Run("refuses the file when a candidate fails verification, though an unsigned one is allowed", func(t *testing.T) {
		_, err := find(t, Policy{CertificateIdentity: fixtureIdentity, CertificateOIDCIssuer: fixtureIssuer, AllowUnsigned: true}, KindProvenance, "wrong-signer-and-unsigned.intoto.jsonl")
		require.ErrorIs(t, err, ErrFailedVerification)
		require.ErrorContains(t, err, "bundle file testdata/wrong-signer-and-unsigned.intoto.jsonl: attestation 1 failed verification: ")
	})

	t.Run("refuses the file when a candidate fails verification, though another verifies", func(t *testing.T) {
		path := jsonlFile(t, "github-provenance.sigstore.json", "tampered-provenance.sigstore.json")

		_, err := Find(BundleFile{Path: path}, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.ErrorIs(t, err, ErrFailedVerification)
		require.ErrorContains(t, err, "attestation 2 failed verification: ")
	})

	t.Run("names every candidate that failed verification", func(t *testing.T) {
		path := jsonlFile(t, "wrong-signer-provenance.sigstore.json", "tampered-provenance.sigstore.json")

		_, err := Find(BundleFile{Path: path}, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.ErrorIs(t, err, ErrFailedVerification)
		require.ErrorContains(t, err, ": 2 provenance attestations failed verification: attestation 1 failed verification: ")
		require.ErrorContains(t, err, "; attestation 2 failed verification: ")
	})

	t.Run("reports a merely unsigned alternative without failing", func(t *testing.T) {
		path := jsonlFile(t, "github-provenance.sigstore.json", filepath.Join("statements", "github-provenance.json"))

		found, err := Find(BundleFile{Path: path}, KindProvenance, artifactDigest, NewVerifier(pinnedSigner, fixtureTrust(t)))
		require.NoError(t, err)
		require.Equal(t, 1, found.Position)
		require.Len(t, found.Rejected, 1)
	})

	t.Run("reports why the only candidate was refused", func(t *testing.T) {
		_, err := find(t, pinnedSigner, KindProvenance, "tampered-provenance.sigstore.json")
		require.ErrorContains(t, err, "bundle file testdata/tampered-provenance.sigstore.json: the attestation failed verification")
	})

	t.Run("names a file that cannot be read", func(t *testing.T) {
		_, err := find(t, pinnedSigner, KindProvenance, "missing.sigstore.json")
		require.EqualError(t, err, "bundle file testdata/missing.sigstore.json: open testdata/missing.sigstore.json: no such file or directory")
	})
}

func TestSigstoreTrustRoots(t *testing.T) {
	t.Run("refuses a certificate from an authority it does not know, without fetching anything", func(t *testing.T) {
		_, _, err := SigstoreTrustRoots().forCertificate("Example, Inc.")
		require.EqualError(t, err, `the signing certificate was issued by "Example, Inc.", which is neither Sigstore's public-good instance nor GitHub's`)
	})
}

func TestVerifyWithoutAnArtifact(t *testing.T) {
	verifier := func(t *testing.T) *Verifier { return NewVerifier(pinnedSigner, fixtureTrust(t)) }

	t.Run("checks the signature alone when there is no artifact to check the subject against", func(t *testing.T) {
		evidence, err := verifier(t).Verify(readOne(t, filepath.Join("testdata", "wrong-subject-provenance.sigstore.json")), "")
		require.NoError(t, err)
		require.True(t, evidence.Verified)
	})

	t.Run("still refuses a bad signature", func(t *testing.T) {
		_, err := verifier(t).Verify(readOne(t, filepath.Join("testdata", "tampered-provenance.sigstore.json")), "")
		require.ErrorContains(t, err, "the attestation failed verification")
	})
}
