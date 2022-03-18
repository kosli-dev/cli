package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type pullRequestEvidenceBitbucketOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string // This is calculated or provided by the user
	pipelineName       string
	description        string
	buildUrl           string
	payload            EvidencePayload
	bbUsername         string
	bbPassword         string
	bbWorkspace        string
	commit             string
	repository         string
	assert             bool
}

type BitbucketPrEvidence struct {
	PullRequestMergeCommit string `json:"pullRequestMergeCommit"`
	PullRequestURL         string `json:"pullRequestURL"`
	PullRequestState       string `json:"pullRequestState"`
	Approvers              string `json:"approvers"`
}

func newPullRequestEvidenceBitbucketCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceBitbucketOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "bitbucket-pullrequest [ARTIFACT-NAME-OR-PATH]",
		Aliases: []string{"bb-pr", "bitbucket-pr"},
		Short:   "Report a Bitbucket pull request evidence for an artifact in a Merkely pipeline.",
		Long:    controlPullRequestDesc(),
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
	cmd.Flags().StringVar(&o.bbUsername, "bitbucket-username", "", bbUsernameFlag)
	cmd.Flags().StringVar(&o.bbPassword, "bitbucket-password", "", bbPasswordFlag)
	cmd.Flags().StringVar(&o.bbWorkspace, "bitbucket-workspace", DefaultValue(ci, "workspace"), bbWorkspaceFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringVar(&o.repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)

	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().StringVarP(&o.buildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceType, "evidence-type", "e", "", evidenceTypeFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"bitbucket-username", "bitbucket-password",
		"bitbucket-workspace", "commit", "repository", "pipeline", "build-url", "evidence-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestEvidenceBitbucketOptions) run(args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
	pullRequestsEvidence, isCompliant, err := getPullRequestsFromBitbucketApi(o.bbWorkspace,
		o.repository, o.commit, o.bbUsername, o.bbPassword, o.assert)
	if err != nil {
		return err
	}
	o.payload.Contents = map[string]interface{}{}
	o.payload.Contents["is_compliant"] = isCompliant
	o.payload.Contents["url"] = o.buildUrl
	o.payload.Contents["description"] = o.description
	o.payload.Contents["source"] = pullRequestsEvidence
	if err != nil {
		return err
	}

	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}

func getPullRequestsFromBitbucketApi(workspace, repository, commit, username, password string, assert bool) ([]*BitbucketPrEvidence, bool, error) {
	isCompliant := false
	pullRequestsEvidence := []*BitbucketPrEvidence{}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/commit/%s/pullrequests", workspace, repository, commit)
	log.Debug("Getting pull requests from " + url)
	response, err := requests.SendPayload([]byte{}, url, username, password,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return pullRequestsEvidence, false, err
	}
	if response.Resp.StatusCode == 200 {
		isCompliant, pullRequestsEvidence, err = parseBitbucketResponse(commit, password, username, response, assert)
		if err != nil {
			return pullRequestsEvidence, isCompliant, err
		}
	} else if response.Resp.StatusCode == 202 {
		return pullRequestsEvidence, isCompliant, fmt.Errorf("repository pull requests are still being indexed, please retry")
	} else if response.Resp.StatusCode == 404 {
		return pullRequestsEvidence, isCompliant, fmt.Errorf("repository does not exist or pull requests are not indexed." +
			"Please make sure Pull Request Commit Links app is installed")
	} else {
		return pullRequestsEvidence, isCompliant, fmt.Errorf("failed to get pull requests from Bitbucket: %v", response.Body)
	}
	return pullRequestsEvidence, isCompliant, nil
}

func parseBitbucketResponse(commit, password, username string, response *requests.HTTPResponse, assert bool) (bool, []*BitbucketPrEvidence, error) {
	log.Debug("Pull requests response: " + response.Body)
	pullRequestsEvidence := []*BitbucketPrEvidence{}
	isCompliant := false
	var responseData map[string]interface{}
	err := json.Unmarshal([]byte(response.Body), &responseData)
	if err != nil {
		return isCompliant, pullRequestsEvidence, err
	}
	pullRequests, ok := responseData["values"].([]interface{})
	if !ok {
		return isCompliant, pullRequestsEvidence, nil
	}
	for _, prInterface := range pullRequests {
		pr := prInterface.(map[string]interface{})
		linksInterface := pr["links"].(map[string]interface{})
		apiLinkMap := linksInterface["self"].(map[string]interface{})
		htmlLinkMap := linksInterface["html"].(map[string]interface{})
		evidence, err := getPullRequestDetailsFromBitbucket(apiLinkMap["href"].(string), htmlLinkMap["href"].(string), username, password, commit)
		if err != nil {
			return false, pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	if len(pullRequestsEvidence) > 0 {
		isCompliant = true
	} else {
		if assert {
			return isCompliant, pullRequestsEvidence, fmt.Errorf("no pull requests found for the given commit: %s", commit)
		}
		log.Info("No pull requests found for given commit: " + commit)
	}
	return isCompliant, pullRequestsEvidence, nil
}

func getPullRequestDetailsFromBitbucket(prApiUrl, prHtmlLink, username, password, commit string) (*BitbucketPrEvidence, error) {
	log.Debug("Getting pull request details for" + prApiUrl)
	evidence := &BitbucketPrEvidence{}
	response, err := requests.SendPayload([]byte{}, prApiUrl, username, password,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return evidence, err
	}
	if response.Resp.StatusCode == 200 {
		var responseData map[string]interface{}
		err := json.Unmarshal([]byte(response.Body), &responseData)
		if err != nil {
			return evidence, err
		}

		evidence.PullRequestURL = prHtmlLink
		evidence.PullRequestMergeCommit = commit
		evidence.PullRequestState = responseData["state"].(string)
		participants := responseData["participants"].([]interface{})
		approvers := ""

		if len(participants) > 0 {
			for _, participantInterface := range participants {
				p := participantInterface.(map[string]interface{})
				if p["approved"].(bool) {
					user := p["user"].(map[string]interface{})
					approvers = approvers + user["display_name"].(string) + ","
				}
			}
			approvers = strings.TrimSuffix(approvers, ",")
		} else {
			log.Debug("No approvers found")
		}
		evidence.Approvers = approvers

	} else {
		return evidence, fmt.Errorf("failed to get PR details, got HTTP status %d. Please review repository permissions", response.Resp.StatusCode)
	}
	return evidence, nil
}

func controlPullRequestDesc() string {
	return `
   Check if a pull request exists for an artifact and report the pull-request evidence to the artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"build-url"})
}
