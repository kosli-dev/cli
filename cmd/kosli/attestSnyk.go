package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/snyk"
	"github.com/spf13/cobra"
)

type SnykAttestationPayload struct {
	*CommonAttestationPayload
	SnykResults *snyk.SnykData `json:"snyk_results"`
}

type attestSnykOptions struct {
	*CommonAttestationOptions
	snykSarifFilePath string
	uploadResultsFile bool
	payload           SnykAttestationPayload
}

const attestSnykShortDesc = `Report a snyk attestation to an artifact or a trail in a Kosli flow.  `

const attestSnykLongDesc = attestSnykShortDesc + `
Only SARIF snyk output is accepted. 
Snyk output can be for "snyk code test", "snyk container test", or "snyk iac test".

The ^--scan-results^ .json file is analyzed and a summary of the scan results are reported to Kosli.

By default, the ^--scan-results^ .json file is also uploaded to Kosli's evidence vault.
You can disable that by setting ^--upload-results=false^
` + attestationBindingDesc + `

` + commitDescription

const attestSnykExample = `
# report a snyk attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest snyk yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest snyk \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about a trail:
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about an artifact which has not been reported yet in a trail:
kosli attest snyk \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--commit yourArtifactGitCommit \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about a trail with an attachment:
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--attachments yourEvidencePathName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about a trail without uploading the snyk results file:
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--upload-results=false \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestSnykCmd(out io.Writer) *cobra.Command {
	o := &attestSnykOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: SnykAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:     "snyk [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestSnykShortDesc,
		Long:    attestSnykLongDesc,
		Example: attestSnykExample,
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
			return o.run(args)
		},
	}

	ci := WhichCI()
	addAttestationFlags(cmd, o.CommonAttestationOptions, o.payload.CommonAttestationPayload, ci)
	cmd.Flags().StringVarP(&o.snykSarifFilePath, "scan-results", "R", "", snykSarifResultsFileFlag)
	cmd.Flags().BoolVar(&o.uploadResultsFile, "upload-results", true, uploadSnykResultsFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "scan-results"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestSnykOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/snyk", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	o.payload.SnykResults, err = snyk.ProcessSnykResultFile(o.snykSarifFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Snyk sarif results file [%s]: %s", o.snykSarifFilePath, err)
	}

	if o.uploadResultsFile {
		o.attachments = append(o.attachments, o.snykSarifFilePath)
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
		logger.Info("snyk attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}
