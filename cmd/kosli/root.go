package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var globalUsage = `The Kosli evidence reporting CLI.

Environment variables:
You can set any flag from an environment variable by capitalizing it in snake case and adding the KOSLI_ prefix.
For example, to set --api-token from an environment variable, you can export KOSLI_API_TOKEN
`

const (
	maxAPIRetries = 3
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "kosli"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --namespace is bound to KOSLI_NAMESPACE.
	envPrefix = "KOSLI"

	// the following constants are used in the docs/help
	sha256Desc = "The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag)."

	// flags
	apiTokenFlag            = "The Kosli API token."
	ownerFlag               = "The Kosli user or organization."
	hostFlag                = "[defaulted] The Kosli endpoint."
	dryRunFlag              = "[optional] Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors."
	maxAPIRetryFlag         = "[defaulted] How many times should API calls be retried when the API host is not reachable."
	configFileFlag          = "[optional] The Kosli config file path."
	verboseFlag             = "[optional] Print verbose logs to stdout."
	sha256Flag              = "[conditional] The SHA256 fingerprint for the artifact. Only required if you don't specify '--artifact-type'."
	artifactTypeFlag        = "[conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--sha256'."
	pipelineNameFlag        = "The Kosli pipeline name."
	newPipelineFlag         = "The name of the pipeline to be created or updated."
	pipefileFlag            = "[deprecated] The path to the JSON pipefile."
	environmentNameFlag     = "The environment name."
	environmentLongFlag     = "[optional] Print long environment info."
	environmentJsonFlag     = "[optional] Print environment info as json."
	newEnvNameFlag          = "The name of environment to be created."
	newEnvTypeFlag          = "The type of environment. Valid types are: [K8S, ECS, server, S3, lambda]."
	envAllowListFlag        = "The environment name for which the artifact is allowlisted."
	reasonFlag              = "[optional] The reason why this artifact is allowlisted."
	oldestCommitFlag        = "The source commit sha for the oldest change in the deployment."
	newestCommitFlag        = "[defaulted] The source commit sha for the newest change in the deployment."
	repoRootFlag            = "The directory where the source git repository is volume-mounted."
	approvalDescriptionFlag = "[optional] The approval description."
	artifactDescriptionFlag = "[optional] The artifact description."
	evidenceDescriptionFlag = "[optional] The evidence description."
	envDescriptionFlag      = "[optional] The environment description."
	pipelineDescriptionFlag = "[optional] The Kosli pipeline description."
	visibilityFlag          = "[defaulted] The visibility of the Kosli pipeline. Valid visibilities are [public, private]."
	templateFlag            = "[defaulted] The comma-separated list of required compliance controls names."
	approvalUserDataFlag    = "[optional] The path to a JSON file containing additional data you would like to attach to this approval."
	evidenceUserDataFlag    = "[optional] The path to a JSON file containing additional data you would like to attach to this evidence."
	deploymentUserDataFlag  = "[optional] The path to a JSON file containing additional data you would like to attach to this deployment."
	gitCommitFlag           = "The git commit from which the artifact was created. (defaulted in some CIs: https://docs.kosli.com/ci-defaults)."
	evidenceBuildUrlFlag    = "The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults)."
	buildUrlFlag            = "The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults)."
	commitUrlFlag           = "The url for the git commit that created the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults)."
	compliantFlag           = "[defaulted] Whether the artifact is compliant or not."
	evidenceCompliantFlag   = "[defaulted] Whether the evidence is compliant or not."
	evidenceTypeFlag        = "The type of evidence being reported."
	bbUsernameFlag          = "Bitbucket user name."
	bbPasswordFlag          = "Bitbucket password."
	bbWorkspaceFlag         = "Bitbucket workspace."
	commitPREvidenceFlag    = "Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults)."
	repositoryFlag          = "Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults)."
	assertPREvidenceFlag    = "[optional] Exit with non-zero code if no pull requests found for the given commit."
	assertStatusFlag        = "[optional] Exit with non-zero code if Kosli server is not responding."
	githubTokenFlag         = "Github token."
	githubOrgFlag           = "Github organization."
	registryProviderFlag    = "[conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry."
	registryUsernameFlag    = "[conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry."
	registryPasswordFlag    = "[conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry."
	resultsDirFlag          = "[defaulted] The path to a folder with JUnit test results."
	ecsClusterFlag          = "The name of the ECS cluster."
	ecsServiceFlag          = "The name of the ECS service."
	kubeconfigFlag          = "[defaulted] The kubeconfig path for the target cluster."
	namespaceFlag           = "[conditional] The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace."
	excludeNamespaceFlag    = "[conditional] The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace."
	functionNameFlag        = "The name of the AWS Lambda function."
	functionVersionFlag     = "[optional] The version of the AWS Lambda function."
	awsKeyIdFlag            = "The AWS access key ID."
	awsSecretKeyFlag        = "The AWS secret key."
	awsRegionFlag           = "The AWS region."
	bucketNameFlag          = "The name of the S3 bucket."
	pathsFlag               = "The comma separated list of artifact directories."
	shortFlag               = "[optional] Print only the Kosli cli version number."
)

var global *GlobalOpts

type GlobalOpts struct {
	ApiToken      string
	Owner         string
	Host          string
	DryRun        bool
	MaxAPIRetries int
	ConfigFile    string
	Verbose       bool
}

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	global = new(GlobalOpts)
	cmd := &cobra.Command{
		Use:              "kosli",
		Short:            "The Kosli CLI.",
		Long:             globalUsage,
		SilenceUsage:     true,
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			err := initializeConfig(cmd)
			if err != nil {
				return err
			}

			if global.ApiToken == "DRY_RUN" {
				global.DryRun = true
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&global.ApiToken, "api-token", "a", "", apiTokenFlag)
	cmd.PersistentFlags().StringVarP(&global.Owner, "owner", "o", "", ownerFlag)
	cmd.PersistentFlags().StringVarP(&global.Host, "host", "H", "https://app.kosli.com", hostFlag)
	cmd.PersistentFlags().BoolVarP(&global.DryRun, "dry-run", "D", false, dryRunFlag)
	cmd.PersistentFlags().IntVarP(&global.MaxAPIRetries, "max-api-retries", "r", maxAPIRetries, maxAPIRetryFlag)
	cmd.PersistentFlags().StringVarP(&global.ConfigFile, "config-file", "c", defaultConfigFilename, configFileFlag)
	cmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, verboseFlag)

	// Add subcommands
	cmd.AddCommand(

		newVersionCmd(out),
		newFingerprintCmd(out),
		newPipelineCmd(out),
		newEnvironmentCmd(out),
		newAssertCmd(out),
		newStatusCmd(out),
		// Hidden documentation generator command: 'kosli docs'
		newDocsCmd(out),
	)

	return cmd, nil
}

func initializeConfig(cmd *cobra.Command) error {
	if global.Verbose {
		log.Level = logrus.DebugLevel
	}

	v := viper.New()

	// If provided, extract the custom config file dir and name
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
				log.Fatalf("failed to bind viper to env variable: %v", err)
			}
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				log.Fatalf("failed to set flag: %v", err)
			}
		}
	})
}
