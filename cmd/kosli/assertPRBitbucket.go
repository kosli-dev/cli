package main

import (
	"io"

	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	"github.com/spf13/cobra"
)

type assertPullRequestBitbucketOptions struct {
	bbConfig *bbUtils.Config
	commit   string
}

const assertPRBitbucketShortDesc = `Assert a Bitbucket pull request for a git commit exists.  `

const assertPRBitbucketLongDesc = assertPRBitbucketShortDesc + `
The command exits with non-zero exit code 
if no pull requests were found for the commit.`

const assertPRBitbucketExample = `
kosli assert pullrequest bitbucket  \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourGitCommit \
	--repository yourBitbucketGitRepository
`

func newAssertPullRequestBitbucketCmd(out io.Writer) *cobra.Command {
	o := new(assertPullRequestBitbucketOptions)
	o.bbConfig = new(bbUtils.Config)
	o.bbConfig.Logger = logger
	o.bbConfig.KosliClient = kosliClient
	cmd := &cobra.Command{
		Use:     "bitbucket",
		Aliases: []string{"bb"},
		Short:   assertPRBitbucketShortDesc,
		Long:    assertPRBitbucketLongDesc,
		Example: assertPRBitbucketExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	addBitbucketFlags(cmd, o.bbConfig, ci)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"bitbucket-username", "bitbucket-password",
		"bitbucket-workspace", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertPullRequestBitbucketOptions) run(args []string) error {
	pullRequestsEvidence, err := getPullRequestsEvidence(o.bbConfig, o.commit, true)
	if err != nil {
		return err
	}
	logger.Info("found [%d] pull request(s) in Bitbucket for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}
