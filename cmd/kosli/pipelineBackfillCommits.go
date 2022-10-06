package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type pipelineBackfillArtifactCommitsOptions struct {
	pipelineName string
	srcRepoRoot  string
	payload      ArtifactCommitsBackfillPayload
}

type ArtifactCommitsBackfillPayload struct {
	RepoUrl     string            `json:"repo_url"`
	CommitsList []*ArtifactCommit `json:"git_commit_list"`
}

// const artifactCreationExample = `
// # Report to a Kosli pipeline that a file type artifact has been created
// kosli pipeline artifact report creation FILE.tgz \
// 	--api-token yourApiToken \
// 	--artifact-type file \
// 	--build-url https://exampleci.com \
// 	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--owner yourOrgName \
// 	--pipeline yourPipelineName

// # Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
// kosli pipeline artifact report creation \
// 	--api-token yourApiToken \
// 	--build-url https://exampleci.com \
// 	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--owner yourOrgName \
// 	--pipeline yourPipelineName \
// 	--sha256 yourSha256
// `

func newPipelineBackfillArtifactCommits(out io.Writer) *cobra.Command {
	o := new(pipelineBackfillArtifactCommitsOptions)
	cmd := &cobra.Command{
		Use:   "backfill-commits PIPELINE-NAME",
		Short: "Calculate and report the changelog of each artifact in a Kosli pipeline.",
		// Long:    pipelineBackfillArtifactCommitsDesc(),
		// Example: artifactCreationExample,
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "pipeline name argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)

	err := RequireFlags(cmd, []string{"repo-root"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pipelineBackfillArtifactCommitsOptions) run(args []string) error {
	// Get all artifacts for a pipeline
	// find repo URL
	// for each artifact,
	// find the commit of the previous artifact
	// get the commit list
	// send a backfill request
	var err error
	pipelineName := args[0]
	o.payload.RepoUrl, err = getRepoUrl(o.srcRepoRoot)
	if err != nil {
		return err
	}

	artifactsRaw, err := getPipelineArtifacts(pipelineName)
	if err != nil {
		return err
	}

	for _, artifactRaw := range artifactsRaw {
		evidenceMap := artifactRaw["evidence"].(map[string]interface{})
		artifactData := evidenceMap["artifact"].(map[string]interface{})
		gitCommit := artifactData["git_commit"].(string)
		artifactDigest := artifactData["sha256"].(string)

		previousCommitUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/previous_commit",
			global.Host, global.Owner, o.pipelineName, artifactDigest)

		response, err := requests.DoBasicAuthRequest([]byte{}, previousCommitUrl, "", global.ApiToken,
			global.MaxAPIRetries, http.MethodGet, map[string]string{}, log)
		if err != nil {
			return err
		}

		var previousCommitResponse map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &previousCommitResponse)
		if err != nil {
			return err
		}

		o.payload.CommitsList = []*ArtifactCommit{}
		if previousCommitResponse["previous_commit"] != nil {
			previousCommit := previousCommitResponse["previous_commit"].(string)
			o.payload.CommitsList, err = listCommitsBetween(o.srcRepoRoot, previousCommit, gitCommit)
			if err != nil {
				return err
			}
		}

		url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/backfill_commits", global.Host, global.Owner, o.pipelineName, artifactDigest)
		_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
			global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
		if err != nil {
			return err
		}
	}

	return err
}

func getPipelineArtifacts(pipelineName string) ([]map[string]interface{}, error) {
	var artifacts []map[string]interface{}
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/?page=%d&per_page=%d",
		global.Host, global.Owner, pipelineName, 1, 15)
	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return artifacts, err
	}

	err = json.Unmarshal([]byte(response.Body), &artifacts)
	if err != nil {
		return artifacts, err
	}
	return artifacts, nil
}
