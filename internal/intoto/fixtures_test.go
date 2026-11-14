package intoto

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/secure-systems-lab/go-securesystemslib/dsse"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protocommon "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	protorekor "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/testing/ca"
	"github.com/sigstore/sigstore-go/pkg/tlog"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"
	sigdsse "github.com/sigstore/sigstore/pkg/signature/dsse"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// regenerateFixturesVar names the variable that rewrites the signed fixtures
// in testdata instead of skipping:
//
//	REGENERATE_INTOTO_FIXTURES=1 go test ./internal/intoto -run TestRegenerateFixtures
//
// Every run makes new keys, so the bundles and trusted_root.json must be
// committed together. testdata/README.md describes what each fixture is for.
const regenerateFixturesVar = "REGENERATE_INTOTO_FIXTURES"

const (
	fixtureIssuer        = "https://token.actions.githubusercontent.com"
	fixtureIdentity      = "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main"
	fixtureOtherIdentity = "https://github.com/mallory/widget/.github/workflows/build.yml@refs/heads/main"
	fixtureRunURI        = "https://github.com/acme/widget/actions/runs/1234567890/attempts/1"
	fixtureRerunURI      = "https://github.com/acme/widget/actions/runs/1234567890/attempts/2"
	// A workflow file on a branch or a pull request signs as the repository
	// too, in a run of its own.
	fixtureBranchIdentity      = "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/feature-x"
	fixtureBranchRunURI        = "https://github.com/acme/widget/actions/runs/2222222222/attempts/1"
	fixturePullRequestIdentity = "https://github.com/acme/widget/.github/workflows/build.yml@refs/pull/12/merge"
	fixturePullRequestRunURI   = "https://github.com/acme/widget/actions/runs/3333333333/attempts/1"
)

var (
	oidIssuerV1            = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 57264, 1, 1}
	oidIssuerV2            = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 57264, 1, 8}
	oidSourceRepositoryURI = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 57264, 1, 12}
	oidRunInvocationURI    = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 57264, 1, 21}
)

// fixtureSigstore signs like GitHub Actions on Sigstore's public-good
// instance: a Fulcio certificate naming the workflow, a Rekor entry with an
// inclusion proof, and an RFC 3161 timestamp. The Fulcio authority is built
// here rather than taken from ca.VirtualSigstore, whose certificates can only
// name an email address.
type fixtureSigstore struct {
	t         *testing.T
	virtual   *ca.VirtualSigstore
	fulcio    *root.FulcioCertificateAuthority
	fulcioKey crypto.Signer
	signedAt  time.Time
}

func newFixtureSigstore(t *testing.T) *fixtureSigstore {
	virtual, err := ca.NewVirtualSigstore()
	require.NoError(t, err)
	rootCert, rootKey, err := ca.GenerateRootCa()
	require.NoError(t, err)
	intermediate, intermediateKey, err := ca.GenerateFulcioIntermediate(rootCert, rootKey)
	require.NoError(t, err)
	return &fixtureSigstore{
		t:       t,
		virtual: virtual,
		fulcio: &root.FulcioCertificateAuthority{
			Root:                rootCert,
			Intermediates:       []*x509.Certificate{intermediate},
			ValidityPeriodStart: time.Now().Add(-time.Hour),
			URI:                 "https://fulcio.example.test",
		},
		fulcioKey: intermediateKey,
		signedAt:  time.Now(),
	}
}

func (s *fixtureSigstore) trustedRoot() []byte {
	trustedRoot, err := root.NewTrustedRoot(root.TrustedRootMediaType01,
		[]root.CertificateAuthority{s.fulcio},
		rawLogIDs(s.t, s.virtual.CTLogs()),
		s.virtual.TimestampingAuthorities(),
		rawLogIDs(s.t, s.virtual.RekorLogs()),
	)
	require.NoError(s.t, err)
	content, err := trustedRoot.MarshalJSON()
	require.NoError(s.t, err)
	return indent(s.t, content)
}

// rawLogIDs gives each log the raw key ID a trusted root serialises.
// ca.VirtualSigstore holds the hex string's bytes instead, which work only
// in memory: read back from JSON, no log entry would find its log.
func rawLogIDs(t *testing.T, logs map[string]*root.TransparencyLog) map[string]*root.TransparencyLog {
	for logID, log := range logs {
		id, err := hex.DecodeString(logID)
		require.NoError(t, err)
		log.ID = id
	}
	return logs
}

func (s *fixtureSigstore) leafCertificate(identity, runURI string) (*x509.Certificate, *ecdsa.PrivateKey) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(s.t, err)
	san, err := url.Parse(identity)
	require.NoError(s.t, err)
	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	require.NoError(s.t, err)
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{},
		NotBefore:    s.signedAt.Add(-time.Minute),
		NotAfter:     s.signedAt.Add(10 * time.Minute),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		URIs:         []*url.URL{san},
		ExtraExtensions: []pkix.Extension{
			{Id: oidIssuerV1, Value: []byte(fixtureIssuer)},
			{Id: oidIssuerV2, Value: utf8String(s.t, fixtureIssuer)},
			{Id: oidSourceRepositoryURI, Value: utf8String(s.t, "https://github.com/acme/widget")},
			{Id: oidRunInvocationURI, Value: utf8String(s.t, runURI)},
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, s.fulcio.Intermediates[0], &key.PublicKey, s.fulcioKey)
	require.NoError(s.t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(s.t, err)
	return cert, key
}

// bundle signs statement as identity, in the run the fixtures share, and
// returns the Sigstore bundle, logged at loggedAfter past the signing time.
func (s *fixtureSigstore) bundle(statement []byte, identity string, loggedAfter time.Duration) []byte {
	return s.bundleFromRun(statement, identity, fixtureRunURI, loggedAfter)
}

// bundleFromRun is bundle with a certificate issued to the run at runURI.
func (s *fixtureSigstore) bundleFromRun(statement []byte, identity, runURI string, loggedAfter time.Duration) []byte {
	leaf, key := s.leafCertificate(identity, runURI)
	envelope := signEnvelope(s.t, key, statement)
	sig, err := base64.StdEncoding.DecodeString(envelope.Signatures[0].Sig)
	require.NoError(s.t, err)

	entry, err := s.virtual.GenerateTlogEntry(leaf, envelope, sig, s.signedAt.Add(loggedAfter).Unix(), true)
	require.NoError(s.t, err)
	logEntry := entry.TransparencyLogEntry()
	var body struct {
		Kind       string `json:"kind"`
		APIVersion string `json:"apiVersion"`
	}
	require.NoError(s.t, json.Unmarshal(logEntry.CanonicalizedBody, &body))
	logEntry.KindVersion = &protorekor.KindVersion{Kind: body.Kind, Version: body.APIVersion}
	// The inclusion proof alone does not vouch for the integrated time; the
	// log's signed entry timestamp does, as in a bundle Rekor returns.
	logID, err := s.virtual.RekorLogID()
	require.NoError(s.t, err)
	set, err := s.virtual.RekorSignPayload(tlog.RekorPayload{
		Body:           base64.StdEncoding.EncodeToString(logEntry.CanonicalizedBody),
		IntegratedTime: logEntry.IntegratedTime,
		LogIndex:       logEntry.LogIndex,
		LogID:          logID,
	})
	require.NoError(s.t, err)
	logEntry.InclusionPromise = &protorekor.InclusionPromise{SignedEntryTimestamp: set}

	timestamp, err := s.virtual.TimestampResponse(sig)
	require.NoError(s.t, err)

	b := &protobundle.Bundle{
		MediaType: "application/vnd.dev.sigstore.bundle.v0.3+json",
		VerificationMaterial: &protobundle.VerificationMaterial{
			Content:     &protobundle.VerificationMaterial_Certificate{Certificate: &protocommon.X509Certificate{RawBytes: leaf.Raw}},
			TlogEntries: []*protorekor.TransparencyLogEntry{logEntry},
			TimestampVerificationData: &protobundle.TimestampVerificationData{
				Rfc3161Timestamps: []*protocommon.RFC3161SignedTimestamp{{SignedTimestamp: timestamp}},
			},
		},
		Content: &protobundle.Bundle_DsseEnvelope{DsseEnvelope: &protodsse.Envelope{
			Payload:     statement,
			PayloadType: inTotoPayloadType,
			Signatures:  []*protodsse.Signature{{Sig: sig}},
		}},
	}
	content, err := protojson.Marshal(b)
	require.NoError(s.t, err)
	return content
}

func signEnvelope(t *testing.T, key *ecdsa.PrivateKey, statement []byte) *dsse.Envelope {
	signer, err := signature.LoadECDSASignerVerifier(key, crypto.SHA256)
	require.NoError(t, err)
	envelopeSigner, err := dsse.NewEnvelopeSigner(&sigdsse.SignerAdapter{SignatureSigner: signer, Pub: &key.PublicKey})
	require.NoError(t, err)
	envelope, err := envelopeSigner.SignPayload(context.Background(), inTotoPayloadType, statement)
	require.NoError(t, err)
	return envelope
}

// tamper changes the signed statement inside a bundle without re-signing it.
func tamper(t *testing.T, signedBundle []byte, from, to string) []byte {
	var b protobundle.Bundle
	require.NoError(t, protojson.Unmarshal(signedBundle, &b))
	envelope := b.GetDsseEnvelope()
	require.Contains(t, string(envelope.Payload), from)
	envelope.Payload = []byte(strings.Replace(string(envelope.Payload), from, to, 1))
	content, err := protojson.Marshal(&b)
	require.NoError(t, err)
	return content
}

func utf8String(t *testing.T, s string) []byte {
	value, err := asn1.MarshalWithParams(s, "utf8")
	require.NoError(t, err)
	return value
}

func indent(t *testing.T, content []byte) []byte {
	var value any
	require.NoError(t, json.Unmarshal(content, &value))
	out, err := json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	return append(out, '\n')
}

func writeFixture(t *testing.T, name string, content []byte) {
	require.NoError(t, os.WriteFile(filepath.Join("testdata", name), content, 0o644))
}

func TestRegenerateFixtures(t *testing.T) {
	if os.Getenv(regenerateFixturesVar) == "" {
		t.Skipf("set %s=1 to rewrite the signed fixtures in testdata", regenerateFixturesVar)
	}
	s := newFixtureSigstore(t)
	statement := func(name string) []byte { return compact(t, statementFixture(t, name)) }
	jsonl := func(lines ...[]byte) []byte {
		out := []byte{}
		for _, line := range lines {
			out = append(out, compact(t, line)...)
			out = append(out, '\n')
		}
		return out
	}

	writeFixture(t, "trusted_root.json", s.trustedRoot())

	for _, name := range []string{"github-provenance", "buildkit-provenance", "slsa-v0.2-provenance", "spdx-sbom", "cyclonedx-sbom", "wrong-subject-provenance"} {
		writeFixture(t, name+".sigstore.json", indent(t, s.bundle(statement(name+".json"), fixtureIdentity, time.Minute)))
	}
	writeFixture(t, "wrong-signer-provenance.sigstore.json",
		indent(t, s.bundle(statement("github-provenance.json"), fixtureOtherIdentity, time.Minute)))
	writeFixture(t, "rerun-provenance.sigstore.json",
		indent(t, s.bundleFromRun(statement("rerun-provenance.json"), fixtureIdentity, fixtureRerunURI, 2*time.Minute)))
	writeFixture(t, "branch-provenance.sigstore.json",
		indent(t, s.bundleFromRun(statement("github-provenance.json"), fixtureBranchIdentity, fixtureBranchRunURI, 3*time.Minute)))
	writeFixture(t, "pull-request-provenance.sigstore.json",
		indent(t, s.bundleFromRun(statement("github-provenance.json"), fixturePullRequestIdentity, fixturePullRequestRunURI, 4*time.Minute)))
	writeFixture(t, "tampered-provenance.sigstore.json",
		indent(t, tamper(t, s.bundle(statement("github-provenance.json"), fixtureIdentity, time.Minute),
			"4f9d2a7c1e6b8a3d5f0e2c9b7a1d4e6f8c3b5a70", "0000000000000000000000000000000000000000")))

	writeFixture(t, "attestations.intoto.jsonl", jsonl(
		s.bundle(statement("vuln-scan.json"), fixtureIdentity, time.Minute),
		s.bundle(statement("github-provenance.json"), fixtureIdentity, time.Minute),
		s.bundle(statement("cyclonedx-sbom.json"), fixtureIdentity, time.Minute),
	))
	writeFixture(t, "newest-first.intoto.jsonl", jsonl(
		s.bundle(statement("github-provenance.json"), fixtureIdentity, 3*time.Minute),
		s.bundle(statement("github-provenance.json"), fixtureIdentity, 2*time.Minute),
	))
	writeFixture(t, "vuln-only.intoto.jsonl", jsonl(
		s.bundle(statement("vuln-scan.json"), fixtureIdentity, time.Minute),
		statement("vuln-scan.json"),
	))
	writeFixture(t, "wrong-signer-and-unsigned.intoto.jsonl", jsonl(
		s.bundle(statement("github-provenance.json"), fixtureOtherIdentity, time.Minute),
		statement("github-provenance.json"),
	))

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	publicKey, err := cryptoutils.MarshalPublicKeyToPEM(&key.PublicKey)
	require.NoError(t, err)
	writeFixture(t, "key.pub", publicKey)
	keySigned, err := json.Marshal(signEnvelope(t, key, statement("github-provenance.json")))
	require.NoError(t, err)
	writeFixture(t, "key-signed-provenance.dsse.json", indent(t, keySigned))

	unsigned, err := json.Marshal(dsse.Envelope{
		PayloadType: inTotoPayloadType,
		Payload:     base64.StdEncoding.EncodeToString(statement("spdx-sbom.json")),
		Signatures:  []dsse.Signature{},
	})
	require.NoError(t, err)
	writeFixture(t, "unsigned-spdx-sbom.dsse.json", indent(t, unsigned))
}
