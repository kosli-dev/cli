package main

import (
	"io"

	azUtils "github.com/kosli-dev/cli/internal/azure"
	"github.com/spf13/cobra"
)

type assertPullRequestAzureOptions struct {
	azureConfig *azUtils.AzureConfig
	commit      string
}

const assertPRAzureShortDesc = `Assert a Azure DevOps pull request for a git commit exists.  `

const assertPRAzureLongDesc = assertPRAzureShortDesc + `
The command exits with non-zero exit code 
if no pull requests were found for the commit.`

const assertPRAzureExample = `
kosli assert pullrequest azure \
	--azure-token yourAzureToken \
	--azure-org-url yourAzureOrgUrl \
	--commit yourGitCommit \
	--project yourAzureDevopsProject \
	--repository yourAzureDevOpsGitRepository
`

func newAssertPullRequestAzureCmd(out io.Writer) *cobra.Command {
	o := new(assertPullRequestAzureOptions)
	azureFlagsValues := new(azUtils.AzureFlagsTempValueHolder)
	cmd := &cobra.Command{
		Use:     "azure",
		Aliases: []string{"az"},
		Short:   assertPRAzureShortDesc,
		Long:    assertPRAzureLongDesc,
		Example: assertPRAzureExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.azureConfig = azUtils.NewAzureConfig(azureFlagsValues.Token,
				azureFlagsValues.OrgUrl, azureFlagsValues.Project, azureFlagsValues.Repository)
			return o.run(args)
		},
	}

	ci := WhichCI()
	addAzureFlags(cmd, azureFlagsValues, ci)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"azure-token", "azure-org-url", "project", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertPullRequestAzureOptions) run(args []string) error {
	pullRequestsEvidence, err := getPullRequestsEvidence(o.azureConfig, o.commit, true)
	if err != nil {
		return err
	}
	logger.Info("found [%d] pull request(s) in Azure DevOps for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}
