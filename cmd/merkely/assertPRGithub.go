package main

import (
	"io"

	"github.com/spf13/cobra"
)

func newAssertPullRequestGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceGithubOptions)
	cmd := &cobra.Command{
		Use:     "github-pullrequest",
		Aliases: []string{"gh-pr", "github-pr"},
		Short:   "Assert if a Github pull request for the commit which produces an artifact exists.",
		Long:    assertGHPullRequestDesc(),
		Args:    NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.assert = true
			pullRequestsEvidence, _, err := o.getGithubPullRequests()
			if err != nil {
				return err
			}
			log.Infof("found [%d] pull request(s) in Github for commit: %s", len(pullRequestsEvidence), o.commit)
			return nil
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.ghToken, "github-token", "", "Github token.")
	cmd.Flags().StringVar(&o.ghOwner, "github-org", DefaultValue(ci, "owner"), "Github organization.")
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), "Git commit for which to find pull request evidence.")
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), "Git repository.")

	err := RequireFlags(cmd, []string{"github-token", "github-org", "commit", "repository"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func assertGHPullRequestDesc() string {
	return `
   Check if a pull request exists in Github for an artifact (based on the git commit that produced it) and fail if it does not. `
}
