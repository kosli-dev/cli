package main

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

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
)

var global *globalOpts

type globalOpts struct {
	apiToken      string
	owner         string
	host          string
	dryRun        bool
	maxAPIRetries int
	configFile    string
}

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	global = new(globalOpts)
	cmd := &cobra.Command{
		Use:              "merkely",
		Short:            "The Merkely evidence reporting CLI.",
		Long:             globalUsage,
		SilenceUsage:     true,
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return initializeConfig(cmd)
		},
	}
	cmd.PersistentFlags().StringVarP(&global.apiToken, "api-token", "a", "", "the merkely API token.")
	cmd.PersistentFlags().StringVarP(&global.owner, "owner", "o", "", "the merkely organization.")
	cmd.PersistentFlags().StringVarP(&global.host, "host", "H", "https://app.merkely.com", "the merkely endpoint.")
	cmd.PersistentFlags().BoolVarP(&global.dryRun, "dry-run", "d", false, "whether to send the request to the endpoint or just log it in stdout.")
	cmd.PersistentFlags().IntVarP(&global.maxAPIRetries, "max-api-retries", "r", maxAPIRetries, "how many times should API calls be retried when the API host is not reachable.")
	cmd.PersistentFlags().StringVarP(&global.configFile, "config-file", "c", defaultConfigFilename, "[optional] the merkely config file path.")

	// Add subcommands
	cmd.AddCommand(

		newVersionCmd(out),
		newReportCmd(out),
		newCreateCmd(out),
		newFingerprintCmd(out),

		// Hidden documentation generator command: 'merkely docs'
		newDocsCmd(out),
	)

	return cmd, nil
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// If provided, extract the custom config file dir and name
	dir, file := filepath.Split(global.configFile)
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
