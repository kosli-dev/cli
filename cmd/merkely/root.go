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

| Name                               | Description                                                                       |
|------------------------------------|-----------------------------------------------------------------------------------|
| $MERKELY_API_TOKEN                 | set the Kosli API token.                                                        |
| $MERKELY_OWNER                     | set the Kosli Pipeline Owner.                                                   |
| $MERKELY_HOST                      | set the Kosli host.                                                             |
| $MERKELY_DRY_RUN                   | indicate whether or not Kosli CLI is running in Dry Run mode.                   |
| $MERKELY_MAX_API_RETRIES           | set the maximum number of API calling retries when the API host is not reachable. |
| $MERKELY_CONFIG_FILE               | set the path to Kosli config file where you can set your options.               |         
`

const (
	maxAPIRetries = 3
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "merkely"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --namespace is bound to MERKELY_NAMESPACE.
	envPrefix = "MERKELY"

	// the following constants are used in the docs/help
	sha256Desc = "The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag)."

	// flags
	apiTokenFlag            = "The Kosli API token."
	ownerFlag               = "The Kosli user or organization."
	hostFlag                = "The Kosli endpoint."
	dryRunFlag              = "Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors."
	maxAPIRetryFlag         = "How many times should API calls be retried when the API host is not reachable."
	configFileFlag          = "[optional] The Kosli config file path."
	verboseFlag             = "Print verbose logs to stdout."
	sha256Flag              = "The SHA256 fingerprint for the artifact. Only required if you don't specify 'artifact-type'."
	artifactTypeFlag        = "The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'"
	pipelineNameFlag        = "The Kosli pipeline name."
	newPipelineFlag         = "The name of the pipeline to be created or updated."
	pipefileFlag            = "[deprecated] The path to the JSON pipefile."
	environmentNameFlag     = "The environment name."
	environmentLongFlag     = "Print long environment info."
	environmentJsonFlag     = "Print environment info as json."
	newEnvNameFlag          = "The name of environment to be created."
	newEnvTypeFlag          = "The type of environment. Valid options are: [K8S, ECS, server, S3]"
	envAllowListFlag        = "The environment name for which the artifact is allowlisted."
	reasonFlag              = "The reason why this artifact is allowlisted."
	oldestCommitFlag        = "The source commit sha for the oldest change in the deployment."
	newestCommitFlag        = "The source commit sha for the newest change in the deployment."
	repoRootFlag            = "The directory where the source git repository is volume-mounted."
	approvalDescriptionFlag = "[optional] The approval description."
	artifactDescriptionFlag = "[optional] The artifact description."
	evidenceDescriptionFlag = "[optional] The evidence description."
	envDescriptionFlag      = "[optional] The environment description."
	pipelineDescriptionFlag = "[optional] The Kosli pipeline description."
	visibilityFlag          = "The visibility of the Kosli pipeline. Options are [public, private]."
	templateFlag            = "The comma-separated list of required compliance controls names."
	approvalUserDataFlag    = "[optional] The path to a JSON file containing additional data you would like to attach to this approval."
	evidenceUserDataFlag    = "[optional] The path to a JSON file containing additional data you would like to attach to this evidence."
	deploymentUserDataFlag  = "[optional] The path to a JSON file containing additional data you would like to attach to this deployment."
	gitCommitFlag           = "The git commit from which the artifact was created."
	evidenceBuildUrlFlag    = "The url of CI pipeline that generated the evidence."
	buildUrlFlag            = "The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.merkely.com/ci-defaults)"
	commitUrlFlag           = "The url for the git commit that created the artifact."
	compliantFlag           = "Whether the artifact is compliant or not."
	evidenceCompliantFlag   = "Whether the evidence is compliant or not."
	evidenceTypeFlag        = "The type of evidence being reported."
	bbUsernameFlag          = "Bitbucket user name."
	bbPasswordFlag          = "Bitbucket password."
	bbWorkspaceFlag         = "Bitbucket workspace."
	commitPREvidenceFlag    = "Git commit for which to find pull request evidence."
	repositoryFlag          = "Git repository."
	assertPREvidenceFlag    = "Exit with non-zero code if no pull requests found for the given commit."
	assertStatusFlag        = "Exit with non-zero code if Kosli server is not responding."
	githubTokenFlag         = "Github token."
	githubOrgFlag           = "Github organization."
	registryProviderFlag    = "The docker registry provider or url."
	registryUsernameFlag    = "The docker registry username."
	registryPasswordFlag    = "The docker registry password or access token."
	resultsDirFlag          = "The path to a folder with JUnit test results."
	ecsClusterFlag          = "The name of the ECS cluster."
	ecsServiceFlag          = "The name of the ECS service."
	kubeconfigFlag          = "The kubeconfig path for the target cluster."
	namespaceFlag           = "The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace."
	excludeNamespaceFlag    = "The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace."
	functionNameFlag        = "The name of the AWS Lambda function."
	functionVersionFlag     = "[optional] The version of the AWS Lambda function."
	awsKeyIdFlag            = "The AWS access key ID"
	awsSecretKeyFlag        = "The AWS secret key"
	awsRegionFlag           = "The AWS region"
	bucketNameFlag          = "The name of the S3 bucket."
	pathsFlag               = "The comma separated list of artifact directories."
	shortFlag               = "only print the version number"
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
	cmd.PersistentFlags().StringVarP(&global.Host, "host", "H", "https://app.merkely.com", hostFlag)
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
		// Hidden documentation generator command: 'merkely docs'
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
	// binds to an environment variable MERKELY_NAMESPACE. This helps
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
		// keys with underscores, e.g. --kube-config to MERKELY_KUBE_CONFIG
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
