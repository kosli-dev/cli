package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type artifactCreationOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	srcRepoRoot        string
	payload            ArtifactPayload
}

type ArtifactPayload struct {
	Sha256      string                `json:"sha256"`
	Filename    string                `json:"filename"`
	Description string                `json:"description"`
	GitCommit   string                `json:"git_commit"`
	BuildUrl    string                `json:"build_url"`
	CommitUrl   string                `json:"commit_url"`
	RepoUrl     string                `json:"repo_url"`
	CommitsList []*gitview.CommitInfo `json:"commits_list"`
}

const artifactCreationShortDesc = `Report an artifact creation to a Kosli pipeline.`

const artifactCreationLongDesc = artifactCreationShortDesc + `
` + sha256Desc

const artifactCreationExample = `
# Report to a Kosli pipeline that a file type artifact has been created
kosli pipeline artifact report creation FILE.tgz \
	--api-token yourApiToken \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
kosli pipeline artifact report creation ANOTHER_FILE.txt \
	--api-token yourApiToken \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 
`

func newArtifactCreationCmd(out io.Writer) *cobra.Command {
	o := new(artifactCreationOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "creation {IMAGE-NAME | FILE-PATH | DIR-PATH}",
		Short:   artifactCreationShortDesc,
		Long:    artifactCreationLongDesc,
		Example: artifactCreationExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Sha256, true)
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
	cmd.Flags().StringVarP(&o.payload.Sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", artifactDescriptionFlag)
	cmd.Flags().StringVarP(&o.payload.GitCommit, "git-commit", "g", DefaultValue(ci, "git-commit"), gitCommitFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.CommitUrl, "commit-url", "u", DefaultValue(ci, "commit-url"), commitUrlFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"pipeline", "git-commit", "build-url", "commit-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *artifactCreationOptions) run(args []string) error {
	if o.payload.Sha256 != "" {
		o.payload.Filename = args[0]
	} else {
		var err error
		o.payload.Sha256, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
		if o.fingerprintOptions.artifactType == "dir" || o.fingerprintOptions.artifactType == "file" {
			o.payload.Filename = filepath.Base(args[0])
		} else {
			o.payload.Filename = args[0]
		}
	}

	gitView, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return err
	}

	previousCommit, err := o.latestCommit(currentBranch(gitView))
	if err != nil {
		return err
	}

	o.payload.CommitsList, err = gitView.ChangeLog(o.payload.GitCommit, previousCommit, logger)
	if err != nil {
		return err
	}

	o.payload.RepoUrl, err = gitView.RepoUrl()
	if err != nil {
		logger.Warning("Repo URL will not be reported, %s", err.Error())
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/", global.Host, global.Owner, o.pipelineName)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("artifact %s was reported with fingerprint: %s", o.payload.Filename, o.payload.Sha256)
	}
	return err
}

// latestCommit retrieves the git commit of the latest artifact for a pipeline in Kosli
func (o *artifactCreationOptions) latestCommit(branchName string) (string, error) {
	latestCommitUrl := fmt.Sprintf(
		"%s/api/v1/projects/%s/%s/artifacts/%s/latest_commit%s",
		global.Host, global.Owner, o.pipelineName, o.payload.Sha256, asBranchParameter(branchName))

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      latestCommitUrl,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return "", err
	}

	var latestCommitResponse map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &latestCommitResponse)
	if err != nil {
		return "", err
	}
	latestCommit := latestCommitResponse["latest_commit"]
	if latestCommit == nil {
		logger.Debug("no previous artifacts were found for pipeline: %s", o.pipelineName)
		return "", nil
	} else {
		logger.Debug("latest artifact for pipeline: %s has the git commit: %s", o.pipelineName, latestCommit.(string))
		return latestCommit.(string), nil
	}
}

func currentBranch(gv *gitview.GitView) string {
	branchName, _ := gv.BranchName()
	return branchName
}

func asBranchParameter(branchName string) string {
	if branchName != "" {
		return fmt.Sprintf("?branch=%s", branchName)
	}
	return ""
}
