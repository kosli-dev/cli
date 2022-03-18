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

var globalUsage = `The Merkely evidence reporting CLI.

Environment variables:
| Name                               | Description                                                                       |
|------------------------------------|-----------------------------------------------------------------------------------|
| $MERKELY_API_TOKEN                 | set the Merkely API token.                                                        |
| $MERKELY_OWNER                     | set the Merkely Pipeline Owner.                                                   |
| $MERKELY_HOST                      | set the Merkely host.                                                             |
| $MERKELY_DRY_RUN                   | indicate whether or not Merkely CLI is running in Dry Run mode.                   |
| $MERKELY_MAX_API_RETRIES           | set the maximum number of API calling retries when the API host is not reachable. |
| $MERKELY_CONFIG_FILE               | set the path to Merkely config file where you can set your options.               |         
`

const (
	maxAPIRetries = 3
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "merkely"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --namespace is bound to MERKELY_NAMESPACE.
	envPrefix = "MERKELY"

	// the following constants are used in the docs/help
	sha256Desc = "The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly."

	// flags
	sha256Flag              = "The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type."
	pipelineNameFlag        = "The Merkely pipeline name."
	oldestCommitFlag        = "The source commit sha for the oldest change in the deployment."
	newestCommitFlag        = "The source commit sha for the newest change in the deployment."
	repoRootFlag            = "The directory where the source git repository is volume-mounted."
	approvalDescriptionFlag = "[optional] The approval description."
	artifactDescriptionFlag = "[optional] The artifact description."
	evidenceDescriptionFlag = "[optional] The evidence description."
	approvalUserDataFlag    = "[optional] The path to a JSON file containing additional data you would like to attach to this approval."
	evidenceUserDataFlag    = "[optional] The path to a JSON file containing additional data you would like to attach to this evidence."
	gitCommitFlag           = "The git commit from which the artifact was created."
	buildUrlFlag            = "The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.merkely.com/ci-defaults)"
	commitUrlFlag           = "The url for the git commit that created the artifact."
	evidenceBuildUrlFlag    = "The url of CI pipeline that generated the evidence."
	compliantFlag           = "Whether the artifact is compliant or not."
	evidenceCompliantFlag   = "Whether the evidence is compliant or not."
	evidenceTypeFlag        = "The type of evidence being reported."
	bbUsernameFlag          = "Bitbucket user name."
	bbPasswordFlag          = "Bitbucket password."
	bbWorkspaceFlag         = "Bitbucket workspace."
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
		Use:              "merkely",
		Short:            "The Merkely evidence reporting CLI.",
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
	cmd.PersistentFlags().StringVarP(&global.ApiToken, "api-token", "a", "", "The merkely API token.")
	cmd.PersistentFlags().StringVarP(&global.Owner, "owner", "o", "", "The merkely user or organization.")
	cmd.PersistentFlags().StringVarP(&global.Host, "host", "H", "https://app.merkely.com", "The merkely endpoint.")
	cmd.PersistentFlags().BoolVarP(&global.DryRun, "dry-run", "D", false, "Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.")
	cmd.PersistentFlags().IntVarP(&global.MaxAPIRetries, "max-api-retries", "r", maxAPIRetries, "How many times should API calls be retried when the API host is not reachable.")
	cmd.PersistentFlags().StringVarP(&global.ConfigFile, "config-file", "c", defaultConfigFilename, "[optional] The merkely config file path.")
	cmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Print verbose logs to stdout.")

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
