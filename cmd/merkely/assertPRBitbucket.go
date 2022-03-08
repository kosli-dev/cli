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

func newAssertPullRequestBitbucketCmd(out io.Writer) *cobra.Command {
	o := new(assertPullRequestBitbucketOptions)
	cmd := &cobra.Command{
		Use:     "bitbucket-pullrequest",
		Aliases: []string{"bb-pr", "bitbucket-pr"},
		Short:   "Assert if a Bitbucket pull request for the commit which produces an artifact exists.",
		Long:    assertBBPullRequestDesc(),
		Args:    NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.bbUsername, "bitbucket-username", "", "Bitbucket user name.")
	cmd.Flags().StringVar(&o.bbPassword, "bitbucket-password", "", "Bitbucket password.")
	cmd.Flags().StringVar(&o.bbWorkspace, "bitbucket-workspace", DefaultValue(ci, "workspace"), "Bitbucket workspace.")
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), "Git commit for which to find pull request evidence.")
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), "Git repository.")

	err := RequireFlags(cmd, []string{"bitbucket-username", "bitbucket-password",
		"bitbucket-workspace", "commit", "repository"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertPullRequestBitbucketOptions) run(args []string) error {
	pullRequestsEvidence, _, err := getPullRequestsFromBitbucketApi(o.bbWorkspace,
		o.repository, o.commit, o.bbUsername, o.bbPassword, true)
	if err != nil {
		return err
	}
	log.Infof("found [%d] pull request(s) in Bitbucket for commit: %s", len(pullRequestsEvidence), o.commit)
	return nil
}

func assertBBPullRequestDesc() string {
	return `
   Check if a pull request exists in Bitbucket for an artifact (based on the git commit that produced it) and fail if it does not. `
}
