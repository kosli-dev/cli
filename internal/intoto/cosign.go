package intoto

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protocommon "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	protorekor "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/tlog"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
)

// Before its bundle format, cosign stored an attestation as a DSSE envelope
// layer and kept what verifies the signature in the layer's annotations.
const (
	cosignEnvelopeMediaType     = "application/vnd.dsse.envelope.v1+json"
	cosignCertificateAnnotation = "dev.sigstore.cosign/certificate"
	cosignBundleAnnotation      = "dev.sigstore.cosign/bundle"
	cosignTimestampAnnotation   = "dev.sigstore.cosign/rfc3161timestamp"
)

// cosignBundleVersion is the Sigstore bundle version that, like cosign's
// annotations, vouches for a transparency log entry with the log's signed
// entry timestamp alone. Later versions require an inclusion proof, which
// cosign did not store.
const cosignBundleVersion = "0.1"

// addCosignMaterial gives an attestation read from a DSSE envelope layer the
// verification material cosign stored in the layer's annotations, as the
// Sigstore bundle a later cosign would have written, so that it is verified
// the same way. Material that cannot be used is kept as the reason a
// certificate's signature cannot be checked; the envelope can still be
// verified with a key.
func (a *Attestation) addCosignMaterial(annotations map[string]string) {
	a.bundle, a.materialErr = cosignBundle(a, annotations)
}

// cosignBundle returns nil when the annotations hold no material: the
// envelope was signed with a key and never logged.
func cosignBundle(a *Attestation, annotations map[string]string) (*bundle.Bundle, error) {
	material := &protobundle.VerificationMaterial{
		Content: &protobundle.VerificationMaterial_PublicKey{PublicKey: &protocommon.PublicKeyIdentifier{}},
	}
	certificate := annotations[cosignCertificateAnnotation]
	if certificate != "" {
		certificates, err := cryptoutils.UnmarshalCertificatesFromPEM([]byte(certificate))
		if err != nil || len(certificates) == 0 {
			return nil, fmt.Errorf("the certificate cosign stored with it is not a PEM certificate")
		}
		material.Content = &protobundle.VerificationMaterial_X509CertificateChain{X509CertificateChain: &protocommon.X509CertificateChain{
			Certificates: []*protocommon.X509Certificate{{RawBytes: certificates[0].Raw}},
		}}
	}
	if logged := annotations[cosignBundleAnnotation]; logged != "" {
		entry, err := cosignLogEntry(logged)
		if err != nil {
			return nil, fmt.Errorf("the transparency log entry cosign stored with it cannot be read: %w", err)
		}
		material.TlogEntries = []*protorekor.TransparencyLogEntry{entry}
	}
	if stamped := annotations[cosignTimestampAnnotation]; stamped != "" {
		var timestamp struct {
			SignedRFC3161Timestamp []byte
		}
		if err := json.Unmarshal([]byte(stamped), &timestamp); err != nil {
			return nil, fmt.Errorf("the signed timestamp cosign stored with it cannot be read: %w", err)
		}
		material.TimestampVerificationData = &protobundle.TimestampVerificationData{
			Rfc3161Timestamps: []*protocommon.RFC3161SignedTimestamp{{SignedTimestamp: timestamp.SignedRFC3161Timestamp}},
		}
	}
	if certificate == "" && len(material.TlogEntries) == 0 && material.TimestampVerificationData == nil {
		return nil, nil
	}

	signed, err := envelopeAsBundle(a)
	if err != nil {
		return nil, err
	}
	mediaType, err := bundle.MediaTypeString(cosignBundleVersion)
	if err != nil {
		return nil, err
	}
	signed.MediaType = mediaType
	signed.VerificationMaterial = material
	return bundle.NewBundle(signed.Bundle)
}

// cosignLogEntry reads the Rekor entry cosign stored as it received it from
// the log, into the form a Sigstore bundle carries one in.
func cosignLogEntry(annotation string) (*protorekor.TransparencyLogEntry, error) {
	var logged struct {
		SignedEntryTimestamp []byte
		Payload              struct {
			Body           string `json:"body"`
			IntegratedTime int64  `json:"integratedTime"`
			LogIndex       int64  `json:"logIndex"`
			LogID          string `json:"logID"`
		}
	}
	if err := json.Unmarshal([]byte(annotation), &logged); err != nil {
		return nil, err
	}
	body, err := base64.StdEncoding.DecodeString(logged.Payload.Body)
	if err != nil {
		return nil, fmt.Errorf("its body is not valid base64: %w", err)
	}
	logID, err := hex.DecodeString(logged.Payload.LogID)
	if err != nil {
		return nil, fmt.Errorf("its log ID is not hexadecimal: %w", err)
	}
	var kind struct {
		Kind       string `json:"kind"`
		APIVersion string `json:"apiVersion"`
	}
	if err := json.Unmarshal(body, &kind); err != nil {
		return nil, fmt.Errorf("its body is not valid JSON: %w", err)
	}

	entry := &protorekor.TransparencyLogEntry{
		LogIndex:          logged.Payload.LogIndex,
		LogId:             &protocommon.LogId{KeyId: logID},
		KindVersion:       &protorekor.KindVersion{Kind: kind.Kind, Version: kind.APIVersion},
		IntegratedTime:    logged.Payload.IntegratedTime,
		InclusionPromise:  &protorekor.InclusionPromise{SignedEntryTimestamp: logged.SignedEntryTimestamp},
		CanonicalizedBody: body,
	}
	// Older cosign logged attestations as Rekor intoto 0.0.1 entries, which
	// sigstore-go does not check.
	if _, err := tlog.ParseTransparencyLogEntry(entry); err != nil {
		return nil, fmt.Errorf("it is a Rekor %s %s entry: %w", kind.Kind, kind.APIVersion, err)
	}
	return entry, nil
}
