package main

import (
	"io"

	azUtils "github.com/kosli-dev/cli/internal/azure"
	"github.com/spf13/cobra"
)

const reportEvidenceCommitPRAzureShortDesc = `Report an Azure Devops pull request evidence for a git commit in Kosli flows.`

const reportEvidenceCommitPRAzureLongDesc = reportEvidenceCommitPRAzureShortDesc + `
It checks if a pull request exists for a commit and report the pull-request evidence to the commit in Kosli. 
`

const reportEvidenceCommitPRAzureExample = `
# report a pull request commit evidence to Kosli
kosli report evidence commit pullrequest azure \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--azure-token yourAzureToken \
	--github-org yourGithubOrg \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--org yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli report evidence commit pullrequest github \
	--commit yourGitCommitSha1 \
	--repository yourGithubGitRepository \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--org yourOrgName \
	--api-token yourAPIToken \
	--assert
`

// TODO: do we need to support assert for this command? see line 74

func newReportEvidenceCommitPRAzureCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitOptions)
	azureFlagsValues := new(azUtils.AzureFlagsTempValueHolder)
	cmd := &cobra.Command{
		Use:     "azure",
		Aliases: []string{"az"},
		Short:   reportEvidenceCommitPRAzureShortDesc,
		Long:    reportEvidenceCommitPRAzureLongDesc,
		Example: reportEvidenceCommitPRAzureExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.retriever = azUtils.NewAzureConfig(azureFlagsValues.Token, azureFlagsValues.BaseURL,
				azureFlagsValues.Org, azureFlagsValues.Project, azureFlagsValues.Repository)
			return o.run(args)
		},
	}

	ci := WhichCI()

	addAzureFlags(cmd, azureFlagsValues, ci)
	addCommitPRFlags(cmd, o, ci)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"azure-token", "azure-org", "commit",
		"repository", "project", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
