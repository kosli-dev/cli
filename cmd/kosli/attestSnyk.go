package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type SnykAttestationPayload struct {
	*CommonAttestationPayload
	SnykResults interface{} `json:"snyk_results"`
}

type attestSnykOptions struct {
	*CommonAttestationOptions
	snykJsonFilePath string
	payload          SnykAttestationPayload
}

const attestSnykShortDesc = `Report a snyk attestation to an artifact or a trail in a Kosli flow.  `

const attestSnykLongDesc = attestSnykShortDesc + `
` + fingerprintDesc

const attestSnykExample = `
# report a snyk attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest snyk yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykJSONScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest snyk \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykJSONScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about a trail:
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykJSONScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about an artifact which has not been reported yet in a trail:
kosli attest snyk \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykJSONScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

# report a snyk attestation about a trail with an evidence file:
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykJSONScanResults \
	--evidence-paths=yourEvidencePathName \
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
		Use:     "snyk [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestSnykShortDesc,
		Long:    attestSnykLongDesc,
		Example: attestSnykExample,
		Args:    cobra.MaximumNArgs(1),
		Hidden:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"fingerprint", "artifact-type"}, false)
			if err != nil {
				return err
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
	cmd.Flags().StringVarP(&o.snykJsonFilePath, "scan-results", "R", "", snykJsonResultsFileFlag)

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

	o.payload.SnykResults, err = LoadJsonData(o.snykJsonFilePath)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.evidencePaths)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Form:     form,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("snyk attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return err
}
