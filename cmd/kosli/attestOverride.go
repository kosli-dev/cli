package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type OverrideAttestationPayload struct {
	*CommonAttestationPayload
	Reason                  string `json:"reason"`
	NewComplianceStatus     bool   `json:"new_compliance_status"`
	OriginalAttestationType string `json:"original_attestation_type"`
}

type attestOverrideOptions struct {
	*CommonAttestationOptions
	payload OverrideAttestationPayload
}

const attestOverrideShortDesc = `Override an attestation in a trail.  `

const attestOverrideLongDesc = attestOverrideShortDesc + `
Use this command to record a manual override of a previously reported attestation. The override sets a new
compliance status and captures the reason as part of the audit trail.
` + attestationBindingDesc + `

` + commitDescription

const attestOverrideExample = `
# override an attestation against a trail to non-compliant:
kosli attest override \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--reason "manual review failed" \
	--new-compliance-status=false \
	--api-token yourAPIToken \
	--org yourOrgName

# override an attestation against an artifact (by fingerprint):
kosli attest override \
	--fingerprint yourArtifactFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--reason "approved out-of-band" \
	--original-attestation-type generic \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestOverrideCmd(out io.Writer) *cobra.Command {
	o := &attestOverrideOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: OverrideAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	cmd := &cobra.Command{
		Use:     "override [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestOverrideShortDesc,
		Long:    attestOverrideLongDesc,
		Example: attestOverrideExample,
		Hidden:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := CustomMaximumNArgs(1, args)
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("new-compliance-status") {
				return ErrorBeforePrintingUsage(cmd, "required flag(s) \"new-compliance-status\" not set")
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
	// the override endpoint takes a JSON body and does not accept attachments
	if err := cmd.Flags().MarkHidden("attachments"); err != nil {
		logger.Error("failed to hide --attachments flag: %v", err)
	}
	cmd.Flags().StringVar(&o.payload.Reason, "reason", "", attestationOverrideReasonFlag)
	cmd.Flags().BoolVar(&o.payload.NewComplianceStatus, "new-compliance-status", false, newComplianceStatusFlag)
	cmd.Flags().StringVar(&o.payload.OriginalAttestationType, "original-attestation-type", "", originalAttestationTypeFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "reason", "original-attestation-type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestOverrideOptions) run(args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/attestations", global.Org, o.flowName, "trail", o.trailName, "override")
	if err != nil {
		return err
	}

	err = o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPost,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("attestation '%s' has been overridden in trail: %s", o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}
