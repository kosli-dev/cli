package main

import (
	"io"

	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	"github.com/spf13/cobra"
)

const pullRequestCommitEvidenceBitbucketShortDesc = `Report a Bitbucket pull request evidence for a commit in a Kosli flow.`

const pullRequestCommitEvidenceBitbucketLongDesc = pullRequestCommitEvidenceBitbucketShortDesc + `
It checks if a pull request exists for the git commit and reports the pull-request evidence to the commit in Kosli.`

const pullRequestCommitEvidenceBitbucketExample = `
# report a pull request evidence to Kosli
kosli report evidence commit pullrequest bitbucket \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli report evidence commit pullrequest bitbucket \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestCommitEvidenceBitbucketCmd(out io.Writer) *cobra.Command {
	config := new(bbUtils.Config)
	config.Logger = logger
	config.KosliClient = kosliClient

	o := new(pullRequestCommitOptions)
	o.retriever = config

	cmd := &cobra.Command{
		Use:     "bitbucket",
		Aliases: []string{"bb"},
		Short:   pullRequestCommitEvidenceBitbucketShortDesc,
		Long:    pullRequestCommitEvidenceBitbucketLongDesc,
		Example: pullRequestCommitEvidenceBitbucketExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.retriever.(*bbUtils.Config).Assert = o.assert
			return o.run(args)
		},
	}

	ci := WhichCI()
	addBitbucketFlags(cmd, o.retriever.(*bbUtils.Config), ci)
	addCommitPRFlags(cmd, o, ci)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"bitbucket-username", "bitbucket-password", "bitbucket-workspace",
		"commit", "repository", "flow", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
