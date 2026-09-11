// Package sbom reads CycloneDX and SPDX software bills of materials and
// normalises the few fields every format has in common.
package sbom

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

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

// The declared version is read from the file rather than the parsed document:
// the SPDX readers convert every document up to their newest model, so a 2.2
// file reports itself as 2.3 once parsed.
var (
	spdxVersionTagValue = regexp.MustCompile(`(?m)^\s*SPDXVersion:\s*SPDX-(\d+\.\d+)\s*$`)
	// A top-level YAML key sits at column zero, which is what separates this
	// from the same word appearing inside a value.
	spdxVersionYAML = regexp.MustCompile(`(?m)^spdxVersion:\s*"?SPDX-`)
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
	// Stripped from the content, not just the detection: both parsers reject a
	// byte-order mark as an unexpected character.
	content = bytes.TrimPrefix(content, utf8BOM)

	trimmed := bytes.TrimLeft(content, " \t\r\n")
	switch {
	case bytes.HasPrefix(trimmed, []byte("<")):
		if xmlRootElement(content) == "RDF" {
			return nil, unsupportedSPDXForm("RDF")
		}
		return readCycloneDXXML(content)
	case bytes.HasPrefix(trimmed, []byte("{")):
		var probe struct {
			SPDXVersion string          `json:"spdxVersion"`
			Context     json.RawMessage `json:"@context"`
		}
		_ = json.Unmarshal(content, &probe)
		if bytes.Contains(probe.Context, []byte("spdx.org")) {
			return nil, unsupportedSPDXForm("3.x JSON-LD")
		}
		if probe.SPDXVersion != "" {
			// The declared text, not the parsed document's. The reader accepts
			// only an exact SPDX-2.1/2.2/2.3, so nothing else reaches this.
			return readSPDX(spdxjson.Read, content, strings.TrimPrefix(probe.SPDXVersion, "SPDX-"))
		}
		return readCycloneDXJSON(content)
	default:
		if spdxVersionYAML.Match(content) {
			return nil, unsupportedSPDXForm("YAML")
		}
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
	document, err := documentFromCycloneDX(bom)
	if err != nil {
		return nil, err
	}
	return &SBOMData{Format: "cyclonedx-" + bom.SpecVersion.String(), Document: document}, nil
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
	document, err := documentFromCycloneDX(bom)
	if err != nil {
		return nil, err
	}
	return &SBOMData{Format: "cyclonedx-" + match[1], Document: document}, nil
}

func decodeCycloneDX(content []byte, format cdx.BOMFileFormat) (*cdx.BOM, error) {
	bom := new(cdx.BOM)
	if err := cdx.NewBOMDecoder(bytes.NewReader(content), format).Decode(bom); err != nil {
		return nil, fmt.Errorf("could not parse the file as a CycloneDX SBOM: %w", err)
	}
	return bom, nil
}

func readSPDX(read func(r io.Reader) (*spdx.Document, error), content []byte, declaredVersion string) (*SBOMData, error) {
	doc, err := safeSPDXRead(read, content)
	if err != nil {
		// The tag-value parser rejects a snippet section that follows a package,
		// which SPDX's own 2.2 example does. The JSON form of the same document
		// reads without complaint.
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
	if err := normaliseNullElements(doc); err != nil {
		return nil, err
	}
	document, err := documentFromSPDX(doc)
	if err != nil {
		return nil, err
	}
	return &SBOMData{Format: "spdx-" + declaredVersion, Document: document}, nil
}

// safeSPDXRead guards the library call alone. A malformed document can make the
// readers dereference a nil element rather than return an error, and a file the
// user supplied must not end the process. Widening this to our own mapping code
// would report a defect here as the user's file being bad.
// normaliseNullElements gives every version the handling 2.2 and 2.3 get while
// unmarshalling: they drop null relationships and reject a null package. The 2.1
// model has no such hook, so nulls survive conversion and reach code that
// dereferences them, the library's own described-package lookup included.
func normaliseNullElements(doc *spdx.Document) error {
	for _, pkg := range doc.Packages {
		if pkg == nil {
			return fmt.Errorf("could not parse the file as an SPDX SBOM: a null entry in its package list")
		}
	}
	relationships := make([]*spdx.Relationship, 0, len(doc.Relationships))
	for _, relationship := range doc.Relationships {
		if relationship != nil {
			relationships = append(relationships, relationship)
		}
	}
	doc.Relationships = relationships
	return nil
}

func safeSPDXRead(read func(r io.Reader) (*spdx.Document, error), content []byte) (doc *spdx.Document, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			doc, err = nil, fmt.Errorf("%v", recovered)
		}
	}()
	return read(bytes.NewReader(content))
}

func documentFromCycloneDX(bom *cdx.BOM) (*Document, error) {
	doc := &Document{PackageCount: packageCount(bom.Components)}
	if bom.Metadata == nil {
		return doc, nil
	}
	createdAt, err := createdAt(bom.Metadata.Timestamp)
	if err != nil {
		return nil, err
	}
	doc.CreatedAt = createdAt
	doc.Tools = toolsFromCycloneDX(bom.Metadata.Tools)
	doc.Subject = subjectFromComponent(bom.Metadata.Component)
	return doc, nil
}

// createdAt rejects a timestamp the attestation could not carry. The server's
// schema types this field as a date-time, so a generator emitting a bare date or
// a local time with no offset would fail there instead, after the upload.
func createdAt(timestamp string) (*string, error) {
	if timestamp == "" {
		return nil, nil
	}
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		return nil, fmt.Errorf("the SBOM's creation timestamp %q is not an RFC 3339 date-time", timestamp)
	}
	return &timestamp, nil
}

// packageCount counts every component that is not a file, at any depth. Some
// generators, syft among them, emit one component per file in the scanned
// image, which would otherwise outnumber the packages by an order of magnitude.
// An assembly is itself a package, so it counts alongside what it contains. The
// subject is not counted: CycloneDX holds it in metadata.component, outside this
// list. SPDX has no such split, so its described package is counted, and the two
// formats differ by one for the same logical SBOM.
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

func documentFromSPDX(doc *spdx.Document) (*Document, error) {
	out := &Document{PackageCount: len(doc.Packages)}
	if doc.CreationInfo != nil {
		created, err := createdAt(doc.CreationInfo.Created)
		if err != nil {
			return nil, err
		}
		out.CreatedAt = created
		for _, creator := range doc.CreationInfo.Creators {
			if creator.CreatorType == "Tool" {
				out.Tools = append(out.Tools, creator.Creator)
			}
		}
	}
	out.Subject = subjectFromSPDX(doc)
	return out, nil
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

func nullIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func xmlRootElement(content []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	for {
		token, err := decoder.Token()
		if err != nil {
			return ""
		}
		if start, ok := token.(xml.StartElement); ok {
			return start.Name.Local
		}
	}
}

func unsupportedSPDXForm(form string) error {
	return fmt.Errorf("this is an SPDX %s document, which is not supported; supply the SBOM as SPDX JSON or tag-value", form)
}
