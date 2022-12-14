package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/gitview"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type pipelineBackfillArtifactCommitsOptions struct {
	srcRepoRoot string
	payload     ArtifactCommitsBackfillPayload
}

type ArtifactCommitsBackfillPayload struct {
	RepoUrl     string                    `json:"repo_url"`
	CommitsList []*gitview.ArtifactCommit `json:"git_commit_list"`
}

func newPipelineBackfillArtifactCommitsCmd(out io.Writer) *cobra.Command {
	o := new(pipelineBackfillArtifactCommitsOptions)
	cmd := &cobra.Command{
		Use:    "backfill-commits PIPELINE-NAME",
		Short:  "Collect and report the changelog of each artifact in a Kosli pipeline.",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
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

	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"repo-root"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pipelineBackfillArtifactCommitsOptions) run(out io.Writer, args []string) error {
	// Get all artifacts for a pipeline
	// find repo URL
	// for each artifact,
	// 1) find the commit of the previous artifact
	// 2) get the commit list
	// 3) send a backfill request
	var err error
	pipelineName := args[0]

	gitView, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return err
	}

	o.payload.RepoUrl, err = gitView.RepoUrl()
	if err != nil {
		return err
	}

	pageNumber := 0
	for {
		pageNumber += 1
		artifactsRaw, err := getPipelineArtifacts(pipelineName, pageNumber)
		if err != nil {
			return err
		}
		if len(artifactsRaw) == 0 {
			return nil
		}
		for _, artifactRaw := range artifactsRaw {
			evidenceMap := artifactRaw["evidence"].(map[string]interface{})
			artifactData := evidenceMap["artifact"].(map[string]interface{})
			gitCommit := artifactData["git_commit"].(string)
			artifactDigest := artifactData["sha256"].(string)
			logger.Debug("Digest: %s -- git commit: %s \n", artifactDigest, gitCommit)

			previousCommitUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/previous_commit",
				global.Host, global.Owner, pipelineName, artifactDigest)

			previousCommitReqParams := &requests.RequestParams{
				Method:   http.MethodGet,
				URL:      previousCommitUrl,
				Password: global.ApiToken,
			}
			response, err := kosliClient.Do(previousCommitReqParams)
			if err != nil {
				return err
			}

			var previousCommitResponse map[string]interface{}
			err = json.Unmarshal([]byte(response.Body), &previousCommitResponse)
			if err != nil {
				return err
			}

			o.payload.CommitsList, err = gitView.ChangeLog(gitCommit, previousCommit(previousCommitResponse), logger)
			if err != nil {
				return err
			}

			for _, commitData := range o.payload.CommitsList {
				logger.Debug("	Commit sha1: %s\n", commitData.Sha1)
			}

			url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/backfill_commits", global.Host, global.Owner, pipelineName, artifactDigest)

			reqParams := &requests.RequestParams{
				Method:   http.MethodPut,
				URL:      url,
				Payload:  o.payload,
				DryRun:   global.DryRun,
				Password: global.ApiToken,
			}
			_, err = kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("[%d] commits reported for artifact %s", len(o.payload.CommitsList), artifactDigest)
			}
			return err
		}
	}
}

func previousCommit(previousCommitResponse map[string]interface{}) string {
	previousCommit := ""
	if previousCommitResponse["previous_commit"] != nil {
		previousCommit = previousCommitResponse["previous_commit"].(string)
		logger.Debug("Previous commit: %s", previousCommit)
	}
	return previousCommit
}

// getPipelineArtifacts returns artifacts from a pipeline
func getPipelineArtifacts(pipelineName string, pageNumber int) ([]map[string]interface{}, error) {
	var artifacts []map[string]interface{}
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/?page=%d&per_page=%d",
		global.Host, global.Owner, pipelineName, pageNumber, 15)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return artifacts, err
	}

	err = json.Unmarshal([]byte(response.Body), &artifacts)
	if err != nil {
		return artifacts, err
	}
	return artifacts, nil
}
