package main

import (
	"io"

	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	"github.com/spf13/cobra"
)

const reportEvidenceCommitPRBitbucketShortDesc = `Report Bitbucket pull request evidence for a commit in Kosli flows.  `

const reportEvidenceCommitPRBitbucketLongDesc = reportEvidenceCommitPRBitbucketShortDesc + `
It checks if a pull request exists for the git commit and reports the pull-request evidence to the commit in Kosli.`

const reportEvidenceCommitPRBitbucketExample = `
# report a pull request evidence to Kosli
kosli report evidence commit pullrequest bitbucket \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--org yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli report evidence commit pullrequest bitbucket \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--org yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newReportEvidenceCommitPRBitbucketCmd(out io.Writer) *cobra.Command {
	config := new(bbUtils.Config)
	config.Logger = logger
	config.KosliClient = kosliClient

	o := new(pullRequestCommitOptions)
	o.retriever = config

	cmd := &cobra.Command{
		Use:     "bitbucket",
		Aliases: []string{"bb"},
		Short:   reportEvidenceCommitPRBitbucketShortDesc,
		Long:    reportEvidenceCommitPRBitbucketLongDesc,
		Example: reportEvidenceCommitPRBitbucketExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
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
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"bitbucket-username", "bitbucket-password", "bitbucket-workspace",
		"commit", "repository", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
