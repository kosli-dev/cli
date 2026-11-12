// Package intoto reads in-toto attestations, verifies the signed ones against a
// pinned signer, and maps their predicates onto the records Kosli keeps.
//
// The command layer only chooses a source and builds a request from what this
// package returns, so other callers can use it directly.
package intoto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	ita1 "github.com/in-toto/attestation/go/v1"
	"github.com/secure-systems-lab/go-securesystemslib/dsse"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	inTotoPayloadType     = "application/vnd.in-toto+json"
	bundleMediaTypePrefix = "application/vnd.dev.sigstore.bundle"
	statementTypePrefix   = "https://in-toto.io/Statement/"
)

// Attestation is one in-toto statement read from a source, together with the
// signature over it, if it has one.
type Attestation struct {
	// Raw is the bytes the attestation was read from: the whole file, or one
	// line of a file holding several.
	Raw []byte
	// Position is the attestation's 1-based place in a file holding several,
	// and 0 in a file holding one.
	Position int
	// Origin names an attestation that was not read from a file, by its digest
	// and how it is attached to an image.
	Origin    string
	Statement *ita1.Statement
	// Predicate is the predicate exactly as the payload carried it. The
	// Statement's copy has been through protobuf's Struct, which turns every
	// number into a float64.
	Predicate json.RawMessage

	bundle   *bundle.Bundle
	envelope *dsse.Envelope
}

// Signed reports whether the attestation carries a signature to verify.
func (a *Attestation) Signed() bool {
	if a.bundle != nil {
		return true
	}
	return a.envelope != nil && len(a.envelope.Signatures) > 0
}

// Describe names the attestation in a message: by its origin when it has one,
// or by its place in the file when the file holds several.
func (a *Attestation) Describe() string {
	if a.Origin != "" {
		return "attestation " + a.Origin
	}
	if a.Position == 0 {
		return "the attestation"
	}
	return fmt.Sprintf("attestation %d", a.Position)
}

// Read parses content holding one attestation, or several one after another
// as in a .intoto.jsonl file. Each is a Sigstore bundle, a DSSE envelope, or a
// bare in-toto statement. Nothing is verified here.
func Read(content []byte) ([]*Attestation, error) {
	documents, err := splitDocuments(content)
	if err != nil {
		return nil, err
	}

	attestations := make([]*Attestation, 0, len(documents))
	for i, document := range documents {
		attestation := &Attestation{Raw: document}
		if len(documents) == 1 {
			// The whole file, so a checksum of Raw is a checksum of the file.
			attestation.Raw = content
		} else {
			attestation.Position = i + 1
		}
		if err := attestation.parse(document); err != nil {
			return nil, fmt.Errorf("%s: %w", attestation.Describe(), err)
		}
		attestations = append(attestations, attestation)
	}
	return attestations, nil
}

// ErrSeveralDocuments is ReadEnvelope's error for content holding more than
// one JSON document, such as an .intoto.jsonl file, which is neither a single
// bundle or envelope nor a document of its own.
var ErrSeveralDocuments = errors.New("holds several JSON documents, not a single bundle or envelope")

// ReadEnvelope reads content as a single Sigstore bundle or DSSE envelope.
// wrapped is false, with no error, for anything else, so a caller can treat
// that content as the document itself; a malformed bundle or envelope is
// still an error, and so is ErrSeveralDocuments.
func ReadEnvelope(content []byte) (attestation *Attestation, wrapped bool, err error) {
	var probe struct {
		MediaType   string `json:"mediaType"`
		PayloadType string `json:"payloadType"`
	}
	if json.Unmarshal(content, &probe) != nil {
		if documents, err := splitDocuments(content); err == nil && len(documents) > 1 {
			return nil, false, ErrSeveralDocuments
		}
		return nil, false, nil
	}
	if !strings.HasPrefix(probe.MediaType, bundleMediaTypePrefix) && probe.PayloadType == "" {
		return nil, false, nil
	}
	attestations, err := Read(content)
	if err != nil {
		return nil, true, err
	}
	return attestations[0], true, nil
}

func splitDocuments(content []byte) ([]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	var documents []json.RawMessage
	for {
		var document json.RawMessage
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("not valid JSON: %w", err)
		}
		documents = append(documents, document)
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("holds no attestation")
	}
	return documents, nil
}

func (a *Attestation) parse(document []byte) error {
	var probe struct {
		MediaType   string `json:"mediaType"`
		PayloadType string `json:"payloadType"`
		Type        string `json:"_type"`
	}
	if err := json.Unmarshal(document, &probe); err != nil {
		return fmt.Errorf("expected a JSON object: %w", err)
	}

	switch {
	case strings.HasPrefix(probe.MediaType, bundleMediaTypePrefix):
		return a.parseBundle(document)
	case probe.PayloadType != "":
		var envelope dsse.Envelope
		if err := json.Unmarshal(document, &envelope); err != nil {
			return fmt.Errorf("not a valid DSSE envelope: %w", err)
		}
		return a.parseEnvelope(&envelope)
	case strings.HasPrefix(probe.Type, statementTypePrefix):
		return a.parseStatement(document)
	}
	return fmt.Errorf("not a Sigstore bundle, a DSSE envelope or an in-toto statement")
}

func (a *Attestation) parseBundle(document []byte) error {
	var b bundle.Bundle
	if err := b.UnmarshalJSON(document); err != nil {
		return fmt.Errorf("not a valid Sigstore bundle: %w", err)
	}
	envelope, err := b.Envelope()
	if err != nil {
		return fmt.Errorf("the Sigstore bundle signs a blob rather than an in-toto attestation")
	}
	a.bundle = &b
	return a.parseEnvelope(envelope.RawEnvelope())
}

func (a *Attestation) parseEnvelope(envelope *dsse.Envelope) error {
	if envelope.PayloadType != inTotoPayloadType {
		return fmt.Errorf("the DSSE envelope carries %q, not an in-toto statement (%s)", envelope.PayloadType, inTotoPayloadType)
	}
	payload, err := envelope.DecodeB64Payload()
	if err != nil {
		return fmt.Errorf("the DSSE envelope's payload is not valid base64: %w", err)
	}
	a.envelope = envelope
	return a.parseStatement(payload)
}

func (a *Attestation) parseStatement(payload []byte) error {
	statement := &ita1.Statement{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(payload, statement); err != nil {
		return fmt.Errorf("not a valid in-toto statement: %w", err)
	}
	if err := statement.Validate(); err != nil {
		return fmt.Errorf("not a valid in-toto statement: %w", err)
	}
	var raw struct {
		Predicate json.RawMessage `json:"predicate"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return fmt.Errorf("not a valid in-toto statement: %w", err)
	}
	a.Statement = statement
	a.Predicate = raw.Predicate
	return nil
}
