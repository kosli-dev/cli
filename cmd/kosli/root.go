package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/security"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var globalUsage = `The Kosli CLI.

Environment variables:
You can set any flag from an environment variable by capitalizing it in snake case and adding the KOSLI_ prefix.
For example, to set --api-token from an environment variable, you can export KOSLI_API_TOKEN=YOUR_API_TOKEN.

Setting the API token to DRY_RUN sets the --dry-run flag.
`

const (
	defaultMaxAPIRetries = 3
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = ".kosli.yml"

	// The default Kosli app host URL.
	defaultHost = "https://app.kosli.com"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --namespace is bound to KOSLI_NAMESPACE.
	envPrefix = "KOSLI"

	// the name of the credentials store secret holding the encryption key for api token storage
	credentialsStoreKeySecretName = "kosli-encryption-key"

	// the following constants are used in the docs/help
	fingerprintDesc = `
The artifact fingerprint can be provided directly with the ^--fingerprint^ flag, or 
calculated based on ^--artifact-type^ flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.

`

	attestationBindingDesc = `

The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the ^--artifact-type^ flag and the artifact name/path argument) or can be provided directly (with the ^--fingerprint^ flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.`
	awsAuthDesc = `

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	`
	azureAuthDesc = `

To authenticate to Azure, you need to create Azure service principal with a secret  
and provide these Azure credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AZURE_CLIENT_ID).  
The service principal needs to have the following permissions:  
  1) Microsoft.Web/sites/Read  
  2) Microsoft.ContainerRegistry/registries/pull/read  

	`
	kosliIgnoreDesc = `To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a ^.kosli_ignore^ file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using ^#^.
The ^.kosli_ignore^ will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.`

	// flags
	apiTokenFlag                         = "The Kosli API token."
	artifactName                         = "[optional] Artifact display name, if different from file, image or directory name."
	artifactDisplayName                  = "[optional] Artifact display name, if different from file, image or directory name."
	orgFlag                              = "The Kosli organization."
	hostFlag                             = "[defaulted] The Kosli endpoint."
	httpProxyFlag                        = "[optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'"
	dryRunFlag                           = "[optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors."
	maxAPIRetryFlag                      = "[defaulted] How many times should API calls be retried when the API host is not reachable."
	configFileFlag                       = "[optional] The Kosli config file path."
	debugFlag                            = "[optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)"
	artifactTypeFlag                     = "The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it)."
	flowNameFlag                         = "The Kosli flow name."
	trailNameFlag                        = "The Kosli trail name."
	trailNameFlagOptional                = "[optional] The Kosli trail name."
	templateArtifactName                 = "The name of the artifact in the yml template file."
	flowNamesFlag                        = "[defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org."
	outputFlag                           = "[defaulted] The format of the output. Valid formats are: [table, json]."
	environmentNameFlag                  = "The environment name."
	approvalEnvironmentNameFlag          = "[defaulted] The environment the artifact is approved for. (defaults to all environments)"
	pageNumberFlag                       = "[defaulted] The page number of a response."
	pageLimitFlag                        = "[defaulted] The number of elements per page."
	pageLimitListTrailsFlag              = "The number of elements per page."
	newEnvTypeFlag                       = "The type of environment. Valid types are: [K8S, ECS, server, S3, lambda, docker, azure-apps, logical]."
	envAllowListFlag                     = "The environment name for which the artifact is allowlisted."
	reasonFlag                           = "The reason why this artifact is allowlisted."
	oldestCommitFlag                     = "[conditional] The source commit sha for the oldest change in the deployment. Can be any commit-ish. Only required if you don't specify '--environment'."
	newestCommitFlag                     = "[defaulted] The source commit sha for the newest change in the deployment. Can be any commit-ish."
	repoRootFlag                         = "[defaulted] The directory where the source git repository is available."
	approvalDescriptionFlag              = "[optional] The approval description."
	deploymentDescriptionFlag            = "[optional] The deployment description."
	evidenceDescriptionFlag              = "[optional] The evidence description."
	jiraBaseUrlFlag                      = "The base url for the jira project, e.g. 'https://kosli.atlassian.net'"
	jiraUsernameFlag                     = "Jira username (for Jira Cloud)"
	jiraAPITokenFlag                     = "Jira API token (for Jira Cloud)"
	jiraPATFlag                          = "Jira personal access token (for self-hosted Jira)"
	jiraIssueFieldFlag                   = "[optional] The comma separated list of fields to include from the Jira issue. Default no fields are included. '*all' will give all fields."
	jiraSecondarySourceFlag              = "[optional] An optional string to search for Jira ticket reference, e.g. '--jira-secondary-source ${{ github.head_ref }}'"
	ignoreBranchMatchFlag                = "Ignore branch name when searching for Jira ticket reference."
	envDescriptionFlag                   = "[optional] The environment description."
	flowDescriptionFlag                  = "[optional] The Kosli flow description."
	trailDescriptionFlag                 = "[optional] The Kosli trail description."
	visibilityFlag                       = "[defaulted] The visibility of the Kosli flow. Valid visibilities are [public, private]."
	templateFlag                         = "[defaulted] The comma-separated list of required compliance controls names."
	templateFileFlag                     = "[optional] The path to a yaml template file. Cannot be used together with --use-empty-template"
	templateFileSimpleFlag               = "[optional] The path to a yaml template file."
	useEmptyTemplateFlag                 = "Use an empty template for the flow creation without specifying a file. Cannot be used together with --template or --template-file"
	approvalUserDataFlag                 = "[optional] The path to a JSON file containing additional data you would like to attach to the approval."
	evidenceUserDataFlag                 = "[optional] The path to a JSON file containing additional data you would like to attach to the evidence."
	attestationUserDataFlag              = "[optional] The path to a JSON file containing additional data you would like to attach to the attestation."
	deploymentUserDataFlag               = "[optional] The path to a JSON file containing additional data you would like to attach to the deployment."
	trailUserDataFlag                    = "[optional] The path to a JSON file containing additional data you would like to attach to the flow trail."
	gitCommitFlag                        = "[defaulted] The git commit from which the artifact was created. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD )."
	evidenceBuildUrlFlag                 = "The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	buildUrlFlag                         = "The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	commitUrlFlag                        = "The url for the git commit that created the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	evidenceCompliantFlag                = "[defaulted] Whether the evidence is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags"
	bbUsernameFlag                       = "Bitbucket username. Only needed if you use --bitbucket-password"
	bbPasswordFlag                       = "Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details."
	bbAccessTokenFlag                    = "Bitbucket repo/project/workspace access token. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#access-tokens for more details."
	bbWorkspaceFlag                      = "Bitbucket workspace ID."
	commitPREvidenceFlag                 = "Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	commitEvidenceFlag                   = "Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	repositoryFlag                       = "Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	assertPREvidenceFlag                 = "[optional] Exit with non-zero code if no pull requests found for the given commit."
	assertJiraEvidenceFlag               = "[optional] Exit with non-zero code if no jira issue reference found, or jira issue does not exist, for the given commit or branch."
	assertStatusFlag                     = "[optional] Exit with non-zero code if Kosli server is not responding."
	azureTokenFlag                       = "Azure Personal Access token."
	azureProjectFlag                     = "Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults )."
	azureOrgUrlFlag                      = "Azure organization url. E.g. \"https://dev.azure.com/myOrg\" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults )."
	azureClientIdFlag                    = "Azure client ID."
	azureClientSecretFlag                = "Azure client secret."
	azureTenantIdFlag                    = "Azure tenant ID."
	azureSubscriptionIdFlag              = "Azure subscription ID."
	azureResourceGroupNameFlag           = "Azure resource group name."
	azureDigestsSourceFlag               = "[defaulted] Where to get the digests from. Valid values are 'acr' and 'logs'."
	githubTokenFlag                      = "Github token."
	githubOrgFlag                        = "Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults )."
	githubBaseURLFlag                    = "[optional] GitHub base URL (only needed for GitHub Enterprise installations)."
	gitlabTokenFlag                      = "Gitlab token."
	gitlabOrgFlag                        = "Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults )."
	gitlabBaseURLFlag                    = "[optional] Gitlab base URL (only needed for on-prem Gitlab installations)."
	registryProviderFlag                 = "[deprecated] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry."
	registryUsernameFlag                 = "[conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry."
	registryPasswordFlag                 = "[conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry."
	resultsDirFlag                       = "[defaulted] The path to a directory with JUnit test results. By default, the directory will be uploaded to Kosli's evidence vault."
	snykJsonResultsFileFlag              = "The path to Snyk SARIF or JSON scan results file from 'snyk test' and 'snyk container test'. By default, the Snyk results will be uploaded to Kosli's evidence vault."
	snykSarifResultsFileFlag             = "The path to Snyk scan SARIF results file from 'snyk test' and 'snyk container test'. By default, the Snyk results will be uploaded to Kosli's evidence vault."
	ecsClusterFlag                       = "The name of the ECS cluster."
	ecsClustersFlag                      = "[optional] The comma-separated list of ECS cluster names to snapshot. Can't be used together with --exclude or --exclude-regex."
	ecsClustersRegexFlag                 = "[optional] The comma-separated list of ECS cluster name regex patterns to snapshot. Can't be used together with --exclude or --exclude-regex."
	ecsExcludeClustersFlag               = "[optional] The comma-separated list of ECS cluster names to exclude. Can't be used together with --exclude or --exclude-regex."
	ecsExcludeClustersRegexFlag          = "[optional] The comma-separated list of ECS cluster name regex patterns to exclude. Can't be used together with --clusters or --clusters-regex."
	ecsServiceFlag                       = "[optional] The name of the ECS service."
	kubeconfigFlag                       = "[defaulted] The kubeconfig path for the target cluster."
	namespacesFlag                       = "[optional] The comma separated list of namespaces names to report artifacts info from. Can't be used together with --exclude-namespaces or --exclude-namespaces-regex."
	excludeNamespacesFlag                = "[optional] The comma separated list of namespaces names to exclude from reporting artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --namespaces or --namespaces-regex."
	namespacesRegexFlag                  = "[optional] The comma separated list of namespaces regex patterns to report artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --exclude-namespaces --exclude-namespaces-regex."
	excludeNamespacesRegexFlag           = "[optional] The comma separated list of namespaces regex patterns to exclude from reporting artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --namespaces or --namespaces-regex."
	functionNameFlag                     = "[optional] The name of the AWS Lambda function."
	functionNamesFlag                    = "[optional] The comma-separated list of AWS Lambda function names to be reported. Cannot be used together with --exclude or --exclude-regex."
	functionNamesRegexFlag               = "[optional] The comma-separated list of AWS Lambda function names regex patterns to be reported. Cannot be used together with --exclude or --exclude-regex."
	excludeFlag                          = "[optional] The comma-separated list of AWS Lambda function names to be excluded. Cannot be used together with --function-names"
	excludeRegexFlag                     = "[optional] The comma-separated list of name regex patterns for AWS Lambda functions to be excluded. Cannot be used together with --function-names. Allowed regex patterns are described in https://github.com/google/re2/wiki/Syntax"
	functionVersionFlag                  = "[optional] The version of the AWS Lambda function."
	awsKeyIdFlag                         = "The AWS access key ID."
	awsSecretKeyFlag                     = "The AWS secret access key."
	awsRegionFlag                        = "The AWS region."
	bucketNameFlag                       = "The name of the S3 bucket."
	bucketPathsFlag                      = "[optional] The comma separated list of file and/or directory paths in the S3 bucket to include when fingerprinting. Cannot be used together with --exclude."
	excludeBucketPathsFlag               = "[optional] The comma separated list of file and/or directory paths in the S3 bucket to exclude when fingerprinting. Cannot be used together with --include."
	pathsFlag                            = "The comma separated list of absolute or relative paths of artifact directories or files. Can take glob patterns, but be aware that each matching path will be reported as an artifact."
	excludePathsFlag                     = "[optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir."
	serverExcludePathsFlag               = "[optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns."
	shortFlag                            = "[optional] Print only the Kosli CLI version number."
	reverseFlag                          = "[defaulted] Reverse the order of output list."
	evidenceNameFlag                     = "The name of the evidence."
	evidenceFingerprintFlag              = "[optional] The SHA256 fingerprint of the evidence file or dir."
	evidenceURLFlag                      = "[optional] The external URL where the evidence file or dir is stored."
	evidencePathsFlag                    = "[optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault."
	fingerprintFlag                      = "[conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'."
	intervalFlag                         = "[optional] Expression to define specified snapshots range."
	showUnchangedArtifactsFlag           = "[defaulted] Show the unchanged artifacts present in both snapshots within the diff output."
	approverFlag                         = "[optional] The user approving an approval."
	attestationFingerprintFlag           = "[conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used."
	attestationCommitFlag                = "[conditional] The git commit for which the attestation is associated to. Becomes required when reporting an attestation for an artifact before reporting it to Kosli. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	attestationRedactCommitInfoFlag      = "[optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch]."
	attestationOriginUrlFlag             = "[optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults )."
	attestationNameFlag                  = "The name of the attestation as declared in the flow or trail yaml template."
	attestationCompliantFlag             = "[defaulted] Whether the attestation is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags"
	attestationRepoRootFlag              = "[defaulted] The directory where the source git repository is available. Only used if --commit is used."
	attestationCustomTypeNameFlag        = "The name of the custom attestation type."
	attestationCustomDataFileFlag        = "The filepath of a json file containing the custom attestation data."
	uploadJunitResultsFlag               = "[defaulted] Whether to upload the provided Junit results directory as an attachment to Kosli or not."
	uploadSnykResultsFlag                = "[defaulted] Whether to upload the provided Snyk results file as an attachment to Kosli or not."
	attestationAssertFlag                = "[optional] Exit with non-zero code if the attestation is non-compliant"
	beginTrailCommitFlag                 = "[defaulted] The git commit from which the trail is begun. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD )."
	attachmentsFlag                      = "[optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault."
	externalFingerprintFlag              = "[optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint."
	externalURLFlag                      = "[optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint"
	annotationFlag                       = "[optional] Annotate the attestation with data using key=value."
	attestationDescription               = "[optional] attestation description"
	excludeScalingFlag                   = "[optional] Exclude scaling events for snapshots. Snapshots with scaling changes will not result in new environment records."
	includeScalingFlag                   = "[optional] Include scaling events for snapshots. Snapshots with scaling changes will result in new environment records."
	includedEnvironments                 = "[optional] Comma separated list of environments to include in logical environment"
	requireProvenanceFlag                = "[defaulted] Require provenance for all artifacts running in environment snapshots."
	deprecatedKosliReportEvidenceMessage = "See **kosli attest** commands."
	setTagsFlag                          = "[optional] The key-value pairs to tag the resource with. The format is: key=value"
	unsetTagsFlag                        = "[optional] The list of tag keys to remove from the resource."
	pathsSpecFileFlag                    = "The path to a paths file in YAML/JSON/TOML format. Cannot be used together with --path ."
	snapshotPathPathFlag                 = "The base path for the artifact to snapshot."
	snapshotPathExcludeFlag              = "[optional] The comma-separated list of literal paths or glob patterns to exclude when fingerprinting the artifact."
	snapshotPathArtifactNameFlag         = "The reported name of the artifact."
	policyDescriptionFlag                = "[optional] policy description."
	policyCommentFlag                    = "[optional] comment about the change made in a policy file when updating a policy."
	policyTypeFlag                       = "[defaulted] the type of policy. One of: [env]"
	attachPolicyEnvFlag                  = "the list of environment names to attach the policy to"
	detachPolicyEnvFlag                  = "the list of environment names to detach the policy from"
	sonarAPITokenFlag                    = "[required] SonarCloud/SonarQube API token."
	sonarWorkingDirFlag                  = "[conditional] The base directory of the repo scanned by SonarCloud/SonarQube. Only required if you have overriden the default in the sonar scanner or you are running the CLI locally in a separate folder from the repo."
	sonarProjectKeyFlag                  = "[conditional] The project key of the SonarCloud/SonarQube project. Only required if you want to use the project key/revision to get the scan results rather than using Sonar's metadata file."
	sonarServerURLFlag                   = "[conditional] The URL of your SonarQube server. Only required if you are using SonarQube and not using SonarQube's metadata file to get scan results."
	sonarRevisionFlag                    = "[conditional] The revision of the SonarCloud/SonarQube project. Only required if you want to use the project key/revision to get the scan results rather than using Sonar's metadata file and you have overridden the default revision, or you aren't using a CI. Defaults to the value of the git commit flag."
	logicalEnvFlag                       = "[required] The logical environment."
	physicalEnvFlag                      = "[required] The physical environment."
	attestationTypeDescriptionFlag       = "[optional] The attestation type description."
	attestationTypeSchemaFlag            = "[optional] Path to the attestation type schema in JSON Schema format."
	attestationTypeJqFlag                = "[optional] The attestation type evaluation JQ rules."
	envNameFlag                          = "The Kosli environment name to assert the artifact against."
	pathsWatchFlag                       = "[optional] Watch the filesystem for changes and report snapshots of artifacts running in specific filesystem paths to Kosli."
)

var global *GlobalOpts

type GlobalOpts struct {
	ApiToken      string
	Org           string
	Host          string
	HttpProxy     string
	DryRun        bool
	MaxAPIRetries int
	ConfigFile    string
	Debug         bool
}

// ConfigGetter defines an interface for getting the default config file path
// the interface allows to mock the default config file path in tests
type ConfigGetter interface {
	defaultConfigFilePath() string
}

// RealConfigGetter is a real implementation of the ConfigGetter interface
type RealConfigGetter struct{}

// defaultConfigFilePath is a method that satisfies the ConfigGetter interface
func (r *RealConfigGetter) defaultConfigFilePath() string {
	if _, ok := os.LookupEnv("DOCS"); ok { // used for docs generation
		return fmt.Sprintf("$HOME/%s", defaultConfigFilename)
	}

	home, err := homedir.Dir()
	if err == nil {
		return filepath.Join(home, defaultConfigFilename)

	}
	return "kosli" // for backward compatibility with old default config location
}

// defaultConfigFilePathFunc is a variable holding the implementation of defaultConfigFilePath
var defaultConfigFilePathFunc = (&RealConfigGetter{}).defaultConfigFilePath

func getConfigFileFlagDefault() string {
	defaultPath := defaultConfigFilePathFunc()
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	return "kosli" // for backward compatibility with old default config location
}

func newRootCmd(out, errOut io.Writer, args []string) (*cobra.Command, error) {
	global = new(GlobalOpts)
	cmd := &cobra.Command{
		Use:              "kosli",
		Short:            "The Kosli CLI.",
		Long:             globalUsage,
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			err := initialize(cmd, out, errOut)
			if err != nil {
				return err
			}

			if global.ApiToken == "DRY_RUN" {
				global.DryRun = true
			}

			// If the user types "--description $variable --sha256 ..." and $variable is "" then Cobra
			// will assign --sha256 as the value of --description, and give a very misleading error message.
			// So we do some extra checking to tell the user about this.
			var flagError error = nil
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if strings.HasPrefix(f.Value.String(), "-") {
					flagError = fmt.Errorf("flag '--%s' has value '%s' which is illegal", f.Name, f.Value.String())
				}

				if _, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
					if f.Changed && f.Value.String() == "" {
						flagError = fmt.Errorf("flag '--%s' is required, but empty string was provided", f.Name)
					}
				}
			})

			return flagError
		},
	}
	cmd.PersistentFlags().StringVarP(&global.ApiToken, "api-token", "a", "", apiTokenFlag)
	cmd.PersistentFlags().StringVar(&global.Org, "org", "", orgFlag)
	cmd.PersistentFlags().StringVarP(&global.Host, "host", "H", defaultHost, hostFlag)
	cmd.PersistentFlags().StringVar(&global.HttpProxy, "http-proxy", "", httpProxyFlag)
	cmd.PersistentFlags().IntVarP(&global.MaxAPIRetries, "max-api-retries", "r", defaultMaxAPIRetries, maxAPIRetryFlag)
	cmd.PersistentFlags().StringVarP(&global.ConfigFile, "config-file", "c", getConfigFileFlagDefault(), configFileFlag)
	cmd.PersistentFlags().BoolVar(&global.Debug, "debug", false, debugFlag)

	// Add subcommands
	cmd.AddCommand(

		newVersionCmd(out),
		newFingerprintCmd(out),
		newAssertCmd(out),
		newStatusCmd(out),
		newExpectCmd(out),
		newSearchCmd(out),
		newCompletionCmd(out),
		// Hidden documentation generator command: 'kosli docs'
		newDocsCmd(out),

		// New syntax commands
		newGetCmd(out),
		newCreateCmd(out),
		newBeginCmd(out),
		newAttestCmd(out),
		newReportCmd(out),
		newDiffCmd(out),
		newAllowCmd(out),
		newListCmd(out),
		newRenameCmd(out),
		newJoinCmd(out),
		newArchiveCmd(out),
		newSnapshotCmd(out),
		newRequestCmd(out),
		newLogCmd(out),
		newDisableCmd(out),
		newEnableCmd(out),
		newTagCmd(out),
		newConfigCmd(out),
		newAttachPolicyCmd(out),
		newDetachPolicyCmd(out),
	)

	cobra.AddTemplateFunc("isBeta", isBeta)
	cobra.AddTemplateFunc("isDeprecated", isDeprecated)
	cmd.SetUsageTemplate(usageTemplate)

	return cmd, nil
}

func initialize(cmd *cobra.Command, out, errOut io.Writer) error {
	v := viper.New()
	logger.SetInfoOut(out) // needed to allow tests to overwrite the logger output stream
	logger.SetErrOut(errOut)
	// assign debug value early here to enable debug logs during config file and env var binding
	// if --debug is used. The value is re-assigned later after binding config file and env vars
	logger.DebugEnabled = global.Debug

	// If provided, extract the custom config file dir and name

	// handle passing the config file as an env variable.
	// we load the config file before we bind env vars to flags,
	// so we check for the config file env var separately here
	configFlag := cmd.Flags().Lookup("config-file")
	if !configFlag.Changed {
		if path, exists := os.LookupEnv("KOSLI_CONFIG_FILE"); exists {
			global.ConfigFile = path
		}
	}
	dir, file := filepath.Split(global.ConfigFile)
	file = strings.TrimSuffix(file, filepath.Ext(file))

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(file)

	// Set as many paths as you like where viper should look for the
	// config file. By default, we are looking in the current working directory.
	if dir == "" {
		dir = "."
	}
	v.AddConfigPath(dir)

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	logger.Debug("processing config file [%s]", global.ConfigFile)
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to parse config file [%s] : %v", global.ConfigFile, err)
		} else {
			logger.Debug("config file [%s] not found. Skipping.", global.ConfigFile)
		}
	}
	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --namespace
	// binds to an environment variable KOSLI_NAMESPACE. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Bind viper config to environment variables
	// Works great for simple config names, but needs help for names
	// like --kube-config which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	// re-assign debug after binding flags to config or env vars as it may have
	// a different value now
	logger.DebugEnabled = global.Debug

	var err error
	kosliClient, err = requests.NewKosliClient(global.HttpProxy, global.MaxAPIRetries, global.Debug, logger)
	if err != nil {
		return err
	}

	return nil
}

// Bind each cobra flag to its associated viper configuration
// (coming either from environment variables or config file)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	// for some reason, logger does not print errors at the point
	// of calling this function, so we ensure to point errors to stderr
	// logger.SetErrOut(errOut)
	// api token in config file is encrypted, so we have to decrypt it
	// but if it is set via env variables, it is not encrypted
	_, apiTokenSetInEnv := os.LookupEnv("KOSLI_API_TOKEN")

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --kube-config to KOSLI_KUBE_CONFIG
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			if err := v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix)); err != nil {
				logger.Error("failed to bind viper to env variable: %v", err)
			}
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		// for api token, decrypt it if it is coming from the config file
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if !apiTokenSetInEnv && f.Name == "api-token" {
				// api token is coming from a config file (it may or may not be encrypted)
				// we try decrypting it first, if that fails, we use it as it is
				// get encryption key
				key, err := security.GetSecretFromCredentialsStore(credentialsStoreKeySecretName)
				if err != nil {
					logger.Warn("failed to decrypt api token from [%s]. Failed to get api token encryption key from credentials store: %s", global.ConfigFile, err)
					logger.Warn("using api token from [%s] as plain text. It is recommended to encrypt your api token by setting it with: kosli config --api-token <token>", global.ConfigFile)
				} else {
					// decrypt token
					decryptedBytes, err := security.AESDecrypt([]byte(val.(string)), []byte(key))
					if err != nil {
						logger.Warn("failed to decrypt api token from [%s]: %s", global.ConfigFile, err)
						logger.Warn("using api token from [%s] as plain text. It is recommended to encrypt your api token by setting it with: kosli config --api-token <token>", global.ConfigFile)
					} else {
						val = string(decryptedBytes)
						logger.Debug("using api token from [%s].", global.ConfigFile)
					}
				}
			}

			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				logger.Error("failed to set flag: %v", err)
			}
		}
	})
}

func isBeta(cmd *cobra.Command) bool {
	if _, ok := cmd.Annotations["betaCLI"]; ok {
		return true
	}
	var beta bool
	cmd.VisitParents(func(cmd *cobra.Command) {
		if _, ok := cmd.Annotations["betaCLI"]; ok {
			beta = true
		}
	})
	return beta
}

func isDeprecated(cmd *cobra.Command) bool {
	return cmd.Deprecated != ""
}

const usageTemplate = `{{- if isBeta .}}Beta Feature:
  {{.CommandPath}} is a beta feature.
  Beta features provide early access to product functionality. These
  features may change between releases without warning, or can be removed in a
  future release.

{{ end }}Usage:{{- if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
 {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
