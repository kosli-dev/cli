package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/types"
	"github.com/spf13/cobra"
)

type PRAttestationPayload struct {
	*CommonAttestationPayload
	GitProvider  string              `json:"git_provider"`
	PullRequests []*types.PREvidence `json:"pull_requests"`
}

type attestPROptions struct {
	*CommonAttestationOptions
	retriever interface{}
	assert    bool
	payload   PRAttestationPayload
}

const attestPRGitlabShortDesc = `Report a Gitlab merge request attestation to an artifact or a trail in a Kosli flow.  `

const attestPRGitlabLongDesc = reportEvidenceArtifactGenericShortDesc + `
It checks if a merge request exists for the artifact (based on its git commit) and reports the merge request evidence to the artifact in Kosli.
` + fingerprintDesc

const attestPRGitlabExample = `
# report a Gitlab merge request attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest pullrequest gitlab yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Gitlab merge request attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest pullrequest gitlab \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Gitlab merge request attestation about a trail:
kosli attest pullrequest gitlab \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Gitlab merge request attestation about an artifact which has not been reported yet in a trail:
kosli attest pullrequest gitlab \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Gitlab merge request attestation about a trail with an evidence file:
kosli attest pullrequest gitlab \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--evidence-paths=yourEvidencePathName \
	--api-token yourAPIToken \
	--org yourOrgName

# fail if a merge request does not exist for your artifact
kosli attest pullrequest gitlab \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert
`

func newAttestGitlabPRCmd(out io.Writer) *cobra.Command {
	o := &attestPROptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: PRAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
		retriever: new(gitlabUtils.GitlabConfig),
	}
	cmd := &cobra.Command{
		Use:     "gitlab [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gl"},
		Short:   attestPRGitlabShortDesc,
		Long:    attestPRGitlabLongDesc,
		Example: attestPRGitlabExample,
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
	addGitlabFlags(cmd, o.getRetriever().(*gitlabUtils.GitlabConfig), ci)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name",
		"gitlab-token", "gitlab-org", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestPROptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/attestations/%s/%s/trail/%s/pull_request", global.Host, global.Org, o.flowName, o.trailName)

	err := o.CommonAttestationOptions.run(args, o.payload.CommonAttestationPayload)
	if err != nil {
		return err
	}

	pullRequestsEvidence, err := getPullRequestsEvidence(o.getRetriever(), o.commitSHA, o.assert)
	if err != nil {
		return err
	}

	o.payload.PullRequests = pullRequestsEvidence

	label := ""
	o.payload.GitProvider, label = getGitProviderAndLabel(o.retriever)

	form, cleanupNeeded, evidencePath, err := prepareAttestationForm(o.payload, o.evidencePaths)
	if err != nil {
		return err
	}
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	logger.Debug("found %d %s(s) for commit: %s\n", len(pullRequestsEvidence), label, o.commitSHA)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Form:     form,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("%s %s attestation '%s' is reported to trail: %s", o.payload.GitProvider, label, o.payload.AttestationName, o.trailName)
	}
	return err
}
