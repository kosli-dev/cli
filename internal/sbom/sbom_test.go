package sbom

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixture(name string) string {
	return filepath.Join("testdata", name)
}

func TestFormatDetection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		file   string
		format string
	}{
		{"cyclonedx json", "cyclonedx-1.6.json", "cyclonedx-1.6"},
		{"cyclonedx xml", "cyclonedx-1.6.xml", "cyclonedx-1.6"},
		{"spdx json 2.3", "spdx-2.3.json", "spdx-2.3"},
		{"spdx json 2.2", "spdx-2.2.json", "spdx-2.2"},
		{"spdx tag-value 2.3", "spdx-2.3.spdx", "spdx-2.3"},
		{"spdx tag-value 2.2", "spdx-2.2-no-snippet.spdx", "spdx-2.2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ProcessSBOMFile(fixture(tc.file))

			require.NoError(t, err)
			assert.Equal(t, tc.format, got.Format)
		})
	}
}

func TestCycloneDXMustIdentifyItself(t *testing.T) {
	for _, tc := range []struct {
		name    string
		file    string
		wantErr string
	}{
		// The decoder returns an empty document rather than an error for any
		// well-formed JSON, so these two checks are the only thing rejecting it.
		{"no bomFormat", "not-an-sbom.json", `bomFormat is ""`},
		{"no specVersion", "cyclonedx-no-spec-version.json", "no specVersion"},
		// XML carries its identity in the namespace, not in bomFormat.
		{"wrong xml namespace", "cyclonedx-wrong-namespace.xml", "expected a http://cyclonedx.org/schema/bom/ namespace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ProcessSBOMFile(fixture(tc.file))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestSPDXVersionComesFromTheFileNotTheParser(t *testing.T) {
	// The SPDX readers convert every document up to their newest model, so the
	// parsed document reports 2.3 whatever the file said.
	for _, file := range []string{"spdx-2.2.json", "spdx-2.2-no-snippet.spdx"} {
		t.Run(file, func(t *testing.T) {
			got, err := ProcessSBOMFile(fixture(file))

			require.NoError(t, err)
			assert.Equal(t, "spdx-2.2", got.Format)
		})
	}
}

func TestSPDXTagValueSnippetGivesAnActionableError(t *testing.T) {
	// SPDX's own 2.2 tag-value example has a snippet after a package, which the
	// tag-value parser cannot read.
	_, err := ProcessSBOMFile(fixture("spdx-2.2.spdx"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "snippet section")
	assert.Contains(t, err.Error(), "JSON form")
}

func TestPackageCountExcludesFiles(t *testing.T) {
	// Syft emits one component per file in the scanned image; only packages count.
	got, err := ProcessSBOMFile(fixture("cyclonedx-with-files.json"))

	require.NoError(t, err)
	assert.Equal(t, 2, got.Document.PackageCount)
}

func TestToolsReadFromBothCycloneDXLayouts(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
	}{
		{"post-1.5 components layout", "cyclonedx-tools.json"},
		{"deprecated pre-1.5 layout", "cyclonedx-tools-deprecated.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ProcessSBOMFile(fixture(tc.file))

			require.NoError(t, err)
			assert.Equal(t, []string{"Awesome Tool 9.1.2"}, got.Document.Tools)
		})
	}
}

func TestCycloneDXExtraction(t *testing.T) {
	got, err := ProcessSBOMFile(fixture("cyclonedx-1.6.json"))

	require.NoError(t, err)
	require.NotNil(t, got.Document.CreatedAt)
	assert.Equal(t, "2020-04-13T20:20:39+00:00", *got.Document.CreatedAt)
	assert.Equal(t, 2, got.Document.PackageCount)
	require.NotNil(t, got.Document.Subject)
	assert.Equal(t, "Acme Application", got.Document.Subject.Name)
	require.NotNil(t, got.Document.Subject.Version)
	assert.Equal(t, "9.1.1", *got.Document.Subject.Version)
}

func TestSPDXSubjectIsNullWhenSeveralPackagesAreDescribed(t *testing.T) {
	// Both official examples describe two packages; guessing one would be wrong.
	got, err := ProcessSBOMFile(fixture("spdx-2.3.json"))

	require.NoError(t, err)
	assert.Nil(t, got.Document.Subject)
	assert.Greater(t, got.Document.PackageCount, 1)
}

func TestUnreadableInputFailsClearly(t *testing.T) {
	for _, tc := range []struct {
		name    string
		file    string
		wantErr string
	}{
		{"gzipped", "gzipped.json.gz", "gzip compressed"},
		{"truncated", "truncated.json", "could not parse"},
		{"binary", "binary.bin", "unrecognised file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ProcessSBOMFile(fixture(tc.file))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestMissingFileFailsClearly(t *testing.T) {
	_, err := ProcessSBOMFile(fixture("no-such-file.json"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no-such-file.json")
}

func TestMalformedSPDXIsAnErrorNotAPanic(t *testing.T) {
	// A null entry in the package list makes the SPDX reader dereference it. A
	// file the user supplied must not take the process down.
	_, err := ProcessSBOMFile(fixture("spdx-null-package.json"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse the file as an SPDX SBOM")
}

func TestADocumentCarryingBothIdentitiesIsRejected(t *testing.T) {
	// A stray spdxVersion key routes the file to the SPDX reader, which accepts
	// anything with a version and returns an empty document. Recording that as a
	// successful SBOM would attest data the file never carried.
	_, err := ProcessSBOMFile(fixture("cyclonedx-hybrid-spdx-key.json"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an SPDX SBOM")
}

func TestPackageCountIncludesNestedComponents(t *testing.T) {
	// One application containing two libraries and a file: the assembly is
	// itself a package, so three, and the file is still excluded.
	got, err := ProcessSBOMFile(fixture("cyclonedx-nested.json"))

	require.NoError(t, err)
	assert.Equal(t, 3, got.Document.PackageCount)
}

func TestSPDXSubjectFromDescribedByRelationship(t *testing.T) {
	// DESCRIBED_BY is the inverse of DESCRIBES and names the subject just as
	// well. The checksum is uppercase in the file and must be recorded lowercase,
	// which is the only form the server's schema accepts.
	got, err := ProcessSBOMFile(fixture("spdx-described-by.json"))

	require.NoError(t, err)
	assert.Equal(t, []string{"syft-1.50.0"}, got.Document.Tools)
	require.NotNil(t, got.Document.Subject)
	assert.Equal(t, "app", got.Document.Subject.Name)
	require.NotNil(t, got.Document.Subject.Version)
	assert.Equal(t, "1.2.3", *got.Document.Subject.Version)
	require.NotNil(t, got.Document.Subject.Purl)
	assert.Equal(t, "pkg:generic/app@1.2.3", *got.Document.Subject.Purl)
	require.NotNil(t, got.Document.Subject.Sha256)
	assert.Equal(t, "3f786850e387550fdab836ed7e6dc881de23001b3f786850e387550fdab836ed", *got.Document.Subject.Sha256)
}

func TestAByteOrderMarkDoesNotDefeatParsing(t *testing.T) {
	got, err := ProcessSBOMFile(fixture("cyclonedx-bom-prefix.json"))

	require.NoError(t, err)
	assert.Equal(t, "cyclonedx-1.6", got.Format)
}

func TestVersionIsNotReadFromAQuotedHeader(t *testing.T) {
	// A text block quoting another document's header precedes the real version.
	got, err := ProcessSBOMFile(fixture("spdx-quoted-version.spdx"))

	require.NoError(t, err)
	assert.Equal(t, "spdx-2.3", got.Format)
}

func TestFieldsTheSBOMDoesNotCarryAreNull(t *testing.T) {
	// The server's schema types every field inside document as nullable and
	// enforces a date-time format and a hex pattern, so an empty string is
	// rejected where an absent value is accepted.
	got, err := ProcessSBOMFile(fixture("cyclonedx-1.6.json"))

	require.NoError(t, err)
	require.NotNil(t, got.Document.Subject)
	assert.Nil(t, got.Document.Subject.Sha256)
	assert.Nil(t, got.Document.Subject.Purl)

	encoded, err := json.Marshal(got.Document)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"sha256":null`)
	assert.Contains(t, string(encoded), `"purl":null`)
}

func TestPrefixedXMLNamespaceIsRecognised(t *testing.T) {
	// Binding the namespace to a prefix rather than defaulting it leaves the
	// xmlns attribute empty, so the resolved name carries the namespace instead.
	got, err := ProcessSBOMFile(fixture("cyclonedx-prefixed-namespace.xml"))

	require.NoError(t, err)
	assert.Equal(t, "cyclonedx-1.6", got.Format)
}
