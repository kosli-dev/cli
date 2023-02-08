package main

import (
	"fmt"
	"io"
	"net/http"

	gh "github.com/google/go-github/v42/github"
	ghUtils "github.com/kosli-dev/cli/internal/github"
	"github.com/kosli-dev/cli/internal/requests"

	"github.com/spf13/cobra"
)

type PullRequestCommitEvidencePayload struct {
	CommitSHA    string        `json:"commit_sha"`
	Pipelines    []string      `json:"pipelines,omitempty"`
	EvidenceName string        `json:"name"`
	BuildUrl     string        `json:"build_url"`
	GitProvider  string        `json:"git_provider"`
	PullRequests []*PrEvidence `json:"pull_requests"`
	UserData     interface{}   `json:"user_data"`
}

type pullRequestCommitEvidenceGithubOptions struct {
	ghToken      string
	ghOwner      string
	repository   string
	assert       bool
	userDataFile string
	payload      PullRequestCommitEvidencePayload
}

const pullRequestCommitEvidenceGithubShortDesc = `Report a Github pull request evidence for a git commit in a Kosli pipeline.`

const pullRequestCommitEvidenceGithubLongDesc = pullRequestCommitEvidenceGithubShortDesc + `
It checks if a pull request exists for a commit and report the pull-request evidence to the commit in Kosli. 
`

const pullRequestCommitEvidenceGithubExample = `
# report a pull request commit evidence to Kosli
kosli commit report evidence github-pullrequest \
	--commit yourGitCommitSha1 \
	--repository yourGithubGitRepository \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--name yourEvidenceName \
	--pipelines yourPipelineName1,yourPipelineName2 \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli commit report evidence github-pullrequest \
	--commit yourGitCommitSha1 \
	--repository yourGithubGitRepository \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--name yourEvidenceName \
	--pipelines yourPipelineName1,yourPipelineName2 \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

// TODO: do we need to support assert for this command? see line 74

func newPullRequestCommitEvidenceGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestCommitEvidenceGithubOptions)
	cmd := &cobra.Command{
		Use:     "github-pullrequest",
		Aliases: []string{"gh-pr", "github-pr"},
		Short:   pullRequestCommitEvidenceGithubShortDesc,
		Long:    pullRequestCommitEvidenceGithubLongDesc,
		Example: pullRequestCommitEvidenceGithubExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.ghToken, "github-token", "", githubTokenFlag)
	cmd.Flags().StringVar(&o.ghOwner, "github-org", DefaultValue(ci, "owner"), githubOrgFlag)
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	cmd.Flags().StringVar(&o.payload.CommitSHA, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)

	cmd.Flags().StringSliceVarP(&o.payload.Pipelines, "pipelines", "p", []string{}, pipelinesFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"github-token", "github-org", "commit",
		"repository", "pipelines", "build-url",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestCommitEvidenceGithubOptions) run(out io.Writer, args []string) error {
	var err error

	url := fmt.Sprintf("%s/api/v1/projects/%s/commit/evidence/pull_request", global.Host, global.Owner)

	// Get repository name from 'owner/repository_name' string
	o.repository = extractRepoName(o.repository)
	pullRequestsEvidence, err := o.getGithubPullRequests()
	if err != nil {
		return err
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}
	o.payload.GitProvider = "github"
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
		logger.Info("github pull request commit evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}

func (o *pullRequestCommitEvidenceGithubOptions) getGithubPullRequests() ([]*PrEvidence, error) {
	pullRequestsEvidence := []*PrEvidence{}

	pullrequests, err := ghUtils.PullRequestsForCommit(o.ghToken, o.ghOwner, o.repository, o.payload.CommitSHA)
	if err != nil {
		return pullRequestsEvidence, err
	}

	for _, pullrequest := range pullrequests {
		evidence, err := o.newPREvidence(pullrequest)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)

	}
	if len(pullRequestsEvidence) == 0 {
		if o.assert {
			return pullRequestsEvidence, fmt.Errorf("no pull requests found for the given commit: %s", o.payload.CommitSHA)
		}
		logger.Info("no pull requests found for given commit: " + o.payload.CommitSHA)
	}
	return pullRequestsEvidence, nil
}

// newPREvidence creates an evidence from a github pull request
func (o *pullRequestCommitEvidenceGithubOptions) newPREvidence(pullrequest *gh.PullRequest) (*PrEvidence, error) {
	evidence := &PrEvidence{}
	evidence.URL = pullrequest.GetHTMLURL()
	evidence.MergeCommit = pullrequest.GetMergeCommitSHA()
	evidence.State = pullrequest.GetState()

	approvers, err := ghUtils.GetPullRequestApprovers(o.ghToken, o.ghOwner, o.repository,
		pullrequest.GetNumber())
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	return evidence, nil

	// lastCommit := pullrequest.Head.GetSHA()
	// opts := gh.ListOptions{}
	// commit, _, err := client.Repositories.GetCommit(ctx, owner, repository, lastCommit, &opts)
	// if err != nil {
	// 	return pullRequestsEvidence, isCompliant, err
	// }
	// evidence.LastCommit = lastCommit
	// evidence.LastCommitter = commit.GetAuthor().GetLogin()
	// if utils.Contains(approvers, evidence.LastCommitter) {
	// 	evidence.SelfApproved = true
	// }
}
