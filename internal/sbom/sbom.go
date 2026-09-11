// Package sbom reads CycloneDX and SPDX software bills of materials and
// normalises the few fields every format has in common.
package sbom

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	spdxjson "github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
	"github.com/spdx/tools-golang/spdxlib"
	"github.com/spdx/tools-golang/tagvalue"
)

// Subject is the single artifact an SBOM describes, when it names one.
// Every field a document may omit is a pointer, so a consumer can tell a value
// the SBOM did not carry from one it carried as empty.
type Subject struct {
	Name    string  `json:"name"`
	Version *string `json:"version"`
	Purl    *string `json:"purl"`
	Sha256  *string `json:"sha256"`
}

// Document is the normalised summary shared by every supported format.
type Document struct {
	CreatedAt    *string  `json:"created_at"`
	Tools        []string `json:"tools"`
	Subject      *Subject `json:"subject"`
	PackageCount int      `json:"package_count"`
}

// SBOMData is what an attestation carries about the file.
type SBOMData struct {
	Format   string    `json:"format"`
	Document *Document `json:"document"`
}

var (
	gzipMagic = []byte{0x1f, 0x8b}
	utf8BOM   = []byte{0xEF, 0xBB, 0xBF}
)

// The declared version is read from the bytes rather than the parsed document:
// the SPDX readers convert every document up to their newest model, so a 2.2
// file reports itself as 2.3 once parsed.
var (
	spdxVersionJSON     = regexp.MustCompile(`"spdxVersion"\s*:\s*"SPDX-(\d+\.\d+)"`)
	spdxVersionTagValue = regexp.MustCompile(`(?m)^\s*SPDXVersion:\s*SPDX-(\d+\.\d+)\s*$`)
	// A tag-value document may quote another document's header inside a text
	// block, which would otherwise be read as its own version.
	tagValueTextBlock  = regexp.MustCompile(`(?s)<text>.*?</text>`)
	cycloneDXNamespace = regexp.MustCompile(`^https?://cyclonedx\.org/schema/bom/(\d+\.\d+)$`)
)

// ProcessSBOMFile reads an SBOM file and returns its format and a normalised
// summary. It confirms the file identifies itself as the format it parses as;
// it does not validate against the format's schema.
func ProcessSBOMFile(file string) (*SBOMData, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return processSBOM(content)
}

func processSBOM(content []byte) (*SBOMData, error) {
	if bytes.HasPrefix(content, gzipMagic) {
		return nil, fmt.Errorf("the file is gzip compressed; supply the uncompressed SBOM")
	}
	// Stripped from the content itself, not just from the detection: both
	// parsers reject a byte-order mark as an unexpected character.
	content = bytes.TrimPrefix(content, utf8BOM)

	trimmed := bytes.TrimLeft(content, " \t\r\n")
	switch {
	case bytes.HasPrefix(trimmed, []byte("<")):
		return readCycloneDXXML(content)
	case bytes.HasPrefix(trimmed, []byte("{")):
		if version := declaredSPDXVersion(spdxVersionJSON, content); version != "" {
			return readSPDX(spdxjson.Read, content, version)
		}
		return readCycloneDXJSON(content)
	default:
		if version := declaredSPDXVersion(spdxVersionTagValue, tagValueTextBlock.ReplaceAll(content, nil)); version != "" {
			return readSPDX(tagvalue.Read, content, version)
		}
	}
	return nil, fmt.Errorf("unrecognised file: expected a CycloneDX or SPDX SBOM in JSON, XML or tag-value form")
}

func declaredSPDXVersion(re *regexp.Regexp, content []byte) string {
	if match := re.FindSubmatch(content); match != nil {
		return string(match[1])
	}
	return ""
}

func readCycloneDXJSON(content []byte) (*SBOMData, error) {
	bom, err := decodeCycloneDX(content, cdx.BOMFileFormatJSON)
	if err != nil {
		return nil, err
	}
	// The decoder accepts any well-formed JSON, returning an empty document
	// rather than an error, so these two fields are what identify a CycloneDX
	// file. Both are JSON-only.
	if bom.BOMFormat != "CycloneDX" {
		return nil, fmt.Errorf("not a CycloneDX SBOM: bomFormat is %q, expected \"CycloneDX\"", bom.BOMFormat)
	}
	if bom.SpecVersion == 0 {
		return nil, fmt.Errorf("not a CycloneDX SBOM: no specVersion")
	}
	return &SBOMData{
		Format:   "cyclonedx-" + bom.SpecVersion.String(),
		Document: documentFromCycloneDX(bom),
	}, nil
}

func readCycloneDXXML(content []byte) (*SBOMData, error) {
	bom, err := decodeCycloneDX(content, cdx.BOMFileFormatXML)
	if err != nil {
		return nil, err
	}
	// bomFormat and specVersion are excluded from the XML mapping, so an XML
	// SBOM identifies itself by its namespace, which is also where its version
	// is. XMLNS holds a default namespace; XMLName.Space resolves a prefixed one.
	namespace := bom.XMLNS
	if namespace == "" {
		namespace = bom.XMLName.Space
	}
	match := cycloneDXNamespace.FindStringSubmatch(namespace)
	if match == nil {
		return nil, fmt.Errorf("not a CycloneDX SBOM: xmlns is %q, expected a http://cyclonedx.org/schema/bom/ namespace", namespace)
	}
	return &SBOMData{
		Format:   "cyclonedx-" + match[1],
		Document: documentFromCycloneDX(bom),
	}, nil
}

func decodeCycloneDX(content []byte, format cdx.BOMFileFormat) (*cdx.BOM, error) {
	bom := new(cdx.BOM)
	if err := cdx.NewBOMDecoder(bytes.NewReader(content), format).Decode(bom); err != nil {
		return nil, fmt.Errorf("could not parse the file as a CycloneDX SBOM: %w", err)
	}
	return bom, nil
}

func readSPDX(read func(r io.Reader) (*spdx.Document, error), content []byte, declaredVersion string) (data *SBOMData, err error) {
	// A malformed document can make the SPDX readers dereference a nil element
	// rather than return an error; a file the user supplied must not end the
	// process. A panic inside the reader means it could not handle this input.
	defer func() {
		if recovered := recover(); recovered != nil {
			data, err = nil, fmt.Errorf("could not parse the file as an SPDX SBOM: %v", recovered)
		}
	}()

	doc, err := read(bytes.NewReader(content))
	if err != nil {
		// The tag-value parser rejects a snippet section that follows a package,
		// which SPDX's own 2.2 example does. The JSON form of the same document
		// reads without complaint. The wording below is matched against the
		// upstream error, and the test for it is what catches a reworded bump.
		if strings.Contains(err.Error(), "unknown tag Snippet") {
			return nil, fmt.Errorf("this SPDX tag-value document has a snippet section, which the SPDX tag-value parser cannot read; supply the same SBOM in JSON form instead")
		}
		return nil, fmt.Errorf("could not parse the file as an SPDX SBOM: %w", err)
	}
	// The readers accept any document carrying a version, returning an empty
	// one rather than an error, so the identifier is what says this is SPDX.
	if doc.SPDXIdentifier == "" {
		return nil, fmt.Errorf("not an SPDX SBOM: no SPDXID")
	}
	return &SBOMData{
		Format:   "spdx-" + declaredVersion,
		Document: documentFromSPDX(doc),
	}, nil
}

func documentFromCycloneDX(bom *cdx.BOM) *Document {
	doc := &Document{PackageCount: packageCount(bom.Components)}
	if bom.Metadata != nil {
		doc.CreatedAt = nullIfEmpty(bom.Metadata.Timestamp)
		doc.Tools = toolsFromCycloneDX(bom.Metadata.Tools)
		doc.Subject = subjectFromComponent(bom.Metadata.Component)
	}
	return doc
}

// packageCount counts every component that is not a file, at any depth. Some
// generators, syft among them, emit one component per file in the scanned
// image, which would otherwise outnumber the packages by an order of magnitude.
// An assembly is itself a package, so it counts alongside what it contains;
// SPDX counts a containing package the same way.
func packageCount(components *[]cdx.Component) int {
	if components == nil {
		return 0
	}
	count := 0
	for _, component := range *components {
		if component.Type != cdx.ComponentTypeFile {
			count++
		}
		count += packageCount(component.Components)
	}
	return count
}

// toolsFromCycloneDX reads both tool layouts. Spec 1.5 moved tools from a
// dedicated list to components, and the library keeps the older list populated
// for documents that use it, so a document may fill either. The services slot
// the same spec added is deliberately skipped: a service a document consumed is
// not a tool that generated it.
func toolsFromCycloneDX(tools *cdx.ToolsChoice) []string {
	if tools == nil {
		return nil
	}
	var names []string
	if tools.Tools != nil {
		for _, tool := range *tools.Tools {
			names = append(names, nameAndVersion(tool.Name, tool.Version))
		}
	}
	if tools.Components != nil {
		for _, component := range *tools.Components {
			names = append(names, nameAndVersion(component.Name, component.Version))
		}
	}
	return names
}

func nameAndVersion(name, version string) string {
	if version == "" {
		return name
	}
	return name + " " + version
}

func subjectFromComponent(component *cdx.Component) *Subject {
	if component == nil {
		return nil
	}
	subject := &Subject{
		Name:    component.Name,
		Version: nullIfEmpty(component.Version),
		Purl:    nullIfEmpty(component.PackageURL),
	}
	if component.Hashes != nil {
		for _, hash := range *component.Hashes {
			if hash.Algorithm == cdx.HashAlgoSHA256 {
				subject.Sha256 = nullIfEmpty(strings.ToLower(hash.Value))
				break
			}
		}
	}
	return subject
}

func documentFromSPDX(doc *spdx.Document) *Document {
	out := &Document{PackageCount: len(doc.Packages)}
	if doc.CreationInfo != nil {
		out.CreatedAt = nullIfEmpty(doc.CreationInfo.Created)
		for _, creator := range doc.CreationInfo.Creators {
			if creator.CreatorType == "Tool" {
				out.Tools = append(out.Tools, creator.Creator)
			}
		}
	}
	out.Subject = subjectFromSPDX(doc)
	return out
}

// subjectFromSPDX returns a subject only when the document describes exactly
// one package. Describing several is normal and legitimate, and picking one of
// them would be a guess. A lone package needs no DESCRIBES relationship: the
// spec makes one mandatory only when a document holds more than one package.
func subjectFromSPDX(doc *spdx.Document) *Subject {
	described, err := spdxlib.GetDescribedPackageIDs(doc)
	// Every error here means the described package cannot be determined, which
	// is the same answer as describing several.
	if err != nil || len(described) != 1 {
		return nil
	}
	for _, pkg := range doc.Packages {
		if pkg == nil || pkg.PackageSPDXIdentifier != described[0] {
			continue
		}
		subject := &Subject{
			Name:    pkg.PackageName,
			Version: nullIfEmpty(pkg.PackageVersion),
			Purl:    purlFromSPDX(pkg),
		}
		for _, checksum := range pkg.PackageChecksums {
			if checksum.Algorithm == common.SHA256 {
				subject.Sha256 = nullIfEmpty(strings.ToLower(checksum.Value))
				break
			}
		}
		return subject
	}
	return nil
}

func purlFromSPDX(pkg *spdx.Package) *string {
	for _, ref := range pkg.PackageExternalReferences {
		if ref != nil && ref.RefType == common.TypePackageManagerPURL {
			return nullIfEmpty(ref.Locator)
		}
	}
	return nil
}

// nullIfEmpty keeps a value the SBOM did not carry distinguishable from one it
// carried as empty: the attestation records the former as null.
func nullIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
