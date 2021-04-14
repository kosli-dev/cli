package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/merkely-development/watcher/internal/kube"
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
	kubeconfig string
	namespace  string
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
		Run: func(cmd *cobra.Command, args []string) {
			clientset, err := kube.NewK8sClientSet(harvest.kubeconfig)
			if err != nil {
				log.Fatal(err)
			}
			podsData, err := kube.GetPodsData(harvest.namespace, clientset)
			if err != nil {
				log.Fatal(err)
			}

			js, _ := json.MarshalIndent(podsData, "", "    ")
			//TODO: send the json to merkely API
			fmt.Println(string(js))
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	// CLI args are read from command line, then env variables, then config files, then the CLI arg default is used.
	rootCmd.Flags().StringVarP(&harvest.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath, "kubeconfig path for the target cluster")
	rootCmd.Flags().StringVarP(&harvest.namespace, "namespace", "n", "", "the namespace to harvest artifacts info from.")

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
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
