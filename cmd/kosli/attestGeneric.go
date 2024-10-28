package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type GenericAttestationPayload struct {
	*CommonAttestationPayload
	Compliant bool `json:"is_compliant"`
}

type attestGenericOptions struct {
	*CommonAttestationOptions
	payload GenericAttestationPayload
}

const attestGenericShortDesc = `Report a generic attestation to an artifact or a trail in a Kosli flow.  `

const attestGenericLongDesc = attestGenericShortDesc + attestationBindingDesc + `

` + commitDescription

const attestGenericExample = `
# report a generic attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest generic yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a generic attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest generic \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a generic attestation about a trail:
kosli attest generic \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a generic attestation about an artifact which has not been reported yet in a trail:
kosli attest generic \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--commit yourArtifactGitCommit \
	--api-token yourAPIToken \
	--org yourOrgName

# report a generic attestation about a trail with an attachment:
kosli attest generic \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--attachments yourAttachmentPathName \
	--api-token yourAPIToken \
	--org yourOrgName

# report a non-compliant generic attestation about a trail:
kosli attest generic \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--compliant=false \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestGenericCmd(out io.Writer) *cobra.Command {
	o := &attestGenericOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: GenericAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:     "generic [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestGenericShortDesc,
		Long:    attestGenericLongDesc,
		Example: attestGenericExample,
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
	cmd.Flags().BoolVarP(&o.payload.Compliant, "compliant", "C", true, attestationCompliantFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestGenericOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/generic", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
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
		logger.Info("generic attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}
