package intoto

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const artifactDigest = "e88951291bca0807a41f3c0144009bfff5c6c77dc1310d52e8c6d62feee960a8"

func statementFixture(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", "statements", name))
	require.NoError(t, err)
	return content
}

func compact(t *testing.T, content []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	require.NoError(t, json.Compact(&out, content))
	return out.Bytes()
}

// envelope wraps a statement in a DSSE envelope. The signature is not a real
// one: reading never checks it.
func envelope(t *testing.T, payloadType string, statement []byte, signatures ...string) []byte {
	t.Helper()
	sigs := []map[string]string{}
	for _, sig := range signatures {
		sigs = append(sigs, map[string]string{"keyid": "", "sig": base64.StdEncoding.EncodeToString([]byte(sig))})
	}
	content, err := json.Marshal(map[string]any{
		"payloadType": payloadType,
		"payload":     base64.StdEncoding.EncodeToString(statement),
		"signatures":  sigs,
	})
	require.NoError(t, err)
	return content
}

func TestRead(t *testing.T) {
	t.Run("reads a bare statement as one unsigned attestation", func(t *testing.T) {
		content := statementFixture(t, "github-provenance.json")

		attestations, err := Read(content)
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Equal(t, "https://slsa.dev/provenance/v1", attestations[0].Statement.GetPredicateType())
		require.False(t, attestations[0].Signed())
		require.Equal(t, 0, attestations[0].Position)
		require.Equal(t, content, attestations[0].Raw, "Raw should be the whole file, so its checksum is the file's")
	})

	t.Run("reads a v0.1 statement", func(t *testing.T) {
		attestations, err := Read(statementFixture(t, "slsa-v0.2-provenance.json"))
		require.NoError(t, err)
		require.Equal(t, "https://slsa.dev/provenance/v0.2", attestations[0].Statement.GetPredicateType())
	})

	t.Run("keeps the predicate exactly as the payload carried it", func(t *testing.T) {
		attestations, err := Read(statementFixture(t, "cyclonedx-sbom.json"))
		require.NoError(t, err)
		require.Contains(t, string(attestations[0].Predicate), `"version": 1,`)
	})

	t.Run("reads an unsigned DSSE envelope", func(t *testing.T) {
		attestations, err := Read(envelope(t, inTotoPayloadType, statementFixture(t, "spdx-sbom.json")))
		require.NoError(t, err)
		require.Len(t, attestations, 1)
		require.Equal(t, "https://spdx.dev/Document/v2.3", attestations[0].Statement.GetPredicateType())
		require.False(t, attestations[0].Signed())
	})

	t.Run("treats a DSSE envelope with a signature as signed", func(t *testing.T) {
		attestations, err := Read(envelope(t, inTotoPayloadType, statementFixture(t, "spdx-sbom.json"), "signature"))
		require.NoError(t, err)
		require.True(t, attestations[0].Signed())
	})

	t.Run("refuses a DSSE envelope that does not carry an in-toto statement", func(t *testing.T) {
		_, err := Read(envelope(t, "application/vnd.example+json", statementFixture(t, "spdx-sbom.json")))
		require.EqualError(t, err, `the attestation: the DSSE envelope carries "application/vnd.example+json", not an in-toto statement (application/vnd.in-toto+json)`)
	})

	t.Run("reads one attestation per line, numbering them", func(t *testing.T) {
		lines := [][]byte{
			compact(t, statementFixture(t, "vuln-scan.json")),
			compact(t, envelope(t, inTotoPayloadType, statementFixture(t, "github-provenance.json"))),
			compact(t, statementFixture(t, "cyclonedx-sbom.json")),
		}
		content := append(bytes.Join(lines, []byte("\n")), '\n')

		attestations, err := Read(content)
		require.NoError(t, err)
		require.Len(t, attestations, 3)
		for i, a := range attestations {
			require.Equal(t, i+1, a.Position)
			require.Equal(t, lines[i], a.Raw, "Raw should be the line the attestation was read from")
		}
		require.Equal(t, "https://slsa.dev/provenance/v1", attestations[1].Statement.GetPredicateType())
	})

	t.Run("names the line a bad attestation is on", func(t *testing.T) {
		content := append(compact(t, statementFixture(t, "vuln-scan.json")), []byte("\n{\"hello\": \"world\"}\n")...)
		_, err := Read(content)
		require.EqualError(t, err, "attestation 2: not a Sigstore bundle, a DSSE envelope or an in-toto statement")
	})

	t.Run("refuses a file that is not JSON", func(t *testing.T) {
		_, err := Read([]byte("SPDXVersion: SPDX-2.3\n"))
		require.ErrorContains(t, err, "not valid JSON")
	})

	t.Run("refuses an empty file", func(t *testing.T) {
		_, err := Read([]byte("\n"))
		require.EqualError(t, err, "holds no attestation")
	})

	t.Run("refuses a statement with no subject", func(t *testing.T) {
		_, err := Read([]byte(`{"_type": "https://in-toto.io/Statement/v1", "predicateType": "https://slsa.dev/provenance/v1", "predicate": {}}`))
		require.EqualError(t, err, "the attestation: not a valid in-toto statement: at least one subject required")
	})
}

func TestReadEnvelope(t *testing.T) {
	t.Run("reads a Sigstore bundle", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join("testdata", "spdx-sbom.sigstore.json"))
		require.NoError(t, err)

		attestation, wrapped, err := ReadEnvelope(content)
		require.NoError(t, err)
		require.True(t, wrapped)
		require.True(t, attestation.Signed())
		require.Equal(t, content, attestation.Raw)
	})

	t.Run("reads a DSSE envelope", func(t *testing.T) {
		attestation, wrapped, err := ReadEnvelope(envelope(t, inTotoPayloadType, statementFixture(t, "spdx-sbom.json")))
		require.NoError(t, err)
		require.True(t, wrapped)
		require.False(t, attestation.Signed())
	})

	t.Run("leaves anything else, such as the SBOM itself, unwrapped", func(t *testing.T) {
		for _, content := range []string{
			`{"bomFormat": "CycloneDX", "specVersion": "1.6"}`,
			`<?xml version="1.0"?><bom xmlns="http://cyclonedx.org/schema/bom/1.6"/>`,
			"SPDXVersion: SPDX-2.3\n",
			string(statementFixture(t, "spdx-sbom.json")),
		} {
			_, wrapped, err := ReadEnvelope([]byte(content))
			require.NoError(t, err)
			require.False(t, wrapped, content)
		}
	})

	t.Run("refuses several attestations, which no single bundle or envelope can be", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join("testdata", "attestations.intoto.jsonl"))
		require.NoError(t, err)

		_, _, err = ReadEnvelope(content)
		require.ErrorIs(t, err, ErrSeveralDocuments)
	})

	t.Run("reports a malformed envelope rather than treating it as an SBOM", func(t *testing.T) {
		_, wrapped, err := ReadEnvelope([]byte(`{"payloadType": "application/vnd.in-toto+json", "payload": "not base64!", "signatures": []}`))
		require.True(t, wrapped)
		require.ErrorContains(t, err, "the DSSE envelope's payload is not valid base64")
	})
}
