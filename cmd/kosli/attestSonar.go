package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/sonar"
	"github.com/spf13/cobra"
)

type SonarAttestationPayload struct {
	*CommonAttestationPayload
	SonarResults *sonar.SonarResults `json:"sonar_results"`
}

type attestSonarOptions struct {
	*CommonAttestationOptions
	projectKey    string
	apiToken      string
	sonarQubeUrl  string
	branchName    string
	pullRequestID string
	payload       SonarAttestationPayload
}

const attestSonarShortDesc = `Report a sonarcloud or sonarqube attestation to an artifact or a trail in a Kosli flow.  `

const attestSonarLongDesc = attestSonarShortDesc + `
Retrieves the latest scan results for the given project key from SonarCloud or SonarQube (if a SonarQube server URL is given).
The results are parsed to find the status of the project's quality gate which is used to determine the attestation's compliance status.
` + attestationBindingDesc

const attestSonarExample = `
# report a sonarcloud attestation about a trail:
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName \

# report a sonarqube attestation about a trail:
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--sonarqube-url yourSonarQubeURL \
	--api-token yourAPIToken \
	--org yourOrgName \

# report a sonarcloud attestation for a specific branch about a trail:
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--branch-name yourBranchName \
	--api-token yourAPIToken \
	--org yourOrgName \

# report a sonarqube attestation for a pull-request about a trail:
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--sonarqube-url yourSonarQubeURL \
	--pull-request-id yourPullRequestID \
	--api-token yourAPIToken \
	--org yourOrgName \

# report a sonarcloud attestation about a trail with an attachment:
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--attachment yourAttachmentPath \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttestSonarCmd(out io.Writer) *cobra.Command {
	o := &attestSonarOptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: SonarAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}

	cmd := &cobra.Command{
		Use:     "sonar [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   attestSonarShortDesc,
		Long:    attestSonarLongDesc,
		Example: attestSonarExample,
		Args:    cobra.MaximumNArgs(1),
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
	cmd.Flags().StringVar(&o.projectKey, "sonar-project-key", "", sonarProjectKeyFlag)
	cmd.Flags().StringVar(&o.apiToken, "sonar-api-token", "", sonarAPITokenFlag)
	cmd.Flags().StringVar(&o.sonarQubeUrl, "sonarqube-url", "", sonarQubeUrlFlag)
	cmd.Flags().StringVar(&o.branchName, "branch-name", "", sonarBranchNameFlag)
	cmd.Flags().StringVar(&o.pullRequestID, "pull-request-id", "", sonarPullRequestFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name", "sonar-project-key", "sonar-api-token"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestSonarOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/sonar", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	sc := sonar.NewSonarConfig(o.projectKey, o.apiToken, o.sonarQubeUrl, o.branchName, o.pullRequestID)

	o.payload.SonarResults, err = sc.GetSonarResults()
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
		Method:   http.MethodPost,
		URL:      url,
		Form:     form,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("sonar attestation '%s' is reported to trail: %s", o.payload.AttestationName, o.trailName)
	}

	return wrapAttestationError(err)

}
