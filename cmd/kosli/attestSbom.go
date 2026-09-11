package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/sbom"
	"github.com/spf13/cobra"
)

// The API rejects a request body over 10MB, and the JSON payload is counted
// alongside the file, so this leaves room for it rather than sitting on the
// limit. Lifting the ceiling needs direct-to-S3 upload, tracked separately.
const maxSbomFileBytes = 9 * 1024 * 1024

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
	payload      SbomAttestationPayload
}

const attestSbomShortDesc = `Report a software bill of materials to an artifact or a trail in a Kosli flow. `

const attestSbomLongDesc = attestSbomShortDesc + `
The SBOM file is given with the ^--sbom-file^ flag. CycloneDX (JSON and XML) and
SPDX (JSON and tag-value) are supported.

The file is uploaded as it is, so the recorded checksum is the checksum of the
file you supplied and you can verify it by hand. It must be a single file: it is
not compressed, and an already-compressed file is rejected, because the format
and the summary below are read from it.

Kosli reads the format, the creation time, the tools that produced it, the
subject it describes and how many packages it lists. Nothing is checked against
the artifact; the SBOM is recorded as reported.

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
		Use:         "sbom [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:       attestSbomShortDesc,
		Long:        attestSbomLongDesc,
		Example:     attestSbomExample,
		Annotations: map[string]string{betaCLIAnnotation: ""},
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

			// One file per SBOM attestation. Two or more attachments are tarred
			// and gzipped before upload, which would compress the SBOM and break
			// the checksum recorded against it.
			err = MuXRequiredFlags(cmd, []string{"sbom-file", "attachments"}, false)
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

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "sbom-file"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestSbomOptions) run(args []string) error {
	// The slug is the attestation family, not the type. The server tells them
	// apart by type_name in the body.
	url, err := url.JoinPath(global.Host, "api/v2/attestations", global.Org, o.flowName, "trail", o.trailName, "system")
	if err != nil {
		return err
	}

	err = o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	err = o.loadSbom()
	if err != nil {
		return err
	}
	o.attachments = append(o.attachments, o.sbomFilePath)

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer func() {
			if err := os.Remove(evidencePath); err != nil {
				logger.Warn("failed to remove evidence file: %v", err)
			}
		}()
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("sbom attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}

// loadSbom checks the size before reading, because reading happens in one go
// and a large file would otherwise be pulled into memory before the friendlier
// error could be produced.
func (o *attestSbomOptions) loadSbom() error {
	info, err := os.Stat(o.sbomFilePath)
	if err != nil {
		return fmt.Errorf("failed to read SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	// A directory or a pipe reports size 0 and would walk past the size check,
	// then be read unbounded.
	if !info.Mode().IsRegular() {
		return fmt.Errorf("SBOM file [%s] is not a regular file", o.sbomFilePath)
	}
	if info.Size() > maxSbomFileBytes {
		return fmt.Errorf(
			"SBOM file [%s] is %d bytes, above the %d byte limit for an SBOM attestation",
			o.sbomFilePath, info.Size(), maxSbomFileBytes,
		)
	}

	// One read for both, so the recorded fingerprint always describes the bytes
	// the recorded summary was taken from.
	content, err := os.ReadFile(o.sbomFilePath)
	if err != nil {
		return fmt.Errorf("failed to read SBOM file [%s]: %s", o.sbomFilePath, err)
	}
	fingerprint := fmt.Sprintf("%x", sha256.Sum256(content))

	data, err := sbom.ProcessSBOM(content)
	if err != nil {
		return fmt.Errorf("failed to parse SBOM file [%s]: %s", o.sbomFilePath, err)
	}

	o.payload.AttestationData = SbomAttestationData{
		Format:              data.Format,
		OriginalFingerprint: fingerprint,
		Document:            data.Document,
	}
	return o.annotate(data.Format, fingerprint)
}

// rejectReservedAnnotations runs in PreRunE: it needs nothing from the file, so
// a typo should not cost a repository walk and a pass over nine megabytes first.
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

func (o *attestSbomOptions) annotate(format, fingerprint string) error {
	if o.payload.Annotations == nil {
		o.payload.Annotations = map[string]string{}
	}
	o.payload.Annotations[sbomFormatAnnotation] = format
	o.payload.Annotations[sbomSha256Annotation] = fingerprint
	return nil
}
