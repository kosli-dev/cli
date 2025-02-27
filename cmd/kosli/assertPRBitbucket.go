package main

import (
	"fmt"
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
The command exits with non-zero exit code if no pull requests were found for the commit.
Authentication to Bitbucket can be done with access token (recommended) or app passwords. Credentials need to have read access for both repos and pull requests.`

const assertPRBitbucketExample = `
kosli assert pullrequest bitbucket  \
	--bitbucket-access-token yourBitbucketAccessToken \
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := MuXRequiredFlags(cmd, []string{"bitbucket-username", "bitbucket-access-token"}, true)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"bitbucket-password", "bitbucket-access-token"}, true)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	addBitbucketFlags(cmd, o.bbConfig, ci)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValueForCommit(ci, true), commitPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"bitbucket-workspace", "commit", "repository"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertPullRequestBitbucketOptions) run(args []string) error {
	pullRequestsEvidence, err := o.bbConfig.PREvidenceForCommit(o.commit)
	if err != nil {
		return err
	}
	if len(pullRequestsEvidence) == 0 {
		return fmt.Errorf("assert failed: found no pull request(s) in Bitbucket for commit: %s", o.commit)
	}
	logger.Info("found [%d] pull request(s) in Bitbucket for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}
