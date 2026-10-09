package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/kosli-dev/cli/internal/buildexec"
	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type attestArtifactOptions struct {
	fingerprintOptions   *fingerprintOptions
	flowName             string
	gitReference         string
	redactedCommitInfo   []string
	srcRepoRoot          string
	displayName          string
	payload              AttestArtifactPayload
	externalFingerprints map[string]string
	externalURLs         map[string]string
	annotations          map[string]string
	repoID               string
	repoName             string
	repoURL              string
	repoProvider         string
	repoNameExplicit     bool
	buildCmd             []string
	recordBuildCommand   bool
}

type AttestArtifactPayload struct {
	Fingerprint   string                   `json:"fingerprint"`
	Filename      string                   `json:"filename"`
	GitCommit     string                   `json:"git_commit"`
	GitCommitInfo *gitview.BasicCommitInfo `json:"git_commit_info"`
	GitRepoInfo   *gitview.GitRepoInfo     `json:"repo_info,omitempty"`
	BuildUrl      string                   `json:"build_url"`
	CommitUrl     string                   `json:"commit_url"`
	RepoUrl       string                   `json:"repo_url"`
	Name          string                   `json:"template_reference_name"`
	TrailName     string                   `json:"trail_name"`
	ExternalURLs  map[string]*URLInfo      `json:"external_urls,omitempty"`
	Annotations   map[string]string        `json:"annotations,omitempty"`
}

const attestArtifactShortDesc = `Attest an artifact creation to a Kosli flow.  `

const attestArtifactLongDesc = attestArtifactShortDesc + `
` + fingerprintDesc + kosliIgnoreDesc + `
This command requires access to a git repo to associate the artifact to the git commit it is originating from.
You can optionally redact some of the git commit data sent to Kosli using ^--redact-commit-info^.
To record repository information, all three of ^--repo-id^, ^--repo-url^, and ^--repository^ must be set together.
These are automatically set in GitHub Actions, GitLab CI, Bitbucket Pipelines, and Azure DevOps.
In other CI systems, set them explicitly to capture repository metadata.

Everything after ^--^ runs as a build command, and the artifact is attested only if the build succeeds.
Its fingerprint is calculated after the build. The build command (with secret values masked) and its duration
are recorded in the ^build_command^ and ^build_duration_seconds^ annotations; use ^--record-build-command=false^
to leave out the command. If the build fails, nothing is attested and Kosli exits with the build's exit code.`

const attestArtifactExample = `
# Attest that a file type artifact has been created, and let Kosli calculate its fingerprint
kosli attest artifact FILE.tgz \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName


# Attest that an artifact has been created and provide its fingerprint (sha256) 
kosli attest artifact ANOTHER_FILE.txt \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--commit yourCommitShaThatThisArtifactWasBuiltFrom \
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
	--commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--fingerprint yourArtifactFingerprint \
	--external-url label=https://example.com/attachment \
	--external-fingerprint label=yourExternalAttachmentFingerprint \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName

# Build an artifact and attest it in one step (the attestation only happens if the build succeeds)
kosli attest artifact dist/app \
	--artifact-type file \
	--name yourTemplateArtifactName \
	--flow yourFlowName \
	--trail yourTrailName \
	--api-token yourApiToken \
	--org yourOrgName \
	-- go build -o dist/app ./cmd/app
`

func newAttestArtifactCmd(out io.Writer) *cobra.Command {
	o := new(attestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		//Args:    cobra.MaximumNArgs(1), // See CustomMaximumNArgs() below
		Use:     "artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [-- BUILD-COMMAND...]",
		Short:   attestArtifactShortDesc,
		Long:    attestArtifactLongDesc,
		Example: attestArtifactExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			args, buildCmd, err := splitArtifactAndBuildArgs(cmd.ArgsLenAtDash(), args)
			if err != nil {
				return err
			}
			o.buildCmd = buildCmd

			err = CustomMaximumNArgs(1, args)
			if err != nil {
				return err
			}

			err = validateBuildCommand(o.buildCmd, o.payload.Fingerprint)
			if err != nil {
				return err
			}

			err = RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateSliceValues(o.redactedCommitInfo, allowedCommitRedactionValues)
			if err != nil {
				return fmt.Errorf("%s for --redact-commit-info", err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Fingerprint, true)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			if err := validateRepoFlags(o.repoURL, o.repoProvider, cmd.Flags().Changed("repo-url")); err != nil {
				return err
			}

			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.repoNameExplicit = cmd.Flags().Changed("repository")
			args = args[:len(args)-len(o.buildCmd)]
			if len(o.buildCmd) > 0 {
				if err := o.runBuild(args[0], cmd.InOrStdin(), out, cmd.ErrOrStderr()); err != nil {
					return err
				}
			}
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.payload.Fingerprint, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.gitReference, "commit", "g", DefaultValueForCommit(ci, true), gitCommitFlag)
	cmd.Flags().StringSliceVar(&o.redactedCommitInfo, "redact-commit-info", []string{}, attestationRedactCommitInfoFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.CommitUrl, "commit-url", "u", DefaultValue(ci, "commit-url"), commitUrlFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	cmd.Flags().StringVarP(&o.payload.Name, "name", "n", "", templateArtifactName)
	cmd.Flags().StringVarP(&o.displayName, "display-name", "N", "", artifactDisplayName)
	cmd.Flags().StringVarP(&o.payload.TrailName, "trail", "T", "", trailNameFlag)
	cmd.Flags().StringToStringVar(&o.externalFingerprints, "external-fingerprint", map[string]string{}, externalFingerprintFlag)
	cmd.Flags().StringToStringVar(&o.externalURLs, "external-url", map[string]string{}, externalURLFlag)
	cmd.Flags().StringToStringVar(&o.annotations, "annotate", map[string]string{}, annotationFlag)
	cmd.Flags().BoolVar(&o.recordBuildCommand, "record-build-command", true, recordBuildCommandFlag)
	cmd.Flags().StringVar(&o.repoID, "repo-id", DefaultValue(ci, "repo-id"), repoIDFlag)
	cmd.Flags().StringVar(&o.repoName, "repository", DefaultValue(ci, "repository"), attestationRepoNameFlag)
	cmd.Flags().StringVar(&o.repoURL, "repo-url", DefaultValue(ci, "repo-url"), repoURLFlag)
	cmd.Flags().StringVar(&o.repoProvider, "repo-provider", DefaultValue(ci, "repo-provider"), repoProviderFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"trail", "flow", "name", "build-url", "commit-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

// splitArtifactAndBuildArgs splits the positional args at "--" (dash is
// cobra's ArgsLenAtDash) into the artifact args and the build command.
func splitArtifactAndBuildArgs(dash int, args []string) (artifactArgs, buildCmd []string, err error) {
	if dash < 0 {
		return args, nil, nil
	}
	// "-- NAME" predates build commands and lets an artifact name start with "-".
	if dash == 0 && len(args) == 1 {
		return args, nil, nil
	}
	if dash == len(args) {
		return nil, nil, fmt.Errorf("no build command given after --")
	}
	return args[:dash], args[dash:], nil
}

func validateBuildCommand(buildCmd []string, fingerprint string) error {
	if len(buildCmd) == 0 {
		return nil
	}
	if fingerprint != "" {
		return fmt.Errorf("--fingerprint cannot be combined with a build command")
	}
	// Multi-host mode re-runs the command per host and appends host flags after "--".
	if hasMultipleHosts(strings.Split(global.Host, ","), strings.Split(global.ApiToken, ",")) {
		return fmt.Errorf("a build command is not supported with multiple hosts yet")
	}
	return nil
}

// runBuild runs the build command, then fingerprints the artifact it built and
// records the build in the attestation's annotations.
func (o *attestArtifactOptions) runBuild(artifactName string, stdin io.Reader, stdout, stderr io.Writer) error {
	var before string
	if o.fingerprintOptions.artifactType == "file" || o.fingerprintOptions.artifactType == "dir" {
		before, _ = GetSha256Digest(artifactName, o.fingerprintOptions, logger)
	}

	result, err := buildexec.Run(context.Background(), o.buildCmd, []string{"KOSLI_SHIM_DISABLED=1"}, stdin, stdout, stderr)
	if err != nil {
		return &buildFailedError{code: result.ExitCode, err: err}
	}

	o.payload.Fingerprint, err = GetSha256Digest(artifactName, o.fingerprintOptions, logger)
	if err != nil {
		return fmt.Errorf("build command succeeded but artifact %s was not found: %w", artifactName, err)
	}
	// A warning only: reproducible builds legitimately rewrite identical bytes.
	if before != "" && before == o.payload.Fingerprint {
		logger.Warn("artifact %s was not modified by the build command", artifactName)
	}

	cmdline := maskSecrets(strings.Join(o.buildCmd, " "), os.Environ(), os.Getenv("KOSLI_API_TOKEN"), global.ApiToken)
	o.annotations = addBuildAnnotations(o.annotations, cmdline, result.Duration, o.recordBuildCommand)
	return nil
}

var secretEnvNameRegexp = regexp.MustCompile(`(?i)(TOKEN|SECRET|PASSWORD|PASSWD|KEY|CREDENTIAL)`)

// maskSecrets replaces, in cmdline, the values of environment variables whose
// names look secret, and every extra value, with "***".
func maskSecrets(cmdline string, environ []string, extra ...string) string {
	secrets := slices.Clone(extra)
	for _, entry := range environ {
		name, value, _ := strings.Cut(entry, "=")
		if len(value) >= 6 && secretEnvNameRegexp.MatchString(name) {
			secrets = append(secrets, value)
		}
	}
	// Longest first, so a secret that contains another is masked whole.
	slices.SortFunc(secrets, func(a, b string) int { return len(b) - len(a) })
	for _, secret := range secrets {
		if secret != "" {
			cmdline = strings.ReplaceAll(cmdline, secret, "***")
		}
	}
	return cmdline
}

// addBuildAnnotations adds the build command and duration to annotations,
// keeping any value the user set for the same key.
func addBuildAnnotations(annotations map[string]string, cmdline string, duration time.Duration, recordCommand bool) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}
	derived := map[string]string{"build_duration_seconds": strconv.FormatFloat(duration.Seconds(), 'f', 1, 64)}
	if recordCommand {
		derived["build_command"] = cmdline
	}
	for key, value := range derived {
		if _, ok := annotations[key]; !ok {
			annotations[key] = value
		}
	}
	return annotations
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

	o.payload.Annotations, err = processAnnotations(o.annotations)
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

	commitInfo, err := gitView.GetCommitInfoFromCommitSHA(o.gitReference, false, o.redactedCommitInfo)
	if err != nil {
		return err
	}
	o.payload.GitRepoInfo, err = getGitRepoInfoFromEnvironment()
	if err != nil {
		logger.Warn("failed to get git repo info. %s", err.Error())
	}
	o.payload.GitRepoInfo = mergeGitRepoInfo(o.payload.GitRepoInfo, o.repoID, o.repoName, o.repoURL, o.repoProvider, o.repoNameExplicit)
	o.payload.GitCommit = commitInfo.Sha1
	o.payload.GitCommitInfo = &commitInfo.BasicCommitInfo

	o.payload.RepoUrl, err = gitView.RepoURL()
	if err != nil {
		logger.Warn("Repo URL will not be reported, %s", err.Error())
	}

	url, err := url.JoinPath(global.Host, "api/v2/artifacts", global.Org, o.flowName)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPost,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("artifact %s was attested with fingerprint: %s", o.payload.Filename, o.payload.Fingerprint)
	}
	return err
}
