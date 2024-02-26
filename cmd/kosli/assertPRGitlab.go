package main

import (
	"fmt"
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

type assertPullRequestGitlabOptions struct {
	gitlabConfig *gitlabUtils.GitlabConfig
	commit       string
}

const assertPRGitlabShortDesc = `Assert a Gitlab merge request for a git commit exists.  `

const assertPRGitlabLongDesc = assertPRGitlabShortDesc + `
The command exits with non-zero exit code 
if no merge requests were found for the commit.`

const assertPRGitlabExample = `
kosli assert mergerequest gitlab \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourGitCommit \
	--repository yourGithubGitRepository
`

func newAssertPullRequestGitlabCmd(out io.Writer) *cobra.Command {
	o := new(assertPullRequestGitlabOptions)
	o.gitlabConfig = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab",
		Aliases: []string{"gl"},
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
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValueForCommit(ci, true), commitPREvidenceFlag)
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
	pullRequestsEvidence, err := o.gitlabConfig.PREvidenceForCommit(o.commit)
	if err != nil {
		return err
	}
	if len(pullRequestsEvidence) == 0 {
		return fmt.Errorf("assert failed: found no merge request(s) in Gitlab for commit: %s", o.commit)
	}
	logger.Info("found [%d] merge request(s) in Gitlab for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}
