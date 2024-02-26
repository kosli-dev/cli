package main

import (
	"fmt"
	"io"

	ghUtils "github.com/kosli-dev/cli/internal/github"
	"github.com/spf13/cobra"
)

type assertPullRequestGithubOptions struct {
	githubConfig *ghUtils.GithubConfig
	commit       string
}

const assertPRGithubShortDesc = `Assert a Github pull request for a git commit exists.  `

const assertPRGithubLongDesc = assertPRGithubShortDesc + `
The command exits with non-zero exit code 
if no pull requests were found for the commit.`

const assertPRGithubExample = `
kosli assert pullrequest github \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourGitCommit \
	--repository yourGithubGitRepository
`

func newAssertPullRequestGithubCmd(out io.Writer) *cobra.Command {
	o := new(assertPullRequestGithubOptions)
	githubFlagsValues := new(ghUtils.GithubFlagsTempValueHolder)
	cmd := &cobra.Command{
		Use:     "github",
		Aliases: []string{"gh"},
		Short:   assertPRGithubShortDesc,
		Long:    assertPRGithubLongDesc,
		Example: assertPRGithubExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.githubConfig = ghUtils.NewGithubConfig(githubFlagsValues.Token, githubFlagsValues.BaseURL,
				githubFlagsValues.Org, githubFlagsValues.Repository)
			return o.run(args)
		},
	}

	ci := WhichCI()
	addGithubFlags(cmd, githubFlagsValues, ci)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValueForCommit(ci, true), commitPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"github-token", "github-org", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertPullRequestGithubOptions) run(args []string) error {
	pullRequestsEvidence, err := o.githubConfig.PREvidenceForCommit(o.commit)
	if err != nil {
		return err
	}
	if len(pullRequestsEvidence) == 0 {
		return fmt.Errorf("assert failed: found no pull request(s) in Github for commit: %s", o.commit)
	}
	logger.Info("found [%d] pull request(s) in Github for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}
