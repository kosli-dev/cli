package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type pullRequestCommitEvidenceBitbucketOptions struct {
	bbUsername   string
	bbPassword   string
	bbWorkspace  string
	repository   string
	assert       bool
	userDataFile string
	payload      PullRequestCommitEvidencePayload
}

const pullRequestCommitEvidenceBitbucketShortDesc = `Report a Bitbucket pull request evidence for a commit in a Kosli pipeline.`

const pullRequestCommitEvidenceBitbucketLongDesc = pullRequestCommitEvidenceBitbucketShortDesc + `
It checks if a pull request exists for the git commit and reports the pull-request evidence to the commit in Kosli.`

const pullRequestCommitEvidenceBitbucketExample = `
# report a pull request evidence to Kosli
kosli commit report evidence bitbucket-pullrequest \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli commit report evidence bitbucket-pullrequest \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestCommitEvidenceBitbucketCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitEvidenceBitbucketOptions)
	cmd := &cobra.Command{
		Use:     "bitbucket-pullrequest",
		Aliases: []string{"bb-pr", "bitbucket-pr"},
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
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.bbUsername, "bitbucket-username", "", bbUsernameFlag)
	cmd.Flags().StringVar(&o.bbPassword, "bitbucket-password", "", bbPasswordFlag)
	cmd.Flags().StringVar(&o.bbWorkspace, "bitbucket-workspace", DefaultValue(ci, "workspace"), bbWorkspaceFlag)
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	cmd.Flags().StringVar(&o.payload.CommitSHA, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)

	cmd.Flags().StringSliceVarP(&o.payload.Pipelines, "pipelines", "p", []string{}, pipelinesFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"bitbucket-username", "bitbucket-password", "bitbucket-workspace",
		"commit", "repository", "pipelines", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestCommitEvidenceBitbucketOptions) run(args []string) error {
	var err error

	// Get repository name from 'owner/repository_name' string
	o.repository = extractRepoName(o.repository)

	url := fmt.Sprintf("%s/api/v1/projects/%s/commit/evidence/pull_request", global.Host, global.Owner)

	pullRequestsEvidence, err := getPullRequestsFromBitbucketApi(o.bbWorkspace,
		o.repository, o.payload.CommitSHA, o.bbUsername, o.bbPassword, o.assert)
	if err != nil {
		return err
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}
	o.payload.GitProvider = "bitbucket"
	o.payload.PullRequests = pullRequestsEvidence

	logger.Debug("found %d pull request(s) for commit: %s\n", len(pullRequestsEvidence), o.payload.CommitSHA)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("bitbucket pull request commit evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}
