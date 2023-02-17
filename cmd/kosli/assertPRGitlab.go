package main

import (
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

type assertPullRequestGitlabOptions struct {
	gitlabConfig *gitlabUtils.GitlabConfig
	commit       string
}

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
	o := new(assertPullRequestGitlabOptions)
	o.gitlabConfig = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab-mergerequest",
		Short:   assertPRGitlabShortDesc,
		Long:    assertPRGitlabLongDesc,
		Example: assertPRGitlabExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	addGitlabFlags(cmd, o.gitlabConfig, ci)
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

func (o *assertPullRequestGitlabOptions) run(args []string) error {
	pullRequestsEvidence, err := getPullRequestsEvidence(o.gitlabConfig, o.commit, true)
	if err != nil {
		return err
	}
	logger.Info("found [%d] pull request(s) in Gitlab for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}
