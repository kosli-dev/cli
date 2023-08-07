package main

import (
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

const reportEvidenceCommitPRGitlabShortDesc = `Report Gitlab merge request evidence for a commit in Kosli flows.  `

const reportEvidenceCommitPRGitlabLongDesc = reportEvidenceCommitPRGitlabShortDesc + `
It checks if a merge request exists for the git commit and reports the merge-request evidence to the commit in Kosli.`

const reportEvidenceCommitPRGitlabExample = `
# report a merge request evidence to Kosli
kosli report evidence commit pullrequest gitlab \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--org yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli report evidence commit pullrequest gitlab \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--org yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newReportEvidenceCommitPRGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitOptions)
	o.retriever = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab",
		Aliases: []string{"gl"},
		Short:   reportEvidenceCommitPRGitlabShortDesc,
		Long:    reportEvidenceCommitPRGitlabLongDesc,
		Example: reportEvidenceCommitPRGitlabExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	addGitlabFlags(cmd, o.retriever.(*gitlabUtils.GitlabConfig), ci)
	addCommitPRFlags(cmd, o, ci)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"gitlab-token", "gitlab-org", "commit",
		"repository", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
