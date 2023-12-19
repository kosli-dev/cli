package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var globalUsage = `The Kosli evidence reporting CLI.

Environment variables:
You can set any flag from an environment variable by capitalizing it in snake case and adding the KOSLI_ prefix.
For example, to set --api-token from an environment variable, you can export KOSLI_API_TOKEN=YOUR_API_TOKEN.

Setting the API token to DRY_RUN sets the --dry-run flag.
`

const (
	maxAPIRetries = 3
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "kosli"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --namespace is bound to KOSLI_NAMESPACE.
	envPrefix = "KOSLI"

	// the following constants are used in the docs/help
	fingerprintDesc = "The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag)."
	awsAuthDesc     = `

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
  2) microsoft.web/sites/containerlogs/action  
	`

	// flags
	apiTokenFlag                = "The Kosli API token."
	artifactName                = "[optional] Artifact display name, if different from file, image or directory name."
	artifactDisplayName         = "[optional] Artifact display name, if different from file, image or directory name."
	orgFlag                     = "The Kosli organization."
	hostFlag                    = "[defaulted] The Kosli endpoint."
	dryRunFlag                  = "[optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors."
	maxAPIRetryFlag             = "[defaulted] How many times should API calls be retried when the API host is not reachable."
	configFileFlag              = "[optional] The Kosli config file path."
	verboseFlag                 = "[optional] Print verbose logs to stdout."
	debugFlag                   = "[optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)"
	artifactTypeFlag            = "[conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'."
	flowNameFlag                = "The Kosli flow name."
	trailNameFlag               = "The Kosli trail name."
	templateArtifactName        = "The name of the artifact in the yml template file."
	auditTrailNameFlag          = "The Kosli audit trail name."
	workflowIDFlag              = "The ID of the workflow."
	stepNameFlag                = "The name of the step as defined in the audit trail's steps."
	flowNamesFlag               = "[defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org."
	newFlowFlag                 = "The name of the flow to be created or updated."
	outputFlag                  = "[defaulted] The format of the output. Valid formats are: [table, json]."
	pipefileFlag                = "[deprecated] The path to the JSON pipefile."
	environmentNameFlag         = "The environment name."
	approvalEnvironmentNameFlag = "[defaulted] The environment the artifact is approved for. (defaults to all environments)"
	pageNumberFlag              = "[defaulted] The page number of a response."
	pageLimitFlag               = "[defaulted] The number of elements per page."
	newEnvNameFlag              = "The name of environment to be created."
	newEnvTypeFlag              = "The type of environment. Valid types are: [K8S, ECS, server, S3, lambda, docker]."
	envAllowListFlag            = "The environment name for which the artifact is allowlisted."
	reasonFlag                  = "The reason why this artifact is allowlisted."
	oldestCommitFlag            = "[conditional] The source commit sha for the oldest change in the deployment. Can be any commit-ish. Only required if you don't specify '--environment'."
	newestCommitFlag            = "[defaulted] The source commit sha for the newest change in the deployment. Can be any commit-ish."
	repoRootFlag                = "[defaulted] The directory where the source git repository is available."
	approvalDescriptionFlag     = "[optional] The approval description."
	artifactDescriptionFlag     = "[optional] The artifact description."
	deploymentDescriptionFlag   = "[optional] The deployment description."
	evidenceDescriptionFlag     = "[optional] The evidence description."
	jiraBaseUrlFlag             = "The base url for the jira project, e.g. 'https://kosli.atlassian.net/browse/'"
	jiraUsernameFlag            = "Jira username (for Jira Cloud)"
	jiraAPITokenFlag            = "Jira API token (for Jira Cloud)"
	jiraPATFlag                 = "Jira personal access token (for self-hosted Jira)"
	envDescriptionFlag          = "[optional] The environment description."
	flowDescriptionFlag         = "[optional] The Kosli flow description."
	trailDescriptionFlag        = "[optional] The Kosli trail description."
	workflowDescriptionFlag     = "[optional] The Kosli Workflow description."
	visibilityFlag              = "[defaulted] The visibility of the Kosli flow. Valid visibilities are [public, private]."
	templateFlag                = "[defaulted] The comma-separated list of required compliance controls names."
	templateFileFlag            = "The path to a yaml template file."
	stepsFlag                   = "[defaulted] The comma-separated list of required audit trail steps names."
	approvalUserDataFlag        = "[optional] The path to a JSON file containing additional data you would like to attach to the approval."
	evidenceUserDataFlag        = "[optional] The path to a JSON file containing additional data you would like to attach to the evidence."
	attestationUserDataFlag     = "[optional] The path to a JSON file containing additional data you would like to attach to the attestation."
	deploymentUserDataFlag      = "[optional] The path to a JSON file containing additional data you would like to attach to the deployment."
	trailUserDataFlag           = "[optional] The path to a JSON file containing additional data you would like to attach to the flow trail."
	gitCommitFlag               = "[defaulted] The git commit from which the artifact was created. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD )."
	evidenceBuildUrlFlag        = "The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	buildUrlFlag                = "The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	commitUrlFlag               = "The url for the git commit that created the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	evidenceCompliantFlag       = "[defaulted] Whether the evidence is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags"
	evidenceTypeFlag            = "The type of evidence being reported."
	bbUsernameFlag              = "Bitbucket username."
	bbPasswordFlag              = "Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details."
	bbWorkspaceFlag             = "Bitbucket workspace ID."
	commitPREvidenceFlag        = "Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	commitEvidenceFlag          = "Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	repositoryFlag              = "Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	assertPREvidenceFlag        = "[optional] Exit with non-zero code if no pull requests found for the given commit."
	assertJiraEvidenceFlag      = "[optional] Exit with non-zero code if no jira issue reference found, or jira issue does not exist, for the given commit or branch."
	assertStatusFlag            = "[optional] Exit with non-zero code if Kosli server is not responding."
	azureTokenFlag              = "Azure Personal Access token."
	azureProjectFlag            = "Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults )."
	azureOrgUrlFlag             = "Azure organization url. E.g. \"https://dev.azure.com/myOrg\" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults )."
	azureBaseURLFlag            = "[optional] Azure Devops base URL."
	azureClientIdFlag           = "Azure client ID."
	azureClientSecretFlag       = "Azure client secret."
	azureTenantIdFlag           = "Azure tenant ID."
	azureSubscriptionIdFlag     = "Azure subscription ID."
	azureResourceGroupNameFlag  = "Azure resource group name."
	azureDigestsSourceFlag      = "[defaulted] Where to get the digests from. Valid values are 'acr' and 'logs'. Defaults to 'acr'"
	githubTokenFlag             = "Github token."
	githubOrgFlag               = "Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults )."
	githubBaseURLFlag           = "[optional] GitHub base URL (only needed for GitHub Enterprise installations)."
	gitlabTokenFlag             = "Gitlab token."
	gitlabOrgFlag               = "Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults )."
	gitlabBaseURLFlag           = "[optional] Gitlab base URL (only needed for on-prem Gitlab installations)."
	registryProviderFlag        = "[conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry."
	registryUsernameFlag        = "[conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry."
	registryPasswordFlag        = "[conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry."
	resultsDirFlag              = "[defaulted] The path to a directory with JUnit test results. The directory will be uploaded to Kosli's evidence vault."
	snykJsonResultsFileFlag     = "The path to Snyk scan results JSON file from 'snyk test' and 'snyk container test'. The Snyk results will be uploaded to Kosli's evidence vault."
	ecsClusterFlag              = "The name of the ECS cluster."
	ecsServiceFlag              = "[optional] The name of the ECS service."
	kubeconfigFlag              = "[defaulted] The kubeconfig path for the target cluster."
	namespaceFlag               = "[conditional] The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace."
	excludeNamespaceFlag        = "[conditional] The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace."
	functionNameFlag            = "[optional] The name of the AWS Lambda function."
	functionNamesFlag           = "[optional] The comma-separated list of AWS Lambda function names to be reported."
	functionVersionFlag         = "[optional] The version of the AWS Lambda function."
	awsKeyIdFlag                = "The AWS access key ID."
	awsSecretKeyFlag            = "The AWS secret access key."
	awsRegionFlag               = "The AWS region."
	bucketNameFlag              = "The name of the S3 bucket."
	pathsFlag                   = "The comma separated list of artifact directories."
	excludePathsFlag            = "[optional] The comma separated list of directories and files to exclude from fingerprinting. Only applicable for --artifact-type dir."
	shortFlag                   = "[optional] Print only the Kosli CLI version number."
	longFlag                    = "[optional] Print detailed output."
	reverseFlag                 = "[defaulted] Reverse the order of output list."
	evidenceNameFlag            = "The name of the evidence."
	evidenceFingerprintFlag     = "[optional] The SHA256 fingerprint of the evidence file or dir."
	evidenceURLFlag             = "[optional] The external URL where the evidence file or dir is stored."
	evidencePathsFlag           = "[optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault."
	fingerprintFlag             = "[conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'."
	evidenceCommitFlag          = "The git commit SHA1 for which the evidence belongs. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	intervalFlag                = "[optional] Expression to define specified snapshots range."
	showUnchangedArtifactsFlag  = "[defaulted] Show the unchanged artifacts present in both snapshots within the diff output."
	approverFlag                = "[optional] The user approving an approval."
	attestationFingerprintFlag  = "[optional] The SHA256 fingerprint of the artifact to attach the attestation to."
	attestationCommitFlag       = "The git commit associated to the attestation. (defaulted in some CIs: https://docs.kosli.com/ci-defaults )."
	attestationUrlFlag          = "The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults )."
	attestationNameFlag         = "The name of the attestation as declared in the flow or trail yaml template."
	attestationCompliantFlag    = "[defaulted] Whether the attestation is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags"
	attestationRepoRootFlag     = "[defaulted] The directory where the source git repository is available. Only used if --commit is used."
	uploadJunitResultsFlag      = "[defaulted] Whether to upload the provided Junit results directory as evidence to Kosli or not."
	attestationAssertFlag       = "[optional] Exit with non-zero code if the attestation is non-compliant"
)

var global *GlobalOpts

type GlobalOpts struct {
	ApiToken      string
	Org           string
	Host          string
	DryRun        bool
	MaxAPIRetries int
	ConfigFile    string
	Verbose       bool
	Debug         bool
}

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
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
			err := initialize(cmd, out)
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
			})

			return flagError
		},
	}
	cmd.PersistentFlags().StringVarP(&global.ApiToken, "api-token", "a", "", apiTokenFlag)
	cmd.PersistentFlags().StringVar(&global.Org, "org", "", orgFlag)
	cmd.PersistentFlags().StringVarP(&global.Host, "host", "H", "https://app.kosli.com", hostFlag)
	cmd.PersistentFlags().IntVarP(&global.MaxAPIRetries, "max-api-retries", "r", maxAPIRetries, maxAPIRetryFlag)
	cmd.PersistentFlags().StringVarP(&global.ConfigFile, "config-file", "c", defaultConfigFilename, configFileFlag)
	cmd.PersistentFlags().BoolVarP(&global.Debug, "verbose", "v", false, verboseFlag)
	cmd.PersistentFlags().BoolVar(&global.Debug, "debug", false, debugFlag)

	err := cmd.PersistentFlags().MarkDeprecated("verbose", "use --debug instead")
	if err != nil {
		return cmd, err
	}

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
		newSnapshotCmd(out),
		newRequestCmd(out),
		newLogCmd(out),
		newDisableCmd(out),
		newEnableCmd(out),
	)

	cobra.AddTemplateFunc("isBeta", isBeta)
	cmd.SetUsageTemplate(usageTemplate)

	return cmd, nil
}

func initialize(cmd *cobra.Command, out io.Writer) error {
	logger.DebugEnabled = global.Debug
	logger.SetInfoOut(out) // needed to allow tests to overwrite the logger output stream
	kosliClient.SetDebug(global.Debug)
	kosliClient.SetMaxAPIRetries(global.MaxAPIRetries)
	kosliClient.SetLogger(logger)
	v := viper.New()

	// If provided, extract the custom config file dir and name

	// handle passing the config file as an env variable.
	// we load the config file before we bind env vars to flags,
	// so we check for the config file env var separately here
	if global.ConfigFile == defaultConfigFilename {
		if path, exists := os.LookupEnv("KOSLI_CONFIG_FILE"); exists {
			global.ConfigFile = path
		}
	}
	dir, file := filepath.Split(global.ConfigFile)
	file = strings.TrimSuffix(file, filepath.Ext(file))

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(file)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	if dir == "" {
		dir = "."
	}
	v.AddConfigPath(dir)

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --namespace
	// binds to an environment variable KOSLI_NAMESPACE. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --kube-config which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
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
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
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

const usageTemplate = `{{- if isBeta .}}Beta Feature:
  {{.CommandPath}} is a beta feature.
  Beta features provide early access to product functionality. These
  features may change between releases without warning, or can be removed from a
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
