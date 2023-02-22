package main

import (
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

const pullRequestCommitEvidenceGitlabShortDesc = `Report a Gitlab merge request evidence for a commit in a Kosli pipeline.`

const pullRequestCommitEvidenceGitlabLongDesc = pullRequestCommitEvidenceGitlabShortDesc + `
It checks if a merge request exists for the git commit and reports the merge-request evidence to the commit in Kosli.`

const pullRequestCommitEvidenceGitlabExample = `
# report a merge request evidence to Kosli
kosli report evidence commit pullrequest gitlab \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--name yourEvidenceName \
	--flow yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli report evidence commit pullrequest gitlab \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--name yourEvidenceName \
	--flow yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestCommitEvidenceGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitOptions)
	o.retriever = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab",
		Aliases: []string{"gl"},
		Short:   pullRequestCommitEvidenceGitlabShortDesc,
		Long:    pullRequestCommitEvidenceGitlabLongDesc,
		Example: pullRequestCommitEvidenceGitlabExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
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
	// cmd.Flags().StringVar(&o.gitlabConfig.Token, "gitlab-token", "", gitlabTokenFlag)
	// cmd.Flags().StringVar(&o.gitlabConfig.Org, "gitlab-org", DefaultValue(ci, "namespace"), gitlabOrgFlag)
	// cmd.Flags().StringVar(&o.gitlabConfig.BaseURL, "gitlab-base-url", "", gitlabBaseURLFlag)
	// cmd.Flags().StringVar(&o.gitlabConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	// cmd.Flags().StringVar(&o.payload.CommitSHA, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	addGitlabFlags(cmd, o.retriever.(*gitlabUtils.GitlabConfig), ci)
	addCommitPRFlags(cmd, o, ci)

	// cmd.Flags().StringSliceVarP(&o.payload.Pipelines, "pipelines", "p", []string{}, pipelinesFlag)
	// cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	// cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	// cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"gitlab-token", "gitlab-org", "commit",
		"repository", "flow", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
