package main

import (
	"io"

	ghUtils "github.com/kosli-dev/cli/internal/github"

	"github.com/spf13/cobra"
)

const pullRequestCommitEvidenceGithubShortDesc = `Report a Github pull request evidence for a git commit in a Kosli pipeline.`

const pullRequestCommitEvidenceGithubLongDesc = pullRequestCommitEvidenceGithubShortDesc + `
It checks if a pull request exists for a commit and report the pull-request evidence to the commit in Kosli. 
`

const pullRequestCommitEvidenceGithubExample = `
# report a pull request commit evidence to Kosli
kosli commit report evidence github-pullrequest \
	--commit yourGitCommitSha1 \
	--repository yourGithubGitRepository \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--name yourEvidenceName \
	--pipelines yourPipelineName1,yourPipelineName2 \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli commit report evidence github-pullrequest \
	--commit yourGitCommitSha1 \
	--repository yourGithubGitRepository \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--name yourEvidenceName \
	--pipelines yourPipelineName1,yourPipelineName2 \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

// TODO: do we need to support assert for this command? see line 74

func newPullRequestCommitEvidenceGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitOptions)
	o.retriever = new(ghUtils.GithubConfig)
	cmd := &cobra.Command{
		Use:     "github-pullrequest",
		Aliases: []string{"gh-pr", "github-pr"},
		Short:   pullRequestCommitEvidenceGithubShortDesc,
		Long:    pullRequestCommitEvidenceGithubLongDesc,
		Example: pullRequestCommitEvidenceGithubExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()

	addGithubFlags(cmd, o.retriever.(*ghUtils.GithubConfig), ci)
	addCommitPRFlags(cmd, o, ci)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"github-token", "github-org", "commit",
		"repository", "pipelines", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
