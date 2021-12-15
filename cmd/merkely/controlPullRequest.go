package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type pullRequestOptions struct {
	artifactType string
	sha256       string // This is calculated or provided by the user
	pipelineName string
	description  string
	buildUrl     string
	provider     string
	payload      EvidencePayload
}

type prEvidence struct {
	pullRequestMergeCommit string
	pullRequestURL         string
	pullRequestState       string
	approvers              string
}

func newControlPullRequestCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestOptions)
	cmd := &cobra.Command{
		Use:     "pullrequest ARTIFACT-NAME-OR-PATH",
		Aliases: []string{"pull-request", "pr"},
		Short:   "Check if a pull request exists for an artifact and report the pull-request evidence to the artifact in Merkely.",
		Long:    controlPullRequestDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return ValidateArtifactArg(args, o.artifactType, o.sha256)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if o.sha256 == "" {
				o.sha256, err = GetSha256Digest(o.artifactType, args[0])
				if err != nil {
					return err
				}
			}

			url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
			pullRequestsEvidence, isCompliant, err := getPullRequestForCurrentCommit()
			if err != nil {
				return err
			}
			o.payload.Contents = map[string]interface{}{}
			o.payload.Contents["isCompliant"] = isCompliant
			o.payload.Contents["url"] = o.buildUrl
			o.payload.Contents["description"] = o.description
			o.payload.Contents["source"] = pullRequestsEvidence
			if err != nil {
				return err
			}

			_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
				global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
			return err
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.provider, "provider", "bitbucket", "The source code repository provider name. Options are [bitbucket].")
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact related to the evidence. Options are [dir, file, docker].")
	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "[optional] The evidence description.")
	cmd.Flags().StringVarP(&o.buildUrl, "build-url", "b", DefaultValue(ci, "build-url"), "The url of CI pipeline that generated the evidence.")
	cmd.Flags().StringVarP(&o.payload.EvidenceType, "evidence-type", "e", "", "The type of evidence being reported.")

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "evidence-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func getPullRequestForCurrentCommit() ([]*prEvidence, bool, error) {
	workspace := os.Getenv("BITBUCKET_WORKSPACE")
	repository := os.Getenv("BITBUCKET_REPO_SLUG")
	commit := os.Getenv("BITBUCKET_COMMIT")
	user := os.Getenv("BITBUCKET_API_USER")
	password := os.Getenv("BITBUCKET_API_TOKEN")

	return getPullRequestsFromBitbucketApi(workspace, repository, commit, user, password)
}

func getPullRequestsFromBitbucketApi(workspace, repository, commit, username, password string) ([]*prEvidence, bool, error) {
	isCompliant := false
	pullRequestsEvidence := []*prEvidence{}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/commit/%s/pullrequests", workspace, repository, commit)
	log.Debug("Getting pull requests from " + url)
	response, err := requests.SendPayload([]byte{}, url, username, password,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return pullRequestsEvidence, false, err
	}
	if response.StatusCode == 200 {
		isCompliant, pullRequestsEvidence, err = parseBitbucketResponse(commit, password, username, response)
		if err != nil {
			return pullRequestsEvidence, isCompliant, err
		}
	} else if response.StatusCode == 202 {
		return pullRequestsEvidence, isCompliant, fmt.Errorf("repository pull requests are still being indexed, please retry")
	} else if response.StatusCode == 404 {
		return pullRequestsEvidence, isCompliant, fmt.Errorf("repository does not exist or pull requests are not indexed." +
			"Please make sure Pull Request Commit Links app is installed")
	} else {
		return pullRequestsEvidence, isCompliant, fmt.Errorf("failed to get pull requests from Bitbucket: %v", response.Body)
	}
	return pullRequestsEvidence, isCompliant, nil
}

func parseBitbucketResponse(commit, password, username string, response *requests.HTTPResponse) (bool, []*prEvidence, error) {
	log.Debug("Pull requests response: " + response.Body)
	pullRequestsEvidence := []*prEvidence{}
	isCompliant := false
	var responseData map[string]interface{}
	err := json.Unmarshal([]byte(response.Body), &responseData)
	if err != nil {
		return isCompliant, pullRequestsEvidence, err
	}
	pullRequests, ok := responseData["values"].([]map[string]interface{})
	if !ok {
		return isCompliant, pullRequestsEvidence, nil
	}

	for _, pr := range pullRequests {
		links := pr["links"].(map[string]map[string]string)
		evidence, err := getPullRequestDetailsFromBitbucket(links["self"]["href"], links["html"]["href"], username, password, commit)
		if err != nil {
			return false, pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	if len(pullRequestsEvidence) > 0 {
		isCompliant = true
	}
	return isCompliant, pullRequestsEvidence, nil
}

func getPullRequestDetailsFromBitbucket(prApiUrl, prHtmlLink, username, password, commit string) (*prEvidence, error) {
	log.Debug("Getting pull request details for" + prApiUrl)
	evidence := &prEvidence{}
	response, err := requests.SendPayload([]byte{}, prApiUrl, username, password,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return evidence, err
	}
	if response.StatusCode == 200 {
		var responseData map[string]interface{}
		err := json.Unmarshal([]byte(response.Body), &responseData)
		if err != nil {
			return evidence, err
		}

		evidence.pullRequestURL = prHtmlLink
		evidence.pullRequestMergeCommit = commit
		evidence.pullRequestState = responseData["state"].(string)
		participants := responseData["participants"].([]map[string]interface{})
		approvers := ""

		if len(participants) > 0 {
			for _, p := range participants {
				if p["approved"].(bool) {
					user := p["user"].(map[string]interface{})
					approvers = approvers + user["display_name"].(string) + ","
				}
			}
			approvers = strings.TrimSuffix(approvers, ",")
		} else {
			log.Debug("No approvers found")
		}
		evidence.approvers = approvers

	} else {
		return evidence, fmt.Errorf("failed to get PR details, got HTTP status %d. Please review repository permissions", response.StatusCode)
	}
	return evidence, nil
}

func controlPullRequestDesc() string {
	return `
   Check if a pull request exists for an artifact and report the pull-request evidence to the artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"build-url"})
}
