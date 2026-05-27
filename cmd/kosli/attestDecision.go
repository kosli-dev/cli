package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type DecisionAttestationData struct {
	Control   string `json:"control"`
	Compliant bool   `json:"is_compliant"`
}

type DecisionAttestationPayload struct {
	*CommonAttestationPayload
	TypeName        string                  `json:"type_name"`
	AttestationData DecisionAttestationData `json:"attestation_data"`
}

type attestDecisionOptions struct {
	*CommonAttestationOptions
	payload DecisionAttestationPayload
}

const attestDecisionShortDesc = `[BETA] Record a compliance decision against a control in a Kosli trail.  `

const attestDecisionLongDesc = attestDecisionShortDesc + `
Use this command to record the outcome of evaluating a control as part of your delivery
pipeline — whether it was satisfied or not — attached to a specific trail with an optional artifact.
This decision is the evidence that a governance requirement was assessed.
` + attestationBindingDesc + `

` + commitDescription

const attestDecisionExample = `
# record a compliant decision against a trail:
kosli attest decision \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--control RCTL-043 \
	--compliant=true \
	--api-token yourAPIToken \
	--org yourOrgName

# record a non-compliant decision against a trail:
kosli attest decision \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--control RCTL-043 \
	--compliant=false \
	--api-token yourAPIToken \
	--org yourOrgName

# record a decision linked to a specific artifact (by fingerprint):
kosli attest decision \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--control RCTL-043 \
	--compliant=true \
	--fingerprint yourArtifactFingerprint \
	--api-token yourAPIToken \
	--org yourOrgName

# record a decision with an evidence attachment:
kosli attest decision \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--control RCTL-043 \
	--compliant=true \
	--attachments eval-report.json \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestDecisionCmd(out io.Writer) *cobra.Command {
	o := &attestDecisionOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: DecisionAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
			TypeName:                 "decision",
		},
	}
	cmd := &cobra.Command{
		Use:     "decision [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestDecisionShortDesc,
		Long:    attestDecisionLongDesc,
		Example: attestDecisionExample,
		Hidden:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := CustomMaximumNArgs(1, args)
			if err != nil {
				return err
			}

			err = RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			if !cmd.Flags().Changed("compliant") {
				return fmt.Errorf(`required flag(s) "compliant" not set`)
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
	cmd.Flags().StringVar(&o.payload.AttestationData.Control, "control", "", attestationDecisionControlFlag)
	cmd.Flags().BoolVarP(&o.payload.AttestationData.Compliant, "compliant", "C", false, attestationCompliantFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "control"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestDecisionOptions) run(args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/attestations", global.Org, o.flowName, "trail", o.trailName, "system")
	if err != nil {
		return err
	}

	err = o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.attachments)
	if err != nil {
		return err
	}
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
		logger.Info("decision attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return wrapAttestationError(err)
}
