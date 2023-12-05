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

type attestArtifactOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	gitReference       string
	srcRepoRoot        string
	displayName        string
	payload            AttestArtifactPayload
}

type AttestArtifactPayload struct {
	Fingerprint string                `json:"fingerprint"`
	Filename    string                `json:"filename"`
	GitCommit   string                `json:"git_commit"`
	BuildUrl    string                `json:"build_url"`
	CommitUrl   string                `json:"commit_url"`
	RepoUrl     string                `json:"repo_url"`
	CommitsList []*gitview.CommitInfo `json:"commits_list"`
	Name        string                `json:"step_name"`
	TrailName   string                `json:"trail_name"`
}

const attestArtifactShortDesc = `Attest an artifact creation to a Kosli flow.  `

const attestArtifactLongDesc = attestArtifactShortDesc + `
` + fingerprintDesc

const attestArtifactExample = `
# Attest to a Kosli flow that a file type artifact has been created
kosli attest artifact FILE.tgz \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName


# Attest to a Kosli flow that an artifact with a provided fingerprint (sha256) has been created
kosli attest artifact ANOTHER_FILE.txt \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint \
	--trail yourTrailName \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName
`

func newAttestArtifactCmd(out io.Writer) *cobra.Command {
	o := new(attestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "artifact {IMAGE-NAME | FILE-PATH | DIR-PATH}",
		Short:   attestArtifactShortDesc,
		Long:    attestArtifactLongDesc,
		Example: attestArtifactExample,
		Args:    cobra.MaximumNArgs(1),
		Hidden:  true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Fingerprint, true)
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
	cmd.Flags().StringVarP(&o.payload.Fingerprint, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.gitReference, "git-commit", "g", DefaultValue(ci, "git-commit"), gitCommitFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.CommitUrl, "commit-url", "u", DefaultValue(ci, "commit-url"), commitUrlFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	cmd.Flags().StringVarP(&o.payload.Name, "name", "n", "", templateArtifactName)
	cmd.Flags().StringVarP(&o.displayName, "display-name", "N", "", artifactDisplayName)
	cmd.Flags().StringVarP(&o.payload.TrailName, "trail", "T", "", trailNameFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"trail", "flow", "name", "git-commit", "build-url", "commit-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestArtifactOptions) run(args []string) error {

	if o.displayName != "" {
		o.payload.Filename = o.displayName
	} else {
		if o.fingerprintOptions.artifactType == "dir" || o.fingerprintOptions.artifactType == "file" {
			o.payload.Filename = filepath.Base(args[0])
		} else {
			o.payload.Filename = args[0]
		}
	}

	if o.payload.Fingerprint == "" {
		var err error
		o.payload.Fingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	gitView, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return err
	}

	commitObject, err := gitView.GetCommitInfoFromCommitSHA(o.gitReference)
	if err != nil {
		return err
	}
	o.payload.GitCommit = commitObject.Sha1

	previousCommit, err := o.latestCommit(currentBranch(gitView))
	if err == nil {
		o.payload.CommitsList, err = gitView.ChangeLog(o.payload.GitCommit, previousCommit, logger)
		if err != nil && !global.DryRun {
			return err
		}
	} else if !global.DryRun {
		return err
	}

	o.payload.RepoUrl, err = gitView.RepoUrl()
	if err != nil {
		logger.Warning("Repo URL will not be reported, %s", err.Error())
	}

	url := fmt.Sprintf("%s/api/v2/artifacts/%s/%s", global.Host, global.Org, o.flowName)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("artifact %s was attested with fingerprint: %s", o.payload.Filename, o.payload.Fingerprint)
	}
	return err
}

// latestCommit retrieves the git commit of the latest artifact for a flow in Kosli
func (o *attestArtifactOptions) latestCommit(branchName string) (string, error) {
	latestCommitUrl := fmt.Sprintf(
		"%s/api/v2/artifacts/%s/%s/%s/latest_commit%s",
		global.Host, global.Org, o.flowName, o.payload.Fingerprint, asBranchParameter(branchName))

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
		logger.Debug("no previous artifacts were found for flow: %s", o.flowName)
		return "", nil
	} else {
		logger.Debug("latest artifact for flow: %s has the git commit: %s", o.flowName, latestCommit.(string))
		return latestCommit.(string), nil
	}
}
