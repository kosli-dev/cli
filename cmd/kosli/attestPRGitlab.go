package main

import (
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

const attestPRGitlabShortDesc = `Report a Gitlab merge request attestation to an artifact or a trail in a Kosli flow.  `

const attestPRGitlabLongDesc = attestPRGitlabShortDesc + `
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
