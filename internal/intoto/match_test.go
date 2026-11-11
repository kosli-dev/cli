package intoto

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func readFixtures(t *testing.T, names ...string) []*Attestation {
	t.Helper()
	lines := make([][]byte, len(names))
	for i, name := range names {
		lines[i] = compact(t, statementFixture(t, name))
	}
	attestations, err := Read(bytes.Join(lines, []byte("\n")))
	require.NoError(t, err)
	return attestations
}

func TestPredicateKind(t *testing.T) {
	t.Run("provenance matches SLSA v1 and v0.2 only", func(t *testing.T) {
		require.True(t, KindProvenance.Matches("https://slsa.dev/provenance/v1"))
		require.True(t, KindProvenance.Matches("https://slsa.dev/provenance/v0.2"))
		require.False(t, KindProvenance.Matches("https://slsa.dev/provenance/v0.1"))
		require.False(t, KindProvenance.Matches("https://spdx.dev/Document"))
	})

	t.Run("sbom matches SPDX and CycloneDX, versioned or not", func(t *testing.T) {
		require.True(t, KindSBOM.Matches("https://spdx.dev/Document"))
		require.True(t, KindSBOM.Matches("https://spdx.dev/Document/v2.3"))
		require.True(t, KindSBOM.Matches("https://cyclonedx.org/bom"))
		require.True(t, KindSBOM.Matches("https://cyclonedx.org/bom/v1.6"))
		require.False(t, KindSBOM.Matches("https://spdx.dev/Documentation"))
		require.False(t, KindSBOM.Matches("https://slsa.dev/provenance/v1"))
	})

	t.Run("parses the names --predicate-type accepts", func(t *testing.T) {
		kind, err := ParsePredicateKind("sbom")
		require.NoError(t, err)
		require.Equal(t, KindSBOM, kind)

		_, err = ParsePredicateKind("vuln")
		require.EqualError(t, err, `predicate type "vuln" is not supported. Must be one of: provenance, sbom`)
	})
}

func TestMatching(t *testing.T) {
	t.Run("keeps the attestations of the kind asked for", func(t *testing.T) {
		attestations := readFixtures(t, "vuln-scan.json", "github-provenance.json", "cyclonedx-sbom.json", "spdx-sbom.json")

		matching, err := Matching(attestations, KindSBOM, artifactDigest)
		require.NoError(t, err)
		require.Len(t, matching, 2)
		require.Equal(t, 3, matching[0].Position)
		require.Equal(t, 4, matching[1].Position)
	})

	t.Run("accepts a digest with its algorithm prefix, in any case", func(t *testing.T) {
		attestations := readFixtures(t, "github-provenance.json")

		matching, err := Matching(attestations, KindProvenance, "sha256:"+strings.ToUpper(artifactDigest))
		require.NoError(t, err)
		require.Len(t, matching, 1)
	})

	t.Run("names the predicate types found when none is of the kind asked for", func(t *testing.T) {
		attestations := readFixtures(t, "vuln-scan.json", "vuln-scan.json", "spdx-sbom.json")

		_, err := Matching(attestations, KindProvenance, artifactDigest)
		require.EqualError(t, err, "found no provenance attestation (SLSA provenance v1 or v0.2); the predicate types found were: https://cosign.sigstore.dev/attestation/vuln/v1, https://spdx.dev/Document/v2.3")
	})

	t.Run("names the subjects found when none is the artifact", func(t *testing.T) {
		attestations := readFixtures(t, "wrong-subject-provenance.json")

		_, err := Matching(attestations, KindProvenance, artifactDigest)
		require.EqualError(t, err, "no provenance attestation lists the artifact (sha256:"+artifactDigest+") as its subject; the subjects found were: other.bin (sha256:11de96226896b0fcf4ea66a11a63a5fd9f4e211fe94b62b7eebf4485c817b2d8)")
	})

	t.Run("skips an attestation about another artifact when one is about this one", func(t *testing.T) {
		attestations := readFixtures(t, "wrong-subject-provenance.json", "github-provenance.json")

		matching, err := Matching(attestations, KindProvenance, artifactDigest)
		require.NoError(t, err)
		require.Len(t, matching, 1)
		require.Equal(t, 2, matching[0].Position)
	})
}
