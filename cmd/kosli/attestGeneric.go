package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type CommonAttestationPayload struct {
	ArtifactFingerprint string                   `json:"artifact_fingerprint,omitempty"`
	Commit              *gitview.BasicCommitInfo `json:"git_commit_info,omitempty"`
	AttestationName     string                   `json:"step_name"`
	TargetArtifacts     []string                 `json:"target_artifacts,omitempty"`
	EvidenceURL         string                   `json:"evidence_url,omitempty"`
	EvidenceFingerprint string                   `json:"evidence_fingerprint,omitempty"`
	Url                 string                   `json:"url,omitempty"`
	UserData            interface{}              `json:"user_data,omitempty"`
}

type GenericAttestationPayload struct {
	CommonAttestationPayload
	Compliant bool `json:"is_compliant"`
}

type attestGenericOptions struct {
	fingerprintOptions      *fingerprintOptions
	attestationNameTemplate string
	flowName                string
	trailName               string
	userDataFilePath        string
	evidencePaths           []string
	commitSHA               string
	srcRepoRoot             string
	payload                 GenericAttestationPayload
}

func parseAttestationNameTemplate(template string) (string, string, error) {
	parts := strings.Split(template, ".")
	if len(parts) == 1 {
		return parts[0], "", nil
	} else if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else {
		return "", "", errors.New("invalid attestation name format")
	}
}

const attestGenericShortDesc = `Report a generic attestation to an artifact or a trail in a Kosli flow.  `

const attestGenericLongDesc = reportEvidenceArtifactGenericShortDesc + `
` + fingerprintDesc

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
	--api-token yourAPIToken \
	--org yourOrgName

# report a generic attestation about a trail with an evidence file:
kosli attest generic \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--evidence-paths=yourEvidencePathName \
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
	o := new(attestGenericOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "generic [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestGenericShortDesc,
		Long:    attestGenericLongDesc,
		Example: attestGenericExample,
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
	// addArtifactEvidenceFlags(cmd, &o.payload.TypedEvidencePayload, ci)
	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "fingerprint", "F", "", attestationFingerprintFlag)
	cmd.Flags().StringVar(&o.commitSHA, "commit", DefaultValue(ci, "git-commit"), attestationCommitFlag)
	cmd.Flags().StringVarP(&o.payload.Url, "url", "b", DefaultValue(ci, "build-url"), attestationUrlFlag)
	cmd.Flags().StringVarP(&o.attestationNameTemplate, "name", "n", "", attestationNameFlag)
	cmd.Flags().StringVar(&o.payload.EvidenceFingerprint, "evidence-fingerprint", "", evidenceFingerprintFlag)
	cmd.Flags().StringVar(&o.payload.EvidenceURL, "evidence-url", "", evidenceURLFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.trailName, "trail", "T", "", trailNameFlag)
	cmd.Flags().BoolVarP(&o.payload.Compliant, "compliant", "C", true, attestationCompliantFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", attestationUserDataFlag)
	cmd.Flags().StringSliceVarP(&o.evidencePaths, "evidence-paths", "e", []string{}, evidencePathsFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", attestationRepoRootFlag)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "trail", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestGenericOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/generic", global.Host, global.Org, o.flowName, o.trailName)

	var err error

	p1, p2, err := parseAttestationNameTemplate(o.attestationNameTemplate)
	if err != nil {
		return err
	}
	if p1 != "" && p2 != "" {
		o.payload.TargetArtifacts = []string{p1}
		o.payload.AttestationName = p2
	} else {
		o.payload.AttestationName = p1
	}

	if o.fingerprintOptions.artifactType != "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	if o.commitSHA != "" {
		gv, err := gitview.New(o.srcRepoRoot)
		if err != nil {
			return err
		}
		commitInfo, err := gv.GetCommitInfoFromCommitSHA(o.commitSHA)
		if err != nil {
			return err
		}
		o.payload.Commit = &commitInfo.BasicCommitInfo
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := newAttestationForm(o.payload, o.evidencePaths)
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	if err != nil {
		return err
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
		logger.Info("generic attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}
	return err
}
