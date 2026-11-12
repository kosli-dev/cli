package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/intoto"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/sbom"
	"github.com/spf13/cobra"
)

// Both are also carried inside attestation_data, where the server's schema can
// enforce them. These are the copy a reader sees on the trail page.
const (
	sbomFormatAnnotation = "sbom_format"
	sbomSha256Annotation = "sbom_sha256"
)

type SbomAttestationData struct {
	Format              string         `json:"format"`
	OriginalFingerprint string         `json:"original_fingerprint"`
	Document            *sbom.Document `json:"document"`
}

type SbomAttestationPayload struct {
	*CommonAttestationPayload
	TypeName        string              `json:"type_name"`
	AttestationData SbomAttestationData `json:"attestation_data"`
}

type attestSbomOptions struct {
	*CommonAttestationOptions
	sbomFilePath string
	policy       intoto.Policy
	payload      SbomAttestationPayload
}

const attestSbomShortDesc = `Report a software bill of materials to an artifact or a trail in a Kosli flow.  `

const attestSbomLongDesc = attestSbomShortDesc + `
The SBOM file is given with the ^--sbom-file^ flag. CycloneDX (JSON and XML) and
SPDX (JSON and tag-value) are supported.

The file is uploaded as it is, so the recorded checksum is the checksum of the
file you supplied and you can verify it by hand. It must be a single file: it is
not compressed, and a gzipped file is rejected, because the format and the
summary below are read from it.

Kosli reads the format, the creation time, the tools that produced it, the
subject it describes and how many packages it lists. A plain SBOM file is not
checked against the artifact; it is recorded as reported.

The SBOM may also arrive inside an in-toto attestation, as a Sigstore bundle or
a DSSE envelope. The SPDX or CycloneDX document is then read from it, and the
bundle or envelope is what is uploaded and checksummed. When the artifact's
fingerprint is known, because it is calculated from the artifact or given with
^--fingerprint^, the attestation must list that artifact as its subject, signed
or not. When the attestation is bound by the artifact's template name and commit
instead, or to the trail alone, there is no fingerprint to check the subject
against, so it is not checked. To verify its signature, pass
^--certificate-identity^ (or ^--certificate-identity-regexp^) with
^--certificate-oidc-issuer^, or ^--key^. Without them the signature is not
checked, and even a signed bundle is recorded with ^verified^ set to false. The
record is annotated with ^verified^ and the verification details, as
^kosli attest in-toto^ annotates it. To pick the SBOM from a file holding
several attestations, or to record provenance, use ^kosli attest in-toto^.

` + sigstoreTrustRootDesc + `
The SBOM file is the only attachment: this command does not accept additional
attachments, because two or more would be compressed together.

The format and the file checksum are also added as the ^sbom_format^ and
^sbom_sha256^ annotations.
` + attestationBindingDesc + `

` + kosliIgnoreDesc + `

` + commitDescription

const attestSbomExample = `
# report an SBOM about a pre-built docker artifact (kosli finds the fingerprint):
kosli attest sbom yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--sbom-file yourSbomPath \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report an SBOM signed into a Sigstore bundle, verifying who signed it:
kosli attest sbom yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--sbom-file yourSbom.sigstore.json \
	--certificate-identity https://github.com/yourOrg/yourRepo/.github/workflows/build.yml@refs/heads/main \
	--certificate-oidc-issuer https://token.actions.githubusercontent.com \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report an SBOM about a trail:
kosli attest sbom \
	--name yourAttestationName \
	--sbom-file yourSbomPath \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestSbomCmd(out io.Writer) *cobra.Command {
	o := &attestSbomOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: SbomAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
			TypeName:                 "sbom",
		},
	}
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:     "sbom [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestSbomShortDesc,
		Long:    attestSbomLongDesc,
		Example: attestSbomExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := CustomMaximumNArgs(1, args)
			if err != nil {
				return err
			}

			err = RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"fingerprint", "artifact-type"}, false)
			if err != nil {
				return err
			}

			err = o.rejectAttachments(cmd)
			if err != nil {
				return err
			}

			err = o.rejectReservedAnnotations()
			if err != nil {
				return err
			}

			err = ValidateSliceValues(o.redactedCommitInfo, allowedCommitRedactionValues)
			if err != nil {
				return fmt.Errorf("%s for --redact-commit-info", err.Error())
			}

			err = ValidateAttestationArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = o.policy.Validate()
			if err != nil {
				return err
			}

			// A pinned signer means the file must be an attestation to verify,
			// so its keys are refused before anything is read or fetched.
			if o.policy.PinsSigner() {
				err = o.rejectVerificationAnnotations()
				if err != nil {
					return err
				}
			}

			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			o.repoURLExplicit = cmd.Flags().Changed("repo-url")
			o.repoNameExplicit = cmd.Flags().Changed("repository")
			return o.run(args)
		},
	}

	ci := WhichCI()
	addAttestationFlags(cmd, o.CommonAttestationOptions, o.payload.CommonAttestationPayload, ci)
	cmd.Flags().StringVar(&o.sbomFilePath, "sbom-file", "", attestationSbomFileFlag)
	addSignerPolicyFlags(cmd, &o.policy)
	// Required --sbom-file and mutually exclusive --attachments can never both be
	// satisfied, so the flag is hidden from help. Passing it still gets the error.
	_ = cmd.Flags().MarkHidden("attachments")

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "sbom-file"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestSbomOptions) run(args []string) error {
	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	content, err := o.loadSbom()
	if err != nil {
		return err
	}

	return newAttestationSubmitter().submit(attestationSubmission{
		flow:  o.flowName,
		trail: o.trailName,
		// The slug is the attestation family, not the type. The server tells
		// them apart by type_name in the body.
		slug:     "system",
		label:    "sbom",
		payload:  o.payload,
		evidence: o.evidence(content),
	})
}

func (o *attestSbomOptions) loadSbom() ([]byte, error) {
	// Checked before opening. open(2) on a fifo with no writer blocks, so a
	// check after it would never run; and a directory is what a user hits when
	// tab-completion stops a path short, so it gets its own advice.
	info, err := os.Stat(o.sbomFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("SBOM file [%s] is a directory; supply the SBOM file itself", o.sbomFilePath)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("SBOM file [%s] is not a regular file", o.sbomFilePath)
	}

	file, err := os.Open(o.sbomFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	defer func() { _ = file.Close() }()

	// One read serves the fingerprint, the summary and the upload, so the
	// recorded checksum describes the bytes the bucket receives even if the
	// file is still being written. The size is bounded by the org's upload
	// limit, which the server checks when the upload is minted.
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM file [%s]: %s", o.sbomFilePath, err)
	}

	document, evidence, err := o.unwrap(content)
	if err != nil {
		return nil, err
	}
	o.payload.AttestationData, err = sbomAttestationData(content, document)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	// After CommonAttestationOptions.run, which assigns payload.Annotations
	// wholesale from --annotate; called before it, both keys are discarded.
	o.payload.Annotations = withAnnotations(o.payload.Annotations, sbomAnnotations(o.payload.AttestationData))
	if evidence != nil {
		if err := o.recordVerification(evidence); err != nil {
			return nil, err
		}
	}
	return content, nil
}

// unwrap returns the SBOM document content holds. A Sigstore bundle or DSSE
// envelope holds it as an in-toto attestation, which must be an SPDX or
// CycloneDX one about the artifact, when there is one, and is verified when a
// signer is pinned; evidence then says what was established, and is nil for a
// plain SBOM file.
func (o *attestSbomOptions) unwrap(content []byte) (document []byte, evidence *intoto.Evidence, err error) {
	attestation, wrapped, err := intoto.ReadEnvelope(content)
	if errors.Is(err, intoto.ErrSeveralDocuments) {
		return nil, nil, fmt.Errorf("SBOM file [%s] %s. To record the SBOM from a file holding several attestations, such as an .intoto.jsonl file, use kosli attest in-toto --predicate-type sbom", o.sbomFilePath, err)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	if !wrapped {
		if o.policy.PinsSigner() {
			return nil, nil, fmt.Errorf("SBOM file [%s] is not a Sigstore bundle or DSSE envelope, so it has no signature to verify against --certificate-identity, --certificate-identity-regexp or --key", o.sbomFilePath)
		}
		return content, nil, nil
	}

	document, err = intoto.SBOMDocument(attestation)
	if err != nil {
		return nil, nil, fmt.Errorf("SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	if o.payload.ArtifactFingerprint != "" {
		if err := attestation.RequireSubject(o.payload.ArtifactFingerprint); err != nil {
			return nil, nil, fmt.Errorf("SBOM file [%s]: %s", o.sbomFilePath, err)
		}
	}
	if !o.policy.PinsSigner() {
		return document, &intoto.Evidence{}, nil
	}
	if !attestation.Signed() {
		return nil, nil, fmt.Errorf("SBOM file [%s] is not signed, so it cannot be verified against --certificate-identity, --certificate-identity-regexp or --key", o.sbomFilePath)
	}
	evidence, err = intoto.NewVerifier(o.policy, intotoTrustRoots()).Verify(attestation, o.payload.ArtifactFingerprint)
	if err != nil {
		return nil, nil, fmt.Errorf("SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	return document, evidence, nil
}

// recordVerification annotates the record with what verifying the
// attestation established, and links its transparency log entry. These keys
// are reserved only for an SBOM that arrives in an attestation, so a plain
// SBOM file can still be annotated with them.
func (o *attestSbomOptions) recordVerification(evidence *intoto.Evidence) error {
	if err := o.rejectVerificationAnnotations(); err != nil {
		return err
	}
	o.payload.Annotations = withAnnotations(o.payload.Annotations, verificationAnnotations(evidence))

	if evidence.LogEntryURL == "" {
		return nil
	}
	if _, taken := o.externalURLs[attestationExternalURL]; taken {
		return fmt.Errorf("external URL label '%s' is set by this command from the attestation and cannot be provided with --external-url", attestationExternalURL)
	}
	if o.payload.ExternalURLs == nil {
		o.payload.ExternalURLs = map[string]*URLInfo{}
	}
	o.payload.ExternalURLs[attestationExternalURL] = &URLInfo{Href: evidence.LogEntryURL, Fingerprint: evidence.RawSHA256}
	return nil
}

// sbomAttestationData summarises document, the SBOM itself, and fingerprints
// attached, the bytes uploaded with the attestation, so the recorded checksum
// is one a reader can check against the attachment. The two are the same
// bytes unless the SBOM arrives inside an attestation bundle, which is then
// what is attached.
func sbomAttestationData(attached, document []byte) (SbomAttestationData, error) {
	data, err := sbom.ProcessSBOM(document)
	if err != nil {
		return SbomAttestationData{}, err
	}
	return SbomAttestationData{
		Format:              data.Format,
		OriginalFingerprint: fmt.Sprintf("%x", sha256.Sum256(attached)),
		Document:            data.Document,
	}, nil
}

// sbomAnnotations records what the SBOM said about itself where a reader sees
// it on the trail page. The same two values are carried inside
// attestation_data.
func sbomAnnotations(data SbomAttestationData) map[string]string {
	return map[string]string{
		sbomFormatAnnotation: data.Format,
		sbomSha256Annotation: data.OriginalFingerprint,
	}
}

// withAnnotations returns base with extra added, in a new map. base is often
// the --annotate flag's own map -- processAnnotations returns its argument --
// so writing into it would put derived keys into the map a reserved-key check
// reads, and a second run of the same options would be refused for a key
// nobody supplied.
func withAnnotations(base, extra map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(extra))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

// rejectAttachments refuses --attachments however it was set. Two or more
// attachments are tarred and gzipped before upload, which would compress the
// SBOM and break the checksum recorded against it. The flag is hidden from help,
// and a value from KOSLI_ATTACHMENTS or a config file marks it Changed without
// anyone typing it, so in that case the message says where the value came from.
func (o *attestSbomOptions) rejectAttachments(cmd *cobra.Command) error {
	if !cmd.Flags().Changed("attachments") {
		return nil
	}
	source := ""
	if flagCameFromConfig("attachments") {
		source = " (set by " + configValueSource("attachments") + ")"
	}
	return fmt.Errorf(
		"--attachments cannot be used with attest sbom%s: the SBOM file is the only attachment, and a second one would be compressed",
		source,
	)
}

// evidence is the attachment: the SBOM bytes that were hashed, never a path. Handing the uploader a path would make it open and
// read the file a second time, and a file still being written would then be
// uploaded as different bytes from the ones sbom_sha256 describes.
//
// The bytes are a parameter rather than a field, so the attachment cannot be
// built before the file has been read. Built from an empty field, it would be
// a well-formed request carrying no attachment and a fingerprint of nothing.
func (o *attestSbomOptions) evidence(content []byte) *requests.FileBytes {
	return &requests.FileBytes{
		Name: filepath.Base(o.sbomFilePath),
		Data: content,
	}
}

// rejectVerificationAnnotations refuses the --annotate keys that recording an
// attestation's verification sets.
func (o *attestSbomOptions) rejectVerificationAnnotations() error {
	for _, reserved := range verificationAnnotationKeys {
		if _, taken := o.annotations[reserved]; taken {
			return fmt.Errorf("annotation key '%s' is set by this command from the attestation and cannot be provided with --annotate", reserved)
		}
	}
	return nil
}

// rejectReservedAnnotations runs in PreRunE: it needs nothing from the file, so
// a typo should not cost a repository walk and a pass over the whole file first.
func (o *attestSbomOptions) rejectReservedAnnotations() error {
	for _, reserved := range []string{sbomFormatAnnotation, sbomSha256Annotation} {
		if _, taken := o.annotations[reserved]; taken {
			return fmt.Errorf(
				"annotation key '%s' is set by this command from the SBOM file and cannot be provided with --annotate",
				reserved,
			)
		}
	}
	return nil
}
