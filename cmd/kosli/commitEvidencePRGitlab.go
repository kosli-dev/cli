package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/gitlab"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	gitlabSDK "github.com/xanzy/go-gitlab"
)

type pullRequestCommitEvidenceGitlabOptions struct {
	gitlabConfig *gitlab.GitlabConfig
	assert       bool
	userDataFile string
	payload      PullRequestCommitEvidencePayload
}

const pullRequestCommitEvidenceGitlabShortDesc = `Report a Gitlab merge request evidence for a commit in a Kosli pipeline.`

const pullRequestCommitEvidenceGitlabLongDesc = pullRequestCommitEvidenceGitlabShortDesc + `
It checks if a merge request exists for the git commit and reports the merge-request evidence to the commit in Kosli.`

const pullRequestCommitEvidenceGitlabExample = `
# report a merge request evidence to Kosli
kosli commit report evidence gitlab-mergerequest \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli commit report evidence gitlab-mergerequest \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestCommitEvidenceGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitEvidenceGitlabOptions)
	o.gitlabConfig = new(gitlab.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab-mergerequest",
		Short:   pullRequestCommitEvidenceGitlabShortDesc,
		Long:    pullRequestCommitEvidenceGitlabLongDesc,
		Example: pullRequestCommitEvidenceGitlabExample,
		Hidden:  true,
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
	cmd.Flags().StringVar(&o.gitlabConfig.Token, "gitlab-token", "", gitlabTokenFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.Org, "gitlab-org", "", gitlabOrgFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.BaseURL, "gitlab-base-url", "", gitlabBaseURLFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	cmd.Flags().StringVar(&o.payload.CommitSHA, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)

	cmd.Flags().StringSliceVarP(&o.payload.Pipelines, "pipelines", "p", []string{}, pipelinesFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"gitlab-token", "gitlab-org", "commit",
		"repository", "pipelines", "build-url",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestCommitEvidenceGitlabOptions) run(args []string) error {
	var err error

	url := fmt.Sprintf("%s/api/v1/projects/%s/commit/evidence/pull_request", global.Host, global.Owner)
	pullRequestsEvidence, err := o.getGitlabPullRequests()
	if err != nil {
		return err
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}
	o.payload.GitProvider = "gitlab"
	o.payload.PullRequests = pullRequestsEvidence

	logger.Debug("found %d merge request(s) for commit: %s\n", len(pullRequestsEvidence), o.payload.CommitSHA)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("gitlab merge request evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}

func (o *pullRequestCommitEvidenceGitlabOptions) getGitlabPullRequests() ([]*PrEvidence, error) {
	pullRequestsEvidence := []*PrEvidence{}
	mrs, err := o.gitlabConfig.MergeRequestsForCommit(o.payload.CommitSHA)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mr := range mrs {
		evidence, err := o.newPREvidence(mr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}

	if len(pullRequestsEvidence) == 0 {
		if o.assert {
			return pullRequestsEvidence, fmt.Errorf("no merge requests found for the given commit: %s", o.payload.CommitSHA)
		}
		logger.Info("no merge requests found for given commit: " + o.payload.CommitSHA)
	}
	return pullRequestsEvidence, nil
}

// newPREvidence creates an evidence from a gitlab merge request
func (o *pullRequestCommitEvidenceGitlabOptions) newPREvidence(mr *gitlabSDK.MergeRequest) (*PrEvidence, error) {
	evidence := &PrEvidence{}
	evidence.URL = mr.WebURL
	evidence.MergeCommit = mr.MergeCommitSHA
	evidence.State = mr.State
	approvers, err := o.gitlabConfig.GetMergeRequestApprovers(mr.IID)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	return evidence, nil
}
