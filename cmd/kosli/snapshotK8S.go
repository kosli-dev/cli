package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/kube"
	"github.com/kosli-dev/cli/internal/requests"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const snapshotK8SShortDesc = `Report a snapshot of running pods in a K8S cluster or namespace(s) to Kosli.  `

const snapshotK8SLongDesc = snapshotK8SShortDesc + `
Skip ^--namespaces^ and ^--namespaces-regex^ to report all pods in all namespaces in a cluster.
The reported data includes pod container images digests and creation timestamps. You can customize the scope of reporting
to include or exclude namespaces.`

const snapshotK8SExample = `
# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
kosli snapshot k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config (with global flags defined in environment or in a config file):
export KOSLI_API_TOKEN=yourAPIToken
export KOSLI_ORG=yourOrgName

kosli snapshot k8s yourEnvironmentName

# report what is running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
kosli snapshot k8s yourEnvironmentName \
    --exclude-namespaces kube-system,utilities \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
kosli snapshot k8s yourEnvironmentName \
	--namespaces your-namespace \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in a cluster using kubeconfig at a custom path:
kosli snapshot k8s yourEnvironmentName \
	--kubeconfig /path/to/kube/config \
	--api-token yourAPIToken \
	--org yourOrgName
`

const k8sConfigFileFlag = "[optional] The path to a YAML config file that maps multiple Kosli environments to namespace selectors. Cannot be used with a positional environment name argument or namespace flags."

type k8sSnapshotConfig struct {
	Environments []k8sEnvironmentConfig `yaml:"environments"`
}

type k8sEnvironmentConfig struct {
	Name                   string   `yaml:"name"`
	Namespaces             []string `yaml:"namespaces"`
	NamespacesRegex        []string `yaml:"namespacesRegex"`
	ExcludeNamespaces      []string `yaml:"excludeNamespaces"`
	ExcludeNamespacesRegex []string `yaml:"excludeNamespacesRegex"`
}

func (e *k8sEnvironmentConfig) toFilter() *filters.ResourceFilterOptions {
	return &filters.ResourceFilterOptions{
		IncludeNames:      e.Namespaces,
		IncludeNamesRegex: e.NamespacesRegex,
		ExcludeNames:      e.ExcludeNamespaces,
		ExcludeNamesRegex: e.ExcludeNamespacesRegex,
	}
}

func parseK8SSnapshotConfig(path string) (*k8sSnapshotConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	var config k8sSnapshotConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateK8SSnapshotConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateK8SSnapshotConfig(config *k8sSnapshotConfig) error {
	if len(config.Environments) == 0 {
		return fmt.Errorf("invalid config: 'environments' list must contain at least one entry")
	}

	seen := make(map[string]bool)
	for i, env := range config.Environments {
		if env.Name == "" {
			return fmt.Errorf("invalid config: environment entry %d is missing required field 'name'", i+1)
		}

		if seen[env.Name] {
			return fmt.Errorf("invalid config: duplicate environment name '%s'", env.Name)
		}
		seen[env.Name] = true

		hasInclude := len(env.Namespaces) > 0 || len(env.NamespacesRegex) > 0
		hasExclude := len(env.ExcludeNamespaces) > 0 || len(env.ExcludeNamespacesRegex) > 0
		if hasInclude && hasExclude {
			includeType := "namespaces"
			if len(env.Namespaces) == 0 {
				includeType = "namespacesRegex"
			}
			excludeType := "excludeNamespaces"
			if len(env.ExcludeNamespaces) == 0 {
				excludeType = "excludeNamespacesRegex"
			}
			return fmt.Errorf("invalid config for environment '%s': cannot combine '%s' with '%s'",
				env.Name, includeType, excludeType)
		}

		for _, pattern := range env.NamespacesRegex {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid config for environment '%s': invalid regex '%s': %v",
					env.Name, pattern, err)
			}
		}
		for _, pattern := range env.ExcludeNamespacesRegex {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid config for environment '%s': invalid regex '%s': %v",
					env.Name, pattern, err)
			}
		}
	}

	return nil
}

type snapshotK8SOptions struct {
	kubeconfig     string
	configFilePath string
	filter         *filters.ResourceFilterOptions
}

func newSnapshotK8SCmd(out io.Writer) *cobra.Command {
	o := new(snapshotK8SOptions)
	o.filter = new(filters.ResourceFilterOptions)
	cmd := &cobra.Command{
		Use:     "k8s ENVIRONMENT-NAME",
		Aliases: []string{"kubernetes"},
		Short:   snapshotK8SShortDesc,
		Long:    snapshotK8SLongDesc,
		Example: snapshotK8SExample,
		Args: func(cmd *cobra.Command, args []string) error {
			configFileFlag := cmd.Flags().Lookup("config-file")
			if configFileFlag != nil && configFileFlag.Changed && o.configFilePath == "" {
				return fmt.Errorf("cannot use '--config-file' with an empty value")
			}
			useConfigFile := o.configFilePath != ""
			if useConfigFile && len(args) > 0 {
				return fmt.Errorf("cannot use '--config-file' together with a positional environment name argument")
			}
			if !useConfigFile && len(args) == 0 {
				return fmt.Errorf("requires either a positional environment name argument or --config-file")
			}
			if !useConfigFile && len(args) > 1 {
				return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			configFileFlag := cmd.Flags().Lookup("config-file")
			if configFileFlag != nil && configFileFlag.Changed && o.configFilePath == "" {
				return ErrorBeforePrintingUsage(cmd, "cannot use '--config-file' with an empty value")
			}
			useConfigFile := o.configFilePath != ""
			if useConfigFile {
				namespaceFlagNames := []string{"namespaces", "exclude-namespaces", "namespaces-regex", "exclude-namespaces-regex"}
				for _, flagName := range namespaceFlagNames {
					if f := cmd.Flags().Lookup(flagName); f != nil && f.Changed {
						return fmt.Errorf("cannot use '--config-file' together with '--%s'", flagName)
					}
				}
				return nil
			}

			// Include vs exclude namespace flags mutual exclusion (all combinations)
			if err := MuXRequiredFlags(cmd, []string{"namespaces", "exclude-namespaces"}, false); err != nil {
				return err
			}
			if err := MuXRequiredFlags(cmd, []string{"namespaces", "exclude-namespaces-regex"}, false); err != nil {
				return err
			}
			if err := MuXRequiredFlags(cmd, []string{"namespaces-regex", "exclude-namespaces"}, false); err != nil {
				return err
			}
			if err := MuXRequiredFlags(cmd, []string{"namespaces-regex", "exclude-namespaces-regex"}, false); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.configFilePath != "" {
				return o.runMultiEnv()
			}
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath(), kubeconfigFlag)
	// Shadows the global --config-file persistent flag intentionally.
	// The global Kosli config can still be set via KOSLI_CONFIG_FILE env var.
	cmd.Flags().StringVar(&o.configFilePath, "config-file", "", k8sConfigFileFlag)
	cmd.Flags().StringSliceVarP(&o.filter.IncludeNames, "namespaces", "n", []string{}, namespacesFlag)
	cmd.Flags().StringSliceVar(&o.filter.IncludeNamesRegex, "namespaces-regex", []string{}, namespacesRegexFlag)
	cmd.Flags().StringSliceVarP(&o.filter.ExcludeNames, "exclude-namespaces", "x", []string{}, excludeNamespacesFlag)
	cmd.Flags().StringSliceVar(&o.filter.ExcludeNamesRegex, "exclude-namespaces-regex", []string{}, excludeNamespacesRegexFlag)
	addDryRunFlag(cmd)
	return cmd
}

func (o *snapshotK8SOptions) run(args []string) error {
	clientset, err := kube.NewK8sClientSet(o.kubeconfig)
	if err != nil {
		return err
	}
	return o.reportEnvironment(clientset, args[0], o.filter)
}

func (o *snapshotK8SOptions) runMultiEnv() error {
	config, err := parseK8SSnapshotConfig(o.configFilePath)
	if err != nil {
		return err
	}

	clientset, err := kube.NewK8sClientSet(o.kubeconfig)
	if err != nil {
		return err
	}

	var errs []string
	for _, env := range config.Environments {
		if err := o.reportEnvironment(clientset, env.Name, env.toFilter()); err != nil {
			errs = append(errs, fmt.Sprintf("environment '%s': %v", env.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

func (o *snapshotK8SOptions) reportEnvironment(clientset *kube.K8SConnection, envName string, filter *filters.ResourceFilterOptions) error {
	podsData, err := clientset.GetPodsData(filter, logger)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(global.Host, "api/v2/environments", global.Org, envName, "report/K8S")
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     url,
		Payload: &kube.K8sEnvRequest{Artifacts: podsData},
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] pods were reported to environment %s", len(podsData), envName)
	}
	return err
}

func defaultKubeConfigPath() string {
	if _, ok := os.LookupEnv("DOCS"); ok { // used for docs generation
		return "$HOME/.kube/config"
	}
	home, err := homedir.Dir()
	if err == nil {
		path := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(path)
		if err == nil {
			return path
		}
	}
	return ""
}
