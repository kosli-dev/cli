package main

import (
	"fmt"
	"io"
	"net/http"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/kosli-dev/cli/internal/requests"
	gitlabSDK "github.com/xanzy/go-gitlab"

	"github.com/spf13/cobra"
)

type pullRequestEvidenceGitlabOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	payload            PullRequestEvidencePayload
	gitlabConfig       *gitlabUtils.GitlabConfig
	commit             string
	assert             bool
	userDataFile       string
}

const pullRequestEvidenceGitlabShortDesc = `Report a Gitlab merge request evidence for an artifact in a Kosli pipeline.`

const pullRequestEvidenceGitlabLongDesc = pullRequestEvidenceGitlabShortDesc + `
It checks if a merge request exists for the artifact (based on its git commit) and report the merge request evidence to the artifact in Kosli. 
` + sha256Desc

const pullRequestEvidenceGitlabExample = `
# report a merge request evidence to kosli for a docker image
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken

# report a merge request evidence (from an on-prem Gitlab) to kosli for a docker image 
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-base-url https://gitlab.example.org \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a merge request does not exist for your artifact
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestEvidenceGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceGitlabOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	o.gitlabConfig = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab-mergerequest [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gl-mr", "gitlab-mr"},
		Short:   pullRequestEvidenceGitlabShortDesc,
		Long:    pullRequestEvidenceGitlabLongDesc,
		Example: pullRequestEvidenceGitlabExample,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.gitlabConfig.Token, "gitlab-token", "", gitlabTokenFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.Org, "gitlab-org", DefaultValue(ci, "namespace"), gitlabOrgFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.BaseURL, "gitlab-base-url", "", gitlabBaseURLFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringVar(&o.gitlabConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "fingerprint", "f", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"gitlab-token", "gitlab-org", "commit", "name",
		"repository", "pipeline", "build-url",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestEvidenceGitlabOptions) run(out io.Writer, args []string) error {
	var err error
	if o.payload.ArtifactFingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/evidence/pull_request", global.Host, global.Owner, o.pipelineName)
	pullRequestsEvidence, err := getGitlabPullRequests(o.gitlabConfig, o.commit, o.assert)
	if err != nil {
		return err
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}
	o.payload.GitProvider = "gitlab"
	o.payload.PullRequests = pullRequestsEvidence

	logger.Debug("found %d merge request(s) for commit: %s\n", len(pullRequestsEvidence), o.commit)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("gitlab merge request evidence is reported to artifact: %s", o.payload.ArtifactFingerprint)
	}
	return err
}

func getGitlabPullRequests(gitlabConfig *gitlabUtils.GitlabConfig, commit string, assert bool) ([]*PrEvidence, error) {
	pullRequestsEvidence := []*PrEvidence{}
	mrs, err := gitlabConfig.MergeRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mr := range mrs {
		evidence, err := newPRGitlabEvidence(mr, gitlabConfig)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}

	if len(pullRequestsEvidence) == 0 {
		if assert {
			return pullRequestsEvidence, fmt.Errorf("no merge requests found for the given commit: %s", commit)
		}
		logger.Info("no merge requests found for given commit: " + commit)
	}
	return pullRequestsEvidence, nil
}

// newPRGitlabEvidence creates evidence from a gitlab merge request
func newPRGitlabEvidence(mr *gitlabSDK.MergeRequest, gitlabConfig *gitlabUtils.GitlabConfig) (*PrEvidence, error) {
	evidence := &PrEvidence{}
	evidence.URL = mr.WebURL
	evidence.MergeCommit = mr.MergeCommitSHA
	evidence.State = mr.State
	approvers, err := gitlabConfig.GetMergeRequestApprovers(mr.IID)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	return evidence, nil
}
