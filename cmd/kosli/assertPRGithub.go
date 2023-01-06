package main

import (
	"io"

	"github.com/spf13/cobra"
)

const assertPRGithubShortDesc = `Assert if a Github pull request for a git commit exists.`

const assertPRGithubLongDesc = assertPRGithubShortDesc + `
The command exits with non-zero exit code 
if no pull requests were found for the commit.`

const assertPRGithubExample = `
kosli assert github-pullrequest  \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--commit yourGitCommit \
	--repository yourGithubGitRepository
`

func newAssertPullRequestGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceGithubOptions)
	cmd := &cobra.Command{
		Use:     "github-pullrequest",
		Aliases: []string{"gh-pr", "github-pr"},
		Short:   assertPRGithubShortDesc,
		Long:    assertPRGithubLongDesc,
		Example: assertPRGithubExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.assert = true
			pullRequestsEvidence, err := o.getGithubPullRequests()
			if err != nil {
				return err
			}
			logger.Info("found [%d] pull request(s) in Github for commit: %s", len(pullRequestsEvidence), o.commit)
			return nil
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.ghToken, "github-token", "", githubTokenFlag)
	cmd.Flags().StringVar(&o.ghOwner, "github-org", DefaultValue(ci, "owner"), githubOrgFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"github-token", "github-org", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
