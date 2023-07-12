package main

import (
	"io"

	azUtils "github.com/kosli-dev/cli/internal/azure"
	"github.com/spf13/cobra"
)

const reportEvidenceArtifactPRAzureShortDesc = `Report an Azure Devops pull request evidence for an artifact in a Kosli flow.  `

const reportEvidenceArtifactPRAzureLongDesc = reportEvidenceArtifactPRAzureShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.  
` + fingerprintDesc

const reportEvidenceArtifactPRAzureExample = `
# report a pull request evidence to kosli for a docker image
kosli report evidence artifact pullrequest azure yourDockerImageName \
	--artifact-type docker \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--azure-token yourAzureToken \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli report evidence artifact pullrequest azure yourDockerImageName \
	--artifact-type docker \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--azure-token yourAzureToken \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--assert
`

func newReportEvidenceArtifactPRAzureCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	azureFlagsValues := new(azUtils.AzureFlagsTempValueHolder)
	cmd := &cobra.Command{
		Use:     "azure [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"az"},
		Short:   reportEvidenceArtifactPRAzureShortDesc,
		Long:    reportEvidenceArtifactPRAzureLongDesc,
		Example: reportEvidenceArtifactPRAzureExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.retriever = azUtils.NewAzureConfig(azureFlagsValues.Token,
				azureFlagsValues.OrgUrl, azureFlagsValues.Project, azureFlagsValues.Repository)
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	addAzureFlags(cmd, azureFlagsValues, ci)
	addArtifactPRFlags(cmd, o, ci)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"azure-token", "azure-org-url", "commit",
		"repository", "project", "flow", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
