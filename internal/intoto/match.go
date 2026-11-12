package intoto

import (
	"fmt"
	"slices"
	"strings"
)

// PredicateKind is a family of predicate types recorded the same way in Kosli.
type PredicateKind string

const (
	// KindProvenance matches SLSA provenance v1 and v0.2.
	KindProvenance PredicateKind = "provenance"
	// KindSBOM matches SPDX and CycloneDX documents.
	KindSBOM PredicateKind = "sbom"
)

const (
	slsaProvenanceV1  = "https://slsa.dev/provenance/v1"
	slsaProvenanceV02 = "https://slsa.dev/provenance/v0.2"
	spdxDocument      = "https://spdx.dev/Document"
	cycloneDXBOM      = "https://cyclonedx.org/bom"
)

// PredicateKinds lists every kind, in the order a message offers them.
var PredicateKinds = []PredicateKind{KindProvenance, KindSBOM}

// ParsePredicateKind returns the kind named by s.
func ParsePredicateKind(s string) (PredicateKind, error) {
	for _, kind := range PredicateKinds {
		if string(kind) == s {
			return kind, nil
		}
	}
	return "", fmt.Errorf("predicate type %q is not supported. Must be one of: %s", s, kindList())
}

func kindList() string {
	names := make([]string, len(PredicateKinds))
	for i, kind := range PredicateKinds {
		names[i] = string(kind)
	}
	return strings.Join(names, ", ")
}

// Matches reports whether predicateType is one of this kind's predicate types.
// SPDX and CycloneDX producers may append a version to the type URI.
func (k PredicateKind) Matches(predicateType string) bool {
	switch k {
	case KindProvenance:
		return predicateType == slsaProvenanceV1 || predicateType == slsaProvenanceV02
	case KindSBOM:
		return hasVersionedPrefix(predicateType, spdxDocument) || hasVersionedPrefix(predicateType, cycloneDXBOM)
	}
	return false
}

func hasVersionedPrefix(predicateType, base string) bool {
	return predicateType == base || strings.HasPrefix(predicateType, base+"/")
}

func (k PredicateKind) description() string {
	switch k {
	case KindProvenance:
		return "SLSA provenance v1 or v0.2"
	case KindSBOM:
		return "SPDX or CycloneDX SBOM"
	}
	return string(k)
}

// Matching returns the attestations of the given kind that list the artifact
// with SHA-256 digest as a subject. It fails, saying what it found instead,
// when there are none.
func Matching(attestations []*Attestation, kind PredicateKind, digest string) ([]*Attestation, error) {
	digest = normaliseDigest(digest)

	var ofKind []*Attestation
	var otherTypes []string
	for _, a := range attestations {
		if kind.Matches(a.Statement.GetPredicateType()) {
			ofKind = append(ofKind, a)
		} else if !slices.Contains(otherTypes, a.Statement.GetPredicateType()) {
			otherTypes = append(otherTypes, a.Statement.GetPredicateType())
		}
	}
	if len(ofKind) == 0 {
		return nil, fmt.Errorf("found no %s attestation (%s); the predicate types found were: %s",
			kind, kind.description(), strings.Join(otherTypes, ", "))
	}

	var matching []*Attestation
	var subjects []string
	for _, a := range ofKind {
		if a.HasSubject(digest) {
			matching = append(matching, a)
			continue
		}
		for _, subject := range a.Statement.GetSubject() {
			subjects = append(subjects, describeSubject(subject.GetName(), subject.GetDigest()["sha256"]))
		}
	}
	if len(matching) == 0 {
		return nil, fmt.Errorf("no %s attestation lists the artifact (sha256:%s) as its subject; the subjects found were: %s",
			kind, digest, strings.Join(subjects, ", "))
	}
	return matching, nil
}

// HasSubject reports whether the statement lists the artifact with SHA-256
// digest among its subjects.
func (a *Attestation) HasSubject(digest string) bool {
	digest = normaliseDigest(digest)
	for _, subject := range a.Statement.GetSubject() {
		if strings.EqualFold(subject.GetDigest()["sha256"], digest) {
			return true
		}
	}
	return false
}

// RequireSubject is HasSubject as an error that names the artifact, for a
// statement that need not be signed to say which artifact it is about.
func (a *Attestation) RequireSubject(digest string) error {
	if !a.HasSubject(digest) {
		return fmt.Errorf("%s does not list the artifact (sha256:%s) as its subject", a.Describe(), normaliseDigest(digest))
	}
	return nil
}

func describeSubject(name, sha256 string) string {
	if sha256 == "" {
		return fmt.Sprintf("%s (no sha256 digest)", name)
	}
	return fmt.Sprintf("%s (sha256:%s)", name, sha256)
}

// normaliseDigest accepts a digest with or without its algorithm prefix.
func normaliseDigest(digest string) string {
	return strings.ToLower(strings.TrimPrefix(digest, "sha256:"))
}
