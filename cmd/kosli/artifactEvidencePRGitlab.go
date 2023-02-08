package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/spf13/cobra"
)

type pullRequestEvidenceGitlabOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	description        string
	payload            PullRequestEvidencePayload
	glToken            string
	glOwner            string
	commit             string
	repository         string
	assert             bool
	userDataFile       string
}

const pullRequestEvidenceGitlabShortDesc = `Report a Gitlab merge request evidence for an artifact in a Kosli pipeline.`

const pullRequestEvidenceGitlabLongDesc = pullRequestEvidenceGitlabShortDesc + `
It checks if a merge request exists for the artifact (based on its git commit) and report the pull-request evidence to the artifact in Kosli. 
` + sha256Desc

const pullRequestEvidenceGitlabExample = `
# report a pull request evidence to kosli for a docker image
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-token yourGithubToken \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-token yourGithubToken \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestEvidenceGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceGitlabOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "gitlab-mergerequest [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   pullRequestEvidenceGitlabShortDesc,
		Long:    pullRequestEvidenceGitlabLongDesc,
		Example: pullRequestEvidenceGitlabExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if o.payload.EvidenceName == "" {
				return fmt.Errorf("--name is required")
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.glToken, "gitlab-token", "", gitlabTokenFlag)
	cmd.Flags().StringVar(&o.glOwner, "gitlab-org", "", gitlabOrgFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)

	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "fingerprint", "f", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "evidence-type", "e", "", evidenceTypeFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := DeprecateFlags(cmd, map[string]string{
		"evidence-type": "use --name instead",
		"description":   "description is no longer used",
		"sha256":        "use --fingerprint instead",
	})
	if err != nil {
		logger.Error("failed to configure deprecated flags: %v", err)
	}

	err = RequireFlags(cmd, []string{"gitlab-token", "commit",
		"repository", "pipeline", "build-url"})
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
	// Get repository name from 'owner/repository_name' string
	o.repository = extractRepoName(o.repository)
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

func (o *pullRequestEvidenceGitlabOptions) getGitlabPullRequests() ([]*PrEvidence, error) {
	pullRequestsEvidence := []*PrEvidence{}

	project := o.glOwner + "/" + o.repository

	client, err := gitlab.NewClient(o.glToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	mergerequests, _, err := client.Commits.ListMergeRequestsByCommit(project, o.commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mergerequest := range mergerequests {
		evidence, err := o.newPREvidence(mergerequest)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)

	}
	if len(pullRequestsEvidence) == 0 {
		if o.assert {
			return pullRequestsEvidence, fmt.Errorf("no pull requests found for the given commit: %s", o.commit)
		}
		logger.Info("no pull requests found for given commit: " + o.commit)
	}
	return pullRequestsEvidence, nil
}

// newPREvidence creates an evidence from a gitlab merge request
func (o *pullRequestEvidenceGitlabOptions) newPREvidence(mergerequest *gitlab.MergeRequest) (*PrEvidence, error) {
	evidence := &PrEvidence{}
	evidence.URL = mergerequest.WebURL
	evidence.MergeCommit = mergerequest.MergeCommitSHA
	evidence.State = mergerequest.State
	evidence.Approvers = []string{}
	return evidence, nil
}
