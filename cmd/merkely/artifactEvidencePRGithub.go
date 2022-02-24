package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	gh "github.com/google/go-github/v42/github"
	"github.com/merkely-development/reporter/internal/requests"
	"golang.org/x/oauth2"

	"github.com/spf13/cobra"
)

type pullRequestEvidenceGithubOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string // This is calculated or provided by the user
	pipelineName       string
	description        string
	buildUrl           string
	payload            EvidencePayload
	ghToken            string
	ghOwner            string
	commit             string
	repository         string
	assert             bool
}

type GithubPrEvidence struct {
	PullRequestMergeCommit string `json:"pullRequestMergeCommit"`
	PullRequestURL         string `json:"pullRequestURL"`
	PullRequestState       string `json:"pullRequestState"`
	Approvers              string `json:"approvers"`
}

func newPullRequestEvidenceGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceGithubOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "github-pullrequest ARTIFACT-NAME-OR-PATH",
		Aliases: []string{"gh-pr", "github-pr"},
		Short:   "Report a Github pull request evidence for an artifact in a Merkely pipeline.",
		Long:    controlPullRequestGithubDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.sha256, false)
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return ValidateRegisteryFlags(cmd, o.fingerprintOptions)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.ghToken, "github-token", "", "Github token.")
	cmd.Flags().StringVar(&o.ghOwner, "github-org", DefaultValue(ci, "owner"), "Github organization.")
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), "Git commit for which to find pull request evidence.")
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), "Git repository.")

	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "[optional] The evidence description.")
	cmd.Flags().StringVarP(&o.buildUrl, "build-url", "b", DefaultValue(ci, "build-url"), "The url of CI pipeline that generated the evidence.")
	cmd.Flags().StringVarP(&o.payload.EvidenceType, "evidence-type", "e", "", "The type of evidence being reported.")
	cmd.Flags().BoolVar(&o.assert, "assert", false, "Fail if no pull requests found for the given commit.")
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"github-token", "github-org", "commit",
		"repository", "pipeline", "build-url", "evidence-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestEvidenceGithubOptions) run(args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
	pullRequestsEvidence, isCompliant, err := o.getGithubPullRequests()
	if err != nil {
		return err
	}

	o.payload.Contents = map[string]interface{}{}
	o.payload.Contents["is_compliant"] = isCompliant
	o.payload.Contents["url"] = o.buildUrl
	o.payload.Contents["description"] = o.description
	o.payload.Contents["source"] = pullRequestsEvidence

	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}

func (o *pullRequestEvidenceGithubOptions) getGithubPullRequests() ([]*GithubPrEvidence, bool, error) {
	owner := o.ghOwner
	// Get repository name from 'owner/repository_name' string
	repoNameParts := strings.Split(o.repository, "/")
	repository := repoNameParts[len(repoNameParts)-1]
	commit := o.commit

	pullRequestsEvidence := []*GithubPrEvidence{}
	isCompliant := false

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: o.ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := gh.NewClient(tc)
	pullrequests, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repository,
		commit, &gh.PullRequestListOptions{})
	if err != nil {
		return pullRequestsEvidence, isCompliant, err
	}

	for _, pullrequest := range pullrequests {
		evidence := &GithubPrEvidence{}
		evidence.PullRequestURL = pullrequest.GetHTMLURL()
		evidence.PullRequestMergeCommit = pullrequest.GetMergeCommitSHA()
		evidence.PullRequestState = pullrequest.GetState()

		approvers, err := getPullRequestApprovers(client, ctx, owner, repository,
			pullrequest.GetNumber())
		if err != nil {
			return pullRequestsEvidence, isCompliant, err
		}
		evidence.Approvers = approvers
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	if len(pullRequestsEvidence) > 0 {
		isCompliant = true
	} else {
		if o.assert {
			return pullRequestsEvidence, isCompliant, fmt.Errorf("no pull requests found for the given commit: %s", commit)
		}
		log.Info("No pull requests found for given commit: " + commit)
	}
	return pullRequestsEvidence, isCompliant, nil
}

func getPullRequestApprovers(client *gh.Client, context context.Context, owner, repo string, number int) (string, error) {
	approvers := ""
	reviews, _, err := client.PullRequests.ListReviews(context, owner, repo, number, &gh.ListOptions{})
	if err != nil {
		return approvers, err
	}
	for _, r := range reviews {
		if r.GetState() == "APPROVED" {
			approvers = approvers + r.GetUser().GetLogin() + ","
		}
	}
	approvers = strings.TrimSuffix(approvers, ",")
	return approvers, nil
}

func controlPullRequestGithubDesc() string {
	return `
   Check if a pull request exists for an artifact and report the pull-request evidence to the artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"build-url"})
}
