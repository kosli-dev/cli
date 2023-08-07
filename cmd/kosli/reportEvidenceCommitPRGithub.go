package main

import (
	"io"

	ghUtils "github.com/kosli-dev/cli/internal/github"

	"github.com/spf13/cobra"
)

const reportEvidenceCommitPRGithubShortDesc = `Report Github pull request evidence for a git commit in Kosli flows.  `

const reportEvidenceCommitPRGithubLongDesc = reportEvidenceCommitPRGithubShortDesc + `
It checks if a pull request exists for a commit and report the pull-request evidence to the commit in Kosli. 
`

const reportEvidenceCommitPRGithubExample = `
# report a pull request commit evidence to Kosli
kosli report evidence commit pullrequest github \
	--commit yourGitCommitSha1 \
	--repository yourGithubGitRepository \
	--github-token yourGithubToken \
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

func newReportEvidenceCommitPRGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitOptions)
	githubFlagsValues := new(ghUtils.GithubFlagsTempValueHolder)
	cmd := &cobra.Command{
		Use:     "github",
		Aliases: []string{"gh"},
		Short:   reportEvidenceCommitPRGithubShortDesc,
		Long:    reportEvidenceCommitPRGithubLongDesc,
		Example: reportEvidenceCommitPRGithubExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.retriever = ghUtils.NewGithubConfig(githubFlagsValues.Token, githubFlagsValues.BaseURL,
				githubFlagsValues.Org, githubFlagsValues.Repository)
			return o.run(args)
		},
	}

	ci := WhichCI()

	addGithubFlags(cmd, githubFlagsValues, ci)
	addCommitPRFlags(cmd, o, ci)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"github-token", "github-org", "commit",
		"repository", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
