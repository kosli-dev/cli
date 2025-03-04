package main

import (
	"fmt"
	"io"

	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	"github.com/spf13/cobra"
)

const attestPRBitbucketShortDesc = `Report a Bitbucket pull request attestation to an artifact or a trail in a Kosli flow.  `

const attestPRBitbucketLongDesc = attestPRBitbucketShortDesc + `
It checks if a pull request exists for a given merge commit and reports the pull-request attestation to Kosli.
Authentication to Bitbucket can be done with access token (recommended) or app passwords. Credentials need to have read access for both repos and pull requests.
` + attestationBindingDesc

const attestPRBitbucketExample = `
# report a Bitbucket pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest pullrequest bitbucket yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Bitbucket pull request attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest pullrequest bitbucket \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Bitbucket pull request attestation about a trail:
kosli attest pullrequest bitbucket \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Bitbucket pull request attestation about an artifact which has not been reported yet in a trail:
kosli attest pullrequest bitbucket \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Bitbucket pull request attestation about a trail with an attachment:
kosli attest pullrequest bitbucket \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--attachments=yourAttachmentPathName \
	--api-token yourAPIToken \
	--org yourOrgName

# fail if a pull request does not exist for your artifact
kosli attest pullrequest bitbucket \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert
`

func newAttestBitbucketPRCmd(out io.Writer) *cobra.Command {
	config := new(bbUtils.Config)
	config.Logger = logger
	config.KosliClient = kosliClient

	o := &attestPROptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: PRAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
		retriever: config,
	}
	cmd := &cobra.Command{
		// Args:    cobra.MaximumNArgs(1),  // See CustomMaximumNArgs() below
		Use:         "bitbucket [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases:     []string{"bb"},
		Short:       attestPRBitbucketShortDesc,
		Long:        attestPRBitbucketLongDesc,
		Example:     attestPRBitbucketExample,
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

			err = MuXRequiredFlags(cmd, []string{"fingerprint", "artifact-type"}, false)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"bitbucket-username", "bitbucket-access-token"}, true)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"bitbucket-password", "bitbucket-access-token"}, true)
			if err != nil {
				return err
			}

			err = ConditionallyRequiredFlags(cmd, "bitbucket-username", "bitbucket-password")
			if err != nil {
				return err
			}

			err = ConditionallyRequiredFlags(cmd, "bitbucket-password", "bitbucket-username")
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
	addBitbucketFlags(cmd, o.getRetriever().(*bbUtils.Config), ci)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name",
		"bitbucket-workspace", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
