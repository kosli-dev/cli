package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type CustomAttestationPayload struct {
	*CommonAttestationPayload
	TypeName        string      `json:"type_name"`
	AttestationData interface{} `json:"attestation_data"`
}

type attestCustomOptions struct {
	*CommonAttestationOptions
	attestationDataFile string
	payload             CustomAttestationPayload
}

const attestCustomShortDesc = `Report a custom attestation to an artifact or a trail in a Kosli flow. `

const attestCustomLongDesc = attestCustomShortDesc + `
The name of the custom attestation type is specified using the ^--type^ flag.
` + attestationBindingDesc + `

` + commitDescription

const attestCustomExample = `
# report a custom attestation about a pre-built container image artifact (kosli finds the fingerprint):
kosli attest custom yourDockerImageName \
	--artifact-type oci \
	--type customTypeName \
	--name yourAttestationName \
	--attestation-data yourJsonFilePath \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a custom attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest custom \
	--fingerprint yourDockerImageFingerprint \
	--type customTypeName \
	--name yourAttestationName \
	--attestation-data yourJsonFilePath \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a custom attestation about a trail:
kosli attest custom \
	--type customTypeName \
	--name yourAttestationName \
	--attestation-data yourJsonFilePath \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a custom attestation about an artifact which has not been reported yet in a trail:
kosli attest custom \
	--type customTypeName \
	--name yourTemplateArtifactName.yourAttestationName \
	--attestation-data yourJsonFilePath \
	--flow yourFlowName \
	--trail yourTrailName \
	--commit yourArtifactGitCommit \
	--api-token yourAPIToken \
	--org yourOrgName

# report a custom attestation about a trail with an attachment:
kosli attest custom \
    --type customTypeName \
	--name yourAttestationName \
	--attestation-data yourJsonFilePath \
	--flow yourFlowName \
	--trail yourTrailName \
	--attachments yourAttachmentPathName \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestCustomCmd(out io.Writer) *cobra.Command {
	o := &attestCustomOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: CustomAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:     "custom [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestCustomShortDesc,
		Long:    attestCustomLongDesc,
		Example: attestCustomExample,
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
	cmd.Flags().StringVar(&o.payload.TypeName, "type", "", attestationCustomTypeNameFlag)
	cmd.Flags().StringVar(&o.attestationDataFile, "attestation-data", "", attestationCustomDataFileFlag)

	err := RequireFlags(cmd, []string{"type", "attestation-data", "flow", "trail", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestCustomOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/custom", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	o.payload.AttestationData, err = LoadJsonData(o.attestationDataFile)
	if err != nil {
		return fmt.Errorf("failed to load attestation data. %s", err)
	}

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
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
		logger.Info("custom:%s attestation '%s' is reported to trail: %s", o.payload.TypeName, o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}
