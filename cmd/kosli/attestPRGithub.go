package main

import (
	"fmt"
	"io"

	ghUtils "github.com/kosli-dev/cli/internal/github"
	"github.com/spf13/cobra"
)

const attestPRGithubShortDesc = `Report a Github pull request attestation to an artifact or a trail in a Kosli flow.  `

const attestPRGithubLongDesc = attestPRGithubShortDesc + `
It checks if a pull request exists for a given merge commit and reports the pull-request attestation to Kosli.
` + attestationBindingDesc

const attestPRGithubExample = `
# report a Github pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest pullrequest github yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest pullrequest github \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about a trail:
kosli attest pullrequest github \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about an artifact which has not been reported yet in a trail:
kosli attest pullrequest github \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about a trail with an attachment:
kosli attest pullrequest github \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--attachments=yourAttachmentPathName \
	--api-token yourAPIToken \
	--org yourOrgName

# fail if a pull request does not exist for your artifact
kosli attest pullrequest github \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert
`

func newAttestGithubPRCmd(out io.Writer) *cobra.Command {
	o := &attestPROptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: PRAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	githubFlagsValues := new(ghUtils.GithubFlagsTempValueHolder)
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:     "github [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gh"},
		Short:   attestPRGithubShortDesc,
		Long:    attestPRGithubLongDesc,
		Example: attestPRGithubExample,
		Annotations: map[string]string{"pr": "true"},
		PreRunE: func(cmd *cobra.Command, args []string) error {

			err := CustomMaximumNArgs(1, args)
			if err != nil {
				return err
			}

			err = RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateSliceValues(o.redactedCommitInfo, allowedCommitRedactionValues)
			if err != nil {
				return fmt.Errorf("%s for --redact-commit-info", err.Error())
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
			o.retriever = ghUtils.NewGithubConfig(githubFlagsValues.Token, githubFlagsValues.BaseURL,
				githubFlagsValues.Org, githubFlagsValues.Repository)
			return o.run(args)
		},
	}

	ci := WhichCI()
	addAttestationFlags(cmd, o.CommonAttestationOptions, o.payload.CommonAttestationPayload, ci)
	addGithubFlags(cmd, githubFlagsValues, ci)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name",
		"github-token", "github-org", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
