package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/merkely-development/watcher/internal/kube"
	"github.com/merkely-development/watcher/internal/requests"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const envDesc = `
Report actual deployments in an environment back to Merkely.
This allows Merkely to determine Runtime compliance status of the environment.

This command lists the artifacts deployed in the k8s environment and their digests,
before reporting them to Merkely. 
`

const envExample = `
* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config:
merkely report env prod --api-token 1234 --owner exampleOrg

* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config 
(with global flags defined in environment or in  a config file):
merkely report env prod

* report what's running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
merkely report env prod -x kube-system,utilities

* report what's running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
merkely report env prod -n prod-namespace

* report what's running in a cluster using kubeconfig at a custom path:
merkely report env prod -k /path/to/kube/config
`

type envOptions struct {
	kubeconfig        string
	namespaces        []string
	excludeNamespaces []string
}

func newEnvCmd(out io.Writer) *cobra.Command {
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

	o := new(envOptions)
	cmd := &cobra.Command{
		Use:     "env [-n namespace | -x namespace]... [-k /path/to/kube/config] env-name",
		Short:   "report images data from specific namespace or entire cluster to Merkely.",
		Long:    envDesc,
		Aliases: []string{"environment"},
		Example: envExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only environment name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("environment name is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(o.excludeNamespaces) > 0 && len(o.namespaces) > 0 {
				return fmt.Errorf("--namespace and --exclude-namespace can't be used together. This can also happen if you set one of the two options in a config file or env var and the other on the command line")
			}
			envName := args[0]
			url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.host, global.owner, envName)
			clientset, err := kube.NewK8sClientSet(o.kubeconfig)
			if err != nil {
				return err
			}
			podsData, err := kube.GetPodsData(o.namespaces, o.excludeNamespaces, clientset)
			if err != nil {
				return err
			}

			requestBody := &requests.EnvRequest{
				Data: podsData,
				// Owner:       global.owner,
				// Environment: envName,
			}
			js, _ := json.MarshalIndent(requestBody, "", "    ")

			if global.dryRun {
				fmt.Println("############### THIS IS A DRY-RUN  ###############")
				fmt.Println(string(js))
			} else {
				fmt.Println("****** Sending a Test to the API ******")
				fmt.Println(string(js))
				resp, err := requests.DoPut(js, url, global.apiToken, global.maxAPIRetries)
				if err != nil {
					return err
				}
				if resp.StatusCode != 201 && resp.StatusCode != 200 {
					return fmt.Errorf("failed to send scrape data: %v", resp.Body)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath, "kubeconfig path for the target cluster")
	cmd.Flags().StringSliceVarP(&o.namespaces, "namespace", "n", []string{}, "the comma separated list of namespaces (or namespaces regex patterns) to harvest artifacts info from. Can't be used together with --exclude-namespace.")
	cmd.Flags().StringSliceVarP(&o.excludeNamespaces, "exclude-namespace", "x", []string{}, "the comma separated list of namespaces (or namespaces regex patterns) NOT to harvest artifacts info from. Can't be used together with --namespace.")

	return cmd
}
