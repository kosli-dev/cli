package main

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type attestArtifactOptions struct {
	fingerprintOptions   *fingerprintOptions
	flowName             string
	gitReference         string
	srcRepoRoot          string
	displayName          string
	payload              AttestArtifactPayload
	externalFingerprints map[string]string
	externalURLs         map[string]string
}

type AttestArtifactPayload struct {
	Fingerprint   string                   `json:"fingerprint"`
	Filename      string                   `json:"filename"`
	GitCommit     string                   `json:"git_commit"`
	GitCommitInfo *gitview.BasicCommitInfo `json:"git_commit_info"`
	BuildUrl      string                   `json:"build_url"`
	CommitUrl     string                   `json:"commit_url"`
	RepoUrl       string                   `json:"repo_url"`
	Name          string                   `json:"template_reference_name"`
	TrailName     string                   `json:"trail_name"`
	ExternalURLs  map[string]*URLInfo      `json:"external_urls,omitempty"`
}

const attestArtifactShortDesc = `Attest an artifact creation to a Kosli flow.  `

const attestArtifactLongDesc = attestArtifactShortDesc + `
` + fingerprintDesc

const attestArtifactExample = `
# Attest that a file type artifact has been created, and let Kosli calculate its fingerprint
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


# Attest that an artifact has been created and provide its fingerprint (sha256) 
kosli attest artifact ANOTHER_FILE.txt \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--fingerprint yourArtifactFingerprint \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName

# Attest that an artifact has been created and provide external attachments
kosli attest artifact ANOTHER_FILE.txt \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--fingerprint yourArtifactFingerprint \
	--external-url label=https://example.com/attachment \
	--external-fingerprint label=yourExternalAttachmentFingerprint \
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
	cmd.Flags().StringVarP(&o.gitReference, "commit", "g", DefaultValueForCommit(ci, true), gitCommitFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.CommitUrl, "commit-url", "u", DefaultValue(ci, "commit-url"), commitUrlFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	cmd.Flags().StringVarP(&o.payload.Name, "name", "n", "", templateArtifactName)
	cmd.Flags().StringVarP(&o.displayName, "display-name", "N", "", artifactDisplayName)
	cmd.Flags().StringVarP(&o.payload.TrailName, "trail", "T", "", trailNameFlag)
	cmd.Flags().StringToStringVar(&o.externalFingerprints, "external-fingerprint", map[string]string{}, externalFingerprintFlag)
	cmd.Flags().StringToStringVar(&o.externalURLs, "external-url", map[string]string{}, externalURLFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"trail", "flow", "name", "build-url", "commit-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attestArtifactOptions) run(args []string) error {
	var err error
	if o.displayName != "" {
		o.payload.Filename = o.displayName
	} else {
		if o.fingerprintOptions.artifactType == "dir" || o.fingerprintOptions.artifactType == "file" {
			o.payload.Filename = filepath.Base(args[0])
		} else {
			o.payload.Filename = args[0]
		}
	}

	// process external urls
	o.payload.ExternalURLs, err = processExternalURLs(o.externalURLs, o.externalFingerprints)
	if err != nil {
		return err
	}

	if o.payload.Fingerprint == "" {
		o.payload.Fingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	gitView, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return err
	}

	commitInfo, err := gitView.GetCommitInfoFromCommitSHA(o.gitReference, false)
	if err != nil {
		return err
	}
	o.payload.GitCommit = commitInfo.Sha1
	o.payload.GitCommitInfo = &commitInfo.BasicCommitInfo

	o.payload.RepoUrl, err = gitView.RepoURL()
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
