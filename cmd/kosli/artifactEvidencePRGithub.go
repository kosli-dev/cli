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

type PullRequestEvidencePayload struct {
	TypedEvidencePayload
	GitProvider  string        `json:"git_provider"`
	PullRequests []*PrEvidence `json:"pull_requests"`
}

type pullRequestEvidenceGithubOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	description        string
	payload            PullRequestEvidencePayload
	ghToken            string
	ghOwner            string
	commit             string
	repository         string
	assert             bool
	userDataFile       string
}

type PrEvidence struct {
	MergeCommit string   `json:"merge_commit"`
	URL         string   `json:"url"`
	State       string   `json:"state"`
	Approvers   []string `json:"approvers"`
	// LastCommit             string `json:"lastCommit"`
	// LastCommitter          string `json:"lastCommitter"`
	// SelfApproved           bool   `json:"selfApproved"`
}

const pullRequestEvidenceGithubShortDesc = `Report a Github pull request evidence for an artifact in a Kosli pipeline.`

const pullRequestEvidenceGithubLongDesc = pullRequestEvidenceGithubShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and report the pull-request evidence to the artifact in Kosli. 
` + sha256Desc

const pullRequestEvidenceGithubExample = `
# report a pull request evidence to kosli for a docker image
kosli pipeline artifact report evidence github-pullrequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli pipeline artifact report evidence github-pullrequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestEvidenceGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceGithubOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "github-pullrequest [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gh-pr", "github-pr"},
		Short:   pullRequestEvidenceGithubShortDesc,
		Long:    pullRequestEvidenceGithubLongDesc,
		Example: pullRequestEvidenceGithubExample,
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
	cmd.Flags().StringVar(&o.ghToken, "github-token", "", githubTokenFlag)
	cmd.Flags().StringVar(&o.ghOwner, "github-org", DefaultValue(ci, "owner"), githubOrgFlag)
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

	err = RequireFlags(cmd, []string{"github-token", "github-org", "commit",
		"repository", "pipeline", "build-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestEvidenceGithubOptions) run(out io.Writer, args []string) error {
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

	logger.Debug("found %d pull request(s) for commit: %s\n", len(pullRequestsEvidence), o.commit)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("github pull request evidence is reported to artifact: %s", o.payload.ArtifactFingerprint)
	}
	return err
}

func (o *pullRequestEvidenceGithubOptions) getGithubPullRequests() ([]*PrEvidence, error) {
	pullRequestsEvidence := []*PrEvidence{}

	pullrequests, err := ghUtils.PullRequestsForCommit(o.ghToken, o.ghOwner, o.repository, o.commit)
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
			return pullRequestsEvidence, fmt.Errorf("no pull requests found for the given commit: %s", o.commit)
		}
		logger.Info("no pull requests found for given commit: " + o.commit)
	}
	return pullRequestsEvidence, nil
}

// newPREvidence creates an evidence from a github pull request
func (o *pullRequestEvidenceGithubOptions) newPREvidence(pullrequest *gh.PullRequest) (*PrEvidence, error) {
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
