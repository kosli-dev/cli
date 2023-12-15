package main

import (
	"io"

	azUtils "github.com/kosli-dev/cli/internal/azure"
	"github.com/spf13/cobra"
)

const attestPRAzureShortDesc = `Report an Azure Devops pull request attestation to an artifact or a trail in a Kosli flow.  `

const attestPRAzureLongDesc = attestPRAzureShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.
` + fingerprintDesc

const attestPRAzureExample = `
# report an Azure Devops pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest pullrequest azure yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest pullrequest azure \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about a trail:
kosli attest pullrequest azure \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about an artifact which has not been reported yet in a trail:
kosli attest pullrequest azure \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about a trail with an evidence file:
kosli attest pullrequest azure \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--evidence-paths=yourEvidencePathName \
	--api-token yourAPIToken \
	--org yourOrgName

# fail if a pull request does not exist for your artifact
kosli attest pullrequest azure \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert
`

func newAttestAzurePRCmd(out io.Writer) *cobra.Command {
	o := &attestPROptions{
		CommonAttestationOptions: &CommonAttestationOptions{
			fingerprintOptions: &fingerprintOptions{},
		},
		payload: PRAttestationPayload{
			CommonAttestationPayload: &CommonAttestationPayload{},
		},
	}
	azureFlagsValues := new(azUtils.AzureFlagsTempValueHolder)
	cmd := &cobra.Command{
		Use:     "azure [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"az"},
		Short:   attestPRAzureShortDesc,
		Long:    attestPRAzureLongDesc,
		Example: attestPRAzureExample,
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
			o.retriever = azUtils.NewAzureConfig(azureFlagsValues.Token,
				azureFlagsValues.OrgUrl, azureFlagsValues.Project, azureFlagsValues.Repository)
			return o.run(args)
		},
	}

	ci := WhichCI()
	addAttestationFlags(cmd, o.CommonAttestationOptions, o.payload.CommonAttestationPayload, ci)
	addAzureFlags(cmd, azureFlagsValues, ci)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "name",
		"azure-token", "azure-org-url",
		"project", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
