package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type pullRequestEvidenceBitbucketOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	description        string
	payload            PullRequestEvidencePayload
	bbUsername         string
	bbPassword         string
	bbWorkspace        string
	commit             string
	repository         string
	assert             bool
	userDataFile       string
}

const pullRequestEvidenceBitbucketShortDesc = `Report a Bitbucket pull request evidence for an artifact in a Kosli pipeline.`

const pullRequestEvidenceBitbucketLongDesc = pullRequestEvidenceBitbucketShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and report the pull-request evidence to the artifact in Kosli. 
` + sha256Desc

const pullRequestEvidenceBitbucketExample = `
# report a pull request evidence to kosli for a docker image
kosli pipeline artifact report evidence bitbucket-pullrequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli pipeline artifact report evidence bitbucket-pullrequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestEvidenceBitbucketCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestEvidenceBitbucketOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "bitbucket-pullrequest [ARTIFACT-NAME-OR-PATH]",
		Aliases: []string{"bb-pr", "bitbucket-pr"},
		Short:   pullRequestEvidenceBitbucketShortDesc,
		Long:    pullRequestEvidenceBitbucketLongDesc,
		Example: pullRequestEvidenceBitbucketExample,
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
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.bbUsername, "bitbucket-username", "", bbUsernameFlag)
	cmd.Flags().StringVar(&o.bbPassword, "bitbucket-password", "", bbPasswordFlag)
	cmd.Flags().StringVar(&o.bbWorkspace, "bitbucket-workspace", DefaultValue(ci, "workspace"), bbWorkspaceFlag)
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

	err := cmd.Flags().MarkDeprecated("evidence-type", "use --name instead")
	if err != nil {
		logger.Error("failed to configure deprecated flag: %v", err)
	}
	err = cmd.Flags().MarkDeprecated("description", "description is no longer used")
	if err != nil {
		logger.Error("failed to configure deprecated flag: %v", err)
	}
	err = cmd.Flags().MarkDeprecated("sha256", "use --fingerprint instead")
	if err != nil {
		logger.Error("failed to configure deprecated flag: %v", err)
	}
	err = RequireFlags(cmd, []string{"bitbucket-username", "bitbucket-password",
		"bitbucket-workspace", "commit", "repository", "pipeline", "build-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pullRequestEvidenceBitbucketOptions) run(args []string) error {
	var err error
	if o.payload.ArtifactFingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	// Get repository name from 'owner/repository_name' string
	o.repository = extractRepoName(o.repository)

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/evidence/pull_request", global.Host, global.Owner, o.pipelineName)
	pullRequestsEvidence, err := getPullRequestsFromBitbucketApi(o.bbWorkspace,
		o.repository, o.commit, o.bbUsername, o.bbPassword, o.assert)
	if err != nil {
		return err
	}

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}
	o.payload.GitProvider = "bitbucket"
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
		logger.Info("bitbucket pull request evidence is reported to artifact: %s", o.payload.ArtifactFingerprint)
	}
	return err
}

func getPullRequestsFromBitbucketApi(workspace, repository, commit, username, password string, assert bool) ([]*PrEvidence, error) {
	pullRequestsEvidence := []*PrEvidence{}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/commit/%s/pullrequests", workspace, repository, commit)
	logger.Debug("getting pull requests from " + url)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Username: username,
		Password: password,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return pullRequestsEvidence, err
	}
	if response.Resp.StatusCode == 200 {
		pullRequestsEvidence, err = parseBitbucketResponse(commit, workspace, repository, password, username, response, assert)
		if err != nil {
			return pullRequestsEvidence, err
		}
	} else if response.Resp.StatusCode == 202 {
		return pullRequestsEvidence, fmt.Errorf("repository pull requests are still being indexed, please retry")
	} else if response.Resp.StatusCode == 404 {
		return pullRequestsEvidence, fmt.Errorf("repository does not exist or pull requests are not indexed." +
			"Please make sure Pull Request Commit Links app is installed")
	} else {
		return pullRequestsEvidence, fmt.Errorf("failed to get pull requests from Bitbucket: %v", response.Body)
	}
	return pullRequestsEvidence, nil
}

func parseBitbucketResponse(commit, workspace, repository, password, username string, response *requests.HTTPResponse, assert bool) ([]*PrEvidence, error) {
	pullRequestsEvidence := []*PrEvidence{}
	var responseData map[string]interface{}
	err := json.Unmarshal([]byte(response.Body), &responseData)
	if err != nil {
		return pullRequestsEvidence, err
	}
	pullRequests, ok := responseData["values"].([]interface{})
	if !ok {
		return pullRequestsEvidence, nil
	}
	for _, prInterface := range pullRequests {
		pr := prInterface.(map[string]interface{})
		linksInterface := pr["links"].(map[string]interface{})
		apiLinkMap := linksInterface["self"].(map[string]interface{})
		htmlLinkMap := linksInterface["html"].(map[string]interface{})
		evidence, err := getPullRequestDetailsFromBitbucket(apiLinkMap["href"].(string), htmlLinkMap["href"].(string), workspace, repository, username, password, commit)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	if len(pullRequestsEvidence) == 0 {
		if assert {
			return pullRequestsEvidence, fmt.Errorf("no pull requests found for the given commit: %s", commit)
		}
		logger.Info("no pull requests found for given commit: " + commit)
	}
	return pullRequestsEvidence, nil
}

func getPullRequestDetailsFromBitbucket(prApiUrl, prHtmlLink, workspace, repository, username, password, commit string) (*PrEvidence, error) {
	logger.Debug("getting pull request details for " + prApiUrl)
	evidence := &PrEvidence{}

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      prApiUrl,
		Username: username,
		Password: password,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return evidence, err
	}
	if response.Resp.StatusCode == 200 {
		var responseData map[string]interface{}
		err := json.Unmarshal([]byte(response.Body), &responseData)
		if err != nil {
			return evidence, err
		}

		evidence.URL = prHtmlLink
		evidence.MergeCommit = commit
		evidence.State = responseData["state"].(string)
		participants := responseData["participants"].([]interface{})
		approvers := []string{}

		if len(participants) > 0 {
			for _, participantInterface := range participants {
				p := participantInterface.(map[string]interface{})
				if p["approved"].(bool) {
					user := p["user"].(map[string]interface{})
					approvers = append(approvers, user["display_name"].(string))
				}
			}
		} else {
			logger.Debug("no approvers found")
		}
		evidence.Approvers = approvers
		// prID := int(responseData["id"].(float64))
		// evidence.LastCommit, evidence.LastCommitter, err = getBitbucketPRLastCommit(workspace, repository, username, password, prID)
		// if err != nil {
		// 	return evidence, err
		// }
		// if utils.Contains(approvers, evidence.LastCommitter) {
		// 	evidence.SelfApproved = true
		// }
	} else {
		return evidence, fmt.Errorf("failed to get PR details, got HTTP status %d. Please review repository permissions", response.Resp.StatusCode)
	}
	return evidence, nil
}

// func getBitbucketPRLastCommit(workspace, repository, username, password string, prID int) (string, string, error) {
// 	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/commits", workspace, repository, prID)
// 	log.Debug("Getting pull requests commits from" + url)
// 	response, err := requests.SendPayload([]byte{}, url, username, password,
// 		global.MaxAPIRetries, false, http.MethodGet, log)
// 	if err != nil {
// 		return "", "", err
// 	}

// 	if response.Resp.StatusCode == 200 {
// 		var responseData map[string]interface{}
// 		err := json.Unmarshal([]byte(response.Body), &responseData)
// 		if err != nil {
// 			return "", "", err
// 		}
// 		prCommits := responseData["values"].([]interface{})

// 		// the first commit is the merge commit
// 		// TODO: is it safe to always to get the second commit?
// 		lastCommit := prCommits[1].(map[string]interface{})
// 		lastAuthor := lastCommit["author"].(map[string]interface{})
// 		return lastCommit["hash"].(string), lastAuthor["user"].(map[string]interface{})["display_name"].(string), nil

// 	} else {
// 		return "", "", fmt.Errorf("failed to get PR commits, got HTTP status %d", response.Resp.StatusCode)
// 	}
// }
