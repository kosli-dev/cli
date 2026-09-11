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
		{"no bomFormat", "not-an-sbom.json", `bomFormat is ""`},
		{"no specVersion", "cyclonedx-no-spec-version.json", "no specVersion"},
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
		{"truncated", "truncated.json", "not valid JSON"},
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
	_, err := ProcessSBOMFile(fixture("spdx-null-package.json"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse the file as an SPDX SBOM")
}

func TestADocumentCarryingBothIdentitiesIsRejected(t *testing.T) {
	// The SPDX reader accepts anything carrying a version, so without the identity
	// check this would attest an empty document as a successful SBOM.
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
	// The checksum is uppercase in the file; lowercase is the only form the
	// server's schema accepts.
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
	// The server's schema enforces a hex pattern, which an empty string fails and
	// an absent value does not.
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
	got, err := ProcessSBOMFile(fixture("cyclonedx-prefixed-namespace.xml"))

	require.NoError(t, err)
	assert.Equal(t, "cyclonedx-1.6", got.Format)
}

func TestUnsupportedSPDXFormsSayWhatTheyAre(t *testing.T) {
	// Each of these is a valid SPDX document in a form this slice cannot read.
	// Without recognising them the RDF file is reported as a broken CycloneDX
	// file and the others as not an SBOM at all.
	for _, tc := range []struct{ name, file, wantErr string }{
		{"rdf", "spdx-rdf.rdf", "SPDX RDF document"},
		{"yaml", "spdx-yaml.yaml", "SPDX YAML document"},
		{"3.x json-ld", "spdx3-jsonld.json", "SPDX 3.x JSON-LD document"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ProcessSBOMFile(fixture(tc.file))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
			assert.Contains(t, err.Error(), "SPDX JSON or tag-value")
		})
	}
}

func TestATimestampTheAttestationCannotCarryIsRejected(t *testing.T) {
	// A bare date parses as CycloneDX but fails the server's date-time schema,
	// so it is caught here rather than after the upload.
	_, err := ProcessSBOMFile(fixture("cyclonedx-bad-timestamp.json"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), `"2020-04-13"`)
	assert.Contains(t, err.Error(), "RFC 3339")
}

func TestPackageCountExcludesTheSubject(t *testing.T) {
	// CycloneDX keeps the subject in metadata.component, outside the component
	// list, so it is not one of the packages counted.
	got, err := ProcessSBOMFile(fixture("cyclonedx-1.6.json"))

	require.NoError(t, err)
	require.NotNil(t, got.Document.Subject)
	assert.Equal(t, "Acme Application", got.Document.Subject.Name)
	assert.Equal(t, 2, got.Document.PackageCount)
}

func TestToolsAreNullWhenTheSBOMRecordsNone(t *testing.T) {
	// The fixture carries a tools block holding nothing, so the tool-reading path
	// runs and still yields null: the server's schema types tools as nullable,
	// and null distinguishes "none recorded" from "recorded as empty".
	got, err := ProcessSBOMFile(fixture("cyclonedx-empty-tools.json"))

	require.NoError(t, err)
	assert.Nil(t, got.Document.Tools)

	encoded, err := json.Marshal(got.Document)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"tools":null`)
}

func TestNullElementsInAnSPDX21DocumentDoNotCrash(t *testing.T) {
	// Only the 2.2 and 2.3 models normalise null elements while unmarshalling, so
	// in a 2.1 document they survive into code that dereferences them — including
	// the library's own described-package lookup, which runs outside the guard
	// around the reader.
	t.Run("null package is rejected", func(t *testing.T) {
		_, err := ProcessSBOMFile(fixture("spdx-2.1-null-package.json"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "null entry in its package list")
	})

	t.Run("null relationship is dropped, as the 2.3 reader drops it", func(t *testing.T) {
		got, err := ProcessSBOMFile(fixture("spdx-2.1-null-relationship.json"))

		require.NoError(t, err)
		assert.Equal(t, "spdx-2.1", got.Format)
		// One package and no usable relationship: the spec makes DESCRIBES
		// optional for a lone package, so it is still the subject.
		require.NotNil(t, got.Document.Subject)
		assert.Equal(t, "app", got.Document.Subject.Name)
	})
}

func TestTimestampsAreJudgedAsTheServerJudgesThem(t *testing.T) {
	// The server uppercases the value and matches rfc3339_validator, so the
	// lowercase forms RFC 3339 permits are accepted while a leap second is not:
	// that validator's own docstring says leap seconds are unsupported.
	t.Run("lowercase t and z are accepted", func(t *testing.T) {
		got, err := ProcessSBOMFile(fixture("cyclonedx-timestamp-lowercase-t-z.json"))

		require.NoError(t, err)
		require.NotNil(t, got.Document.CreatedAt)
		assert.Equal(t, "2020-04-13t20:20:39z", *got.Document.CreatedAt)
	})

	for _, tc := range []struct{ name, file, want string }{
		{"leap second", "cyclonedx-timestamp-leap-second.json", "2020-04-13T23:59:60Z"},
		{"year zero", "cyclonedx-timestamp-year-zero.json", "0000-01-01T00:00:00Z"},
		{"day its month does not have", "cyclonedx-timestamp-day-out-of-range.json", "2020-02-30T20:20:39Z"},
	} {
		t.Run(tc.name+" is refused", func(t *testing.T) {
			_, err := ProcessSBOMFile(fixture(tc.file))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}
