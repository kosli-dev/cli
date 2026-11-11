package intoto

import "fmt"

// SBOMDocument returns the SPDX or CycloneDX document an SBOM attestation
// carries, as the bytes the signed payload held, for internal/sbom to read.
func SBOMDocument(a *Attestation) ([]byte, error) {
	if !KindSBOM.Matches(a.Statement.GetPredicateType()) {
		return nil, fmt.Errorf("%s is %s, not an SPDX or CycloneDX SBOM", a.Describe(), a.Statement.GetPredicateType())
	}
	return a.Predicate, nil
}
