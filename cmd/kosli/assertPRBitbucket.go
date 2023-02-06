package main

import (
	"io"

	"github.com/spf13/cobra"
)

type assertPullRequestBitbucketOptions struct {
	bbUsername  string
	bbPassword  string
	bbWorkspace string
	commit      string
	repository  string
}

const assertPRBitbucketShortDesc = `Assert if a Bitbucket pull request for a git commit exists. `

const assertPRBitbucketLongDesc = assertPRBitbucketShortDesc + `
The command exits with non-zero exit code 
if no pull requests were found for the commit.`

const assertPRBitbucketExample = `
kosli assert bitbucket-pullrequest  \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourGitCommit \
	--repository yourBitbucketGitRepository
`

func newAssertPullRequestBitbucketCmd(out io.Writer) *cobra.Command {
	o := new(assertPullRequestBitbucketOptions)
	cmd := &cobra.Command{
		Use:     "bitbucket-pullrequest",
		Aliases: []string{"bb-pr", "bitbucket-pr"},
		Short:   assertPRBitbucketShortDesc,
		Long:    assertPRBitbucketLongDesc,
		Example: assertPRBitbucketExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.bbUsername, "bitbucket-username", "", bbUsernameFlag)
	cmd.Flags().StringVar(&o.bbPassword, "bitbucket-password", "", bbPasswordFlag)
	cmd.Flags().StringVar(&o.bbWorkspace, "bitbucket-workspace", DefaultValue(ci, "workspace"), bbWorkspaceFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"bitbucket-username", "bitbucket-password",
		"bitbucket-workspace", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertPullRequestBitbucketOptions) run(args []string) error {
	pullRequestsEvidence, err := getPullRequestsFromBitbucketApi(o.bbWorkspace,
		o.repository, o.commit, o.bbUsername, o.bbPassword, true)
	if err != nil {
		return err
	}
	logger.Info("found [%d] pull request(s) in Bitbucket for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}
