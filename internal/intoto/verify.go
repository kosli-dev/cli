package intoto

import (
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protocommon "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"
)

// Policy says whose signature a signed attestation must carry. There is no
// way to accept any valid certificate: a signer is always pinned.
type Policy struct {
	CertificateIdentity       string
	CertificateIdentityRegexp string
	CertificateOIDCIssuer     string
	// KeyPath names a PEM public key, for attestations signed with a key
	// rather than a Fulcio certificate.
	KeyPath string
	// AllowUnsigned accepts attestations with no signature, recorded as
	// unverified.
	AllowUnsigned bool
}

// Validate checks the policy's parts make sense together. It needs nothing
// from an attestation, so a caller can run it before reading one.
func (p Policy) Validate() error {
	hasIdentity := p.CertificateIdentity != "" || p.CertificateIdentityRegexp != ""
	switch {
	case p.CertificateIdentity != "" && p.CertificateIdentityRegexp != "":
		return fmt.Errorf("only one of --certificate-identity, --certificate-identity-regexp is allowed")
	case p.KeyPath != "" && (hasIdentity || p.CertificateOIDCIssuer != ""):
		return fmt.Errorf("--key cannot be combined with --certificate-identity, --certificate-identity-regexp or --certificate-oidc-issuer: a signature is made either with a key or with a certificate")
	case hasIdentity && p.CertificateOIDCIssuer == "":
		return fmt.Errorf("--certificate-oidc-issuer is required with --certificate-identity or --certificate-identity-regexp")
	case p.CertificateOIDCIssuer != "" && !hasIdentity:
		return fmt.Errorf("--certificate-identity or --certificate-identity-regexp is required with --certificate-oidc-issuer")
	}
	if p.CertificateIdentityRegexp != "" {
		if _, err := regexp.Compile(p.CertificateIdentityRegexp); err != nil {
			return fmt.Errorf("--certificate-identity-regexp is not a valid regular expression: %w", err)
		}
	}
	return nil
}

// PinsSigner reports whether the policy can verify a signature at all.
func (p Policy) PinsSigner() bool {
	return p.KeyPath != "" || p.CertificateOIDCIssuer != ""
}

// Evidence is what checking an attestation established.
type Evidence struct {
	// Verified is false only for an unsigned attestation the policy allowed.
	Verified bool
	// SignerIdentity and SignerIssuer are the certificate's subject
	// alternative name and OIDC issuer; empty for a key signature.
	SignerIdentity string
	SignerIssuer   string
	// RunInvocationURI is the CI run that signed, when the certificate says.
	RunInvocationURI string
	SignedWithKey    bool
	// LogIndex, IntegratedTime and LogEntryURL describe the verified
	// transparency log entry, when there is one.
	LogIndex       *int64
	IntegratedTime time.Time
	LogEntryURL    string
	// SignedTime is the time of a verified RFC 3161 timestamp.
	SignedTime time.Time
	// RawSHA256 is the SHA-256 of the bytes the attestation was read from.
	RawSHA256 string
}

// Time is the trusted time of the signature, or zero when nothing vouches for
// one.
func (e *Evidence) Time() time.Time {
	if !e.IntegratedTime.IsZero() {
		return e.IntegratedTime
	}
	return e.SignedTime
}

// Verifier checks attestations against a policy.
type Verifier struct {
	policy Policy
	trust  *TrustRoots
}

// NewVerifier returns a Verifier that checks signatures against trust.
func NewVerifier(policy Policy, trust *TrustRoots) *Verifier {
	return &Verifier{policy: policy, trust: trust}
}

// Verify checks that a names the artifact with SHA-256 digest as its subject
// and, when a is signed, that the signature is valid and made by the signer
// the policy pins. An empty digest checks the signature alone, for an
// attestation recorded on a trail rather than an artifact.
func (v *Verifier) Verify(a *Attestation, digest string) (*Evidence, error) {
	if digest != "" && !a.HasSubject(digest) {
		return nil, fmt.Errorf("%s does not list the artifact (sha256:%s) as its subject", a.Describe(), normaliseDigest(digest))
	}
	rawSum := sha256.Sum256(a.Raw)
	evidence := &Evidence{RawSHA256: hex.EncodeToString(rawSum[:])}

	if !a.Signed() {
		if !v.policy.AllowUnsigned {
			return nil, fmt.Errorf("%s is not signed. Pass --allow-unsigned to record it as unverified", a.Describe())
		}
		return evidence, nil
	}
	if !v.policy.PinsSigner() {
		return nil, fmt.Errorf("%s is signed, so its signer must be pinned to verify it. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key", a.Describe())
	}

	result, entity, err := v.verifySignature(a, digest)
	if err != nil {
		return nil, fmt.Errorf("%s failed verification: %w", a.Describe(), err)
	}

	evidence.Verified = true
	evidence.SignedWithKey = v.policy.KeyPath != ""
	if summary := result.Signature.Certificate; summary != nil {
		evidence.SignerIdentity = summary.SubjectAlternativeName
		evidence.SignerIssuer = summary.Issuer
		evidence.RunInvocationURI = summary.RunInvocationURI
	}
	for _, ts := range result.VerifiedTimestamps {
		switch ts.Type {
		case "Tlog":
			evidence.IntegratedTime = ts.Timestamp.UTC()
		case "TimestampAuthority":
			evidence.SignedTime = ts.Timestamp.UTC()
		}
	}
	v.recordLogEntry(evidence, entity)
	return evidence, nil
}

func (v *Verifier) verifySignature(a *Attestation, digest string) (*verify.VerificationResult, *bundle.Bundle, error) {
	artifactPolicy := verify.WithoutArtifactUnsafe()
	if digest != "" {
		digestBytes, err := hex.DecodeString(normaliseDigest(digest))
		if err != nil {
			return nil, nil, fmt.Errorf("the artifact digest is not hexadecimal: %w", err)
		}
		artifactPolicy = verify.WithArtifactDigest("sha256", digestBytes)
	}

	if v.policy.KeyPath != "" {
		return v.verifyWithKey(a, artifactPolicy)
	}

	if a.bundle == nil {
		return nil, nil, fmt.Errorf("a DSSE envelope carries no certificate to check --certificate-identity against. Supply the Sigstore bundle instead, or verify the envelope with --key")
	}
	verificationContent, err := a.bundle.VerificationContent()
	if err != nil {
		return nil, nil, err
	}
	leaf := verificationContent.Certificate()
	if leaf == nil {
		return nil, nil, fmt.Errorf("it is signed with a key, not a certificate. Verify it with --key")
	}
	issuerOrg := ""
	if len(leaf.Issuer.Organization) > 0 {
		issuerOrg = leaf.Issuer.Organization[0]
	}
	trustedMaterial, opts, err := v.trust.forCertificate(issuerOrg)
	if err != nil {
		return nil, nil, err
	}
	identity, err := verify.NewShortCertificateIdentity(v.policy.CertificateOIDCIssuer, "", v.policy.CertificateIdentity, v.policy.CertificateIdentityRegexp)
	if err != nil {
		return nil, nil, err
	}
	verifier, err := verify.NewVerifier(trustedMaterial, opts...)
	if err != nil {
		return nil, nil, err
	}
	result, err := verifier.Verify(a.bundle, verify.NewPolicy(artifactPolicy, verify.WithCertificateIdentity(identity)))
	return result, a.bundle, err
}

func (v *Verifier) verifyWithKey(a *Attestation, artifactPolicy verify.ArtifactPolicyOption) (*verify.VerificationResult, *bundle.Bundle, error) {
	key, err := loadKey(v.policy.KeyPath)
	if err != nil {
		return nil, nil, err
	}
	keyMaterial := root.NewTrustedPublicKeyMaterial(func(string) (root.TimeConstrainedVerifier, error) {
		return key, nil
	})

	entity := a.bundle
	if entity == nil {
		entity, err = envelopeAsBundle(a)
		if err != nil {
			return nil, nil, err
		}
	}

	var trustedMaterial root.TrustedMaterial = keyMaterial
	var opts []verify.VerifierOption
	tlogEntries, err := entity.TlogEntries()
	if err != nil {
		return nil, nil, err
	}
	timestamps, err := entity.Timestamps()
	if err != nil {
		return nil, nil, err
	}
	switch {
	case len(tlogEntries) > 0:
		opts = []verify.VerifierOption{verify.WithTransparencyLog(1), verify.WithObserverTimestamps(1)}
	case len(timestamps) > 0:
		opts = []verify.VerifierOption{verify.WithSignedTimestamps(1)}
	default:
		opts = []verify.VerifierOption{verify.WithNoObserverTimestamps()}
	}
	if len(tlogEntries) > 0 || len(timestamps) > 0 {
		trustRoot, err := v.trust.forKey()
		if err != nil {
			return nil, nil, err
		}
		trustedMaterial = root.TrustedMaterialCollection{trustRoot, keyMaterial}
	}

	verifier, err := verify.NewVerifier(trustedMaterial, opts...)
	if err != nil {
		return nil, nil, err
	}
	result, err := verifier.Verify(entity, verify.NewPolicy(artifactPolicy, verify.WithKey()))
	return result, entity, err
}

// envelopeAsBundle puts a bare DSSE envelope in a bundle whose verification
// material is a key hint, which is how sigstore-go verifies a key signature.
func envelopeAsBundle(a *Attestation) (*bundle.Bundle, error) {
	payload, err := a.envelope.DecodeB64Payload()
	if err != nil {
		return nil, err
	}
	signatures := make([]*protodsse.Signature, 0, len(a.envelope.Signatures))
	for _, s := range a.envelope.Signatures {
		sig, err := base64.StdEncoding.DecodeString(s.Sig)
		if err != nil {
			return nil, fmt.Errorf("the DSSE envelope's signature is not valid base64: %w", err)
		}
		signatures = append(signatures, &protodsse.Signature{Sig: sig, Keyid: s.KeyID})
	}
	mediaType, err := bundle.MediaTypeString("0.3")
	if err != nil {
		return nil, err
	}
	return bundle.NewBundle(&protobundle.Bundle{
		MediaType: mediaType,
		VerificationMaterial: &protobundle.VerificationMaterial{
			Content: &protobundle.VerificationMaterial_PublicKey{PublicKey: &protocommon.PublicKeyIdentifier{}},
		},
		Content: &protobundle.Bundle_DsseEnvelope{DsseEnvelope: &protodsse.Envelope{
			Payload:     payload,
			PayloadType: a.envelope.PayloadType,
			Signatures:  signatures,
		}},
	})
}

func loadKey(path string) (root.TimeConstrainedVerifier, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read --key %s: %w", path, err)
	}
	publicKey, err := cryptoutils.UnmarshalPEMToPublicKey(content)
	if err != nil {
		return nil, fmt.Errorf("--key %s is not a PEM public key: %w", path, err)
	}
	verifier, err := signature.LoadVerifier(publicKey, crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("--key %s cannot verify signatures: %w", path, err)
	}
	return root.NewExpiringKey(verifier, time.Time{}, time.Time{}), nil
}

// recordLogEntry adds the transparency log entry, already verified, so a
// reader can find it in the log.
func (v *Verifier) recordLogEntry(evidence *Evidence, entity *bundle.Bundle) {
	entries, err := entity.TlogEntries()
	if err != nil || len(entries) == 0 {
		return
	}
	entry := entries[0]
	index := entry.LogIndex()
	evidence.LogIndex = &index
	if entry.IsRekorV2() {
		return
	}
	evidence.LogEntryURL = v.logEntryURL(hex.EncodeToString([]byte(entry.LogKeyID())), index)
}

// logEntryURL links public-good entries to Sigstore's search page, which a
// person can read, and any other log's to its API.
func (v *Verifier) logEntryURL(logID string, index int64) string {
	var logs map[string]*root.TransparencyLog
	if v.trust.fixed != nil {
		logs = v.trust.fixed.RekorLogs()
	} else if trustedRoot, err := v.trust.publicGood.get(); err == nil {
		logs = trustedRoot.RekorLogs()
	}
	log, ok := logs[logID]
	if !ok || log.BaseURL == "" {
		return ""
	}
	base, err := url.Parse(log.BaseURL)
	if err != nil {
		return ""
	}
	if base.Hostname() == "rekor.sigstore.dev" {
		return fmt.Sprintf("https://search.sigstore.dev/?logIndex=%d", index)
	}
	return fmt.Sprintf("%s/api/v1/log/entries?logIndex=%d", strings.TrimSuffix(log.BaseURL, "/"), index)
}
