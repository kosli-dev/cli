package main

import (
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

const assertPRGitlabShortDesc = `Assert if a Gitlab pull request for a git commit exists.`

const assertPRGitlabLongDesc = assertPRGitlabShortDesc + `
The command exits with non-zero exit code 
if no pull requests were found for the commit.`

const assertPRGitlabExample = `
kosli assert gitlab-mergerequest \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--commit yourGitCommit \
	--repository yourGithubGitRepository
`

func newAssertPullRequestGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceGitlabOptions)
	o.gitlabConfig = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab-mergerequest",
		Short:   assertPRGitlabShortDesc,
		Long:    assertPRGitlabLongDesc,
		Example: assertPRGitlabExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.assert = true
			pullRequestsEvidence, err := getGitlabPullRequests(o.gitlabConfig, o.commit, true)
			if err != nil {
				return err
			}
			logger.Info("found [%d] pull request(s) in Gitlab for commit: %s", len(pullRequestsEvidence), o.commit)
			return nil
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.gitlabConfig.Token, "gitlab-token", "", gitlabTokenFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.Org, "gitlab-org", DefaultValue(ci, "namespace"), gitlabOrgFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"gitlab-token", "gitlab-org", "commit", "repository",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
