package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/sarif"
	"github.com/spf13/cobra"
)

type SarifAttestationPayload struct {
	*CommonAttestationPayload
	SarifResults *sarif.SarifData `json:"sarif_results"`
	Compliant    bool             `json:"is_compliant"`
}

type attestSarifOptions struct {
	*CommonAttestationOptions
	sarifFilePath     string
	uploadResultsFile bool
	payload           SarifAttestationPayload
}

const attestSarifShortDesc = `Report a SARIF attestation to an artifact or a trail in a Kosli flow.  `

const attestSarifLongDesc = attestSarifShortDesc + `
Accepts SARIF v2.1.0 scan results from any compatible tool (e.g. Checkov, Trivy, Semgrep, Snyk, CodeQL).
The tool name and version are taken from the SARIF report's runs[0].tool.driver fields and shown in
the Kosli UI alongside the parsed findings.

The ^--scan-results^ .json file is analyzed and a summary of the scan results is reported to Kosli.

By default, the ^--scan-results^ .json file is also uploaded to Kosli's evidence vault.
You can disable that by setting ^--upload-results=false^.

Compliance is determined by the ^--compliant^ flag (default true). The CLI does not derive
compliance from the SARIF findings — the caller decides whether the scan should be treated
as compliant or not (e.g. based on its own policy or rego rules).
` + attestationBindingDesc + `

` + commitDescription

const attestSarifExample = `
# report a SARIF attestation about a trail (compliant by default):
kosli attest sarif \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourScanSARIFResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a non-compliant SARIF attestation about a trail:
kosli attest sarif \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourScanSARIFResults \
	--compliant=false \
	--api-token yourAPIToken \
	--org yourOrgName

# report a SARIF attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest sarif yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourScanSARIFResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a SARIF attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest sarif \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourScanSARIFResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a SARIF attestation about an artifact which has not been reported yet in a trail:
kosli attest sarif \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--commit yourArtifactGitCommit \
	--scan-results yourScanSARIFResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a SARIF attestation about a trail without uploading the results file:
kosli attest sarif \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourScanSARIFResults \
	--upload-results=false \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestSarifCmd(out io.Writer) *cobra.Command {
	o := &attestSarifOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: SarifAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	cmd := &cobra.Command{
		Use:     "sarif [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestSarifShortDesc,
		Long:    attestSarifLongDesc,
		Example: attestSarifExample,
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
			return o.run(args)
		},
	}

	ci := WhichCI()
	addAttestationFlags(cmd, o.CommonAttestationOptions, o.payload.CommonAttestationPayload, ci)
	cmd.Flags().StringVarP(&o.sarifFilePath, "scan-results", "R", "", sarifResultsFileFlag)
	cmd.Flags().BoolVar(&o.uploadResultsFile, "upload-results", true, uploadSarifResultsFlag)
	cmd.Flags().BoolVarP(&o.payload.Compliant, "compliant", "C", true, attestationCompliantFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "scan-results"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestSarifOptions) run(args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/attestations", global.Org, o.flowName, "trail", o.trailName, "sarif")
	if err != nil {
		return err
	}

	err = o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	o.payload.SarifResults, err = sarif.ProcessSarifResultFile(o.sarifFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse SARIF results file [%s]: %s", o.sarifFilePath, err)
	}

	if o.uploadResultsFile {
		o.attachments = append(o.attachments, o.sarifFilePath)
	}

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer func() {
			if err := os.Remove(evidencePath); err != nil {
				logger.Warn("failed to remove evidence file %s: %v", evidencePath, err)
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
		logger.Info("sarif attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}
