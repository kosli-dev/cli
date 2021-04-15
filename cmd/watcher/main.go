package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/merkely-development/watcher/internal/app"
	"github.com/merkely-development/watcher/internal/version"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "merkely"

	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --namespace is bound to MERKELY_NAMESPACE.
	envPrefix = "MERKELY"
)

// harvestArgs is the harvest command arguments
type harvestArgs struct {
	kubeconfig         string
	namespaces         []string
	excludeNamespaces  []string
	merkelyEnvironment string
	apiToken           string
	owner              string
	host               string
	dryRun             bool
	version            bool
}

func main() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// NewRootCommand Build the cobra command that handles our command line tool.
func NewRootCommand() *cobra.Command {
	harvest := harvestArgs{}
	url := fmt.Sprintf("%s/api/v1/projects/%s", harvest.host, harvest.owner)

	// define the default kubeconfig path
	home, err := homedir.Dir()
	defaultKubeConfigPath := ""
	if err == nil {
		path := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(path)
		if err == nil {
			defaultKubeConfigPath = path
		}
	}

	// Define our command
	rootCmd := &cobra.Command{
		Use:   "merkely",
		Short: "harvest pod info from a cluster",
		Long:  `harvest pod image data from specific namespace or entire cluster`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return initializeConfig(cmd)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(harvest.excludeNamespaces) > 0 && len(harvest.namespaces) > 0 {
				return fmt.Errorf("--namespace and --exclude-namespace can't be used together")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if harvest.version {
				fmt.Printf("Version: %v", version.Get())
				os.Exit(0)
			}

			clientset, err := app.NewK8sClientSet(harvest.kubeconfig)
			if err != nil {
				log.Fatal(err)
			}
			podsData, err := app.GetPodsData(harvest.namespaces, harvest.excludeNamespaces, clientset)
			if err != nil {
				log.Fatal(err)
			}

			requestBody := &app.HarvestRequest{
				PodsData:    podsData,
				Owner:       harvest.owner,
				Environment: harvest.merkelyEnvironment,
			}
			js, _ := json.MarshalIndent(requestBody, "", "    ")

			if harvest.dryRun {
				fmt.Println("############### THIS IS A DRY-RUN  ###############")
				fmt.Println(string(js))
			} else {
				fmt.Println("****** Sending a Test to the API")
				fmt.Println(string(js))
				_, err = app.DoPost(js, url, harvest.apiToken)
				if err != nil {
					log.Fatal(err)
				}
			}
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	// CLI args are read from command line, then env variables, then config files, then the CLI arg default is used.
	rootCmd.Flags().StringVarP(&harvest.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath, "kubeconfig path for the target cluster")
	rootCmd.Flags().StringSliceVarP(&harvest.namespaces, "namespace", "n", []string{}, "the comma separated list of namespaces to harvest artifacts info from. Can't be used together with --exclude-namespace.")
	rootCmd.Flags().StringSliceVarP(&harvest.excludeNamespaces, "exclude-namespace", "x", []string{}, "the comma separated list of namespaces NOT to harvest artifacts info from. Can't be used together with --namespace.")
	rootCmd.Flags().StringVarP(&harvest.merkelyEnvironment, "environment", "e", "", "the name of the merkely environment.")
	rootCmd.Flags().StringVarP(&harvest.apiToken, "api-token", "a", "", "the merkely API token.")
	rootCmd.Flags().StringVarP(&harvest.owner, "owner", "o", "", "the merkely organization.")
	rootCmd.Flags().StringVarP(&harvest.host, "host", "H", "https://app.merkely.com", "the merkely endpoint.")
	rootCmd.Flags().BoolVarP(&harvest.dryRun, "dry-run", "d", false, "whether to send the request to the endpoint or just log it in stdout.")
	rootCmd.Flags().BoolVarP(&harvest.version, "version", "v", false, "print the version.")

	return rootCmd
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(defaultConfigFilename)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")

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
