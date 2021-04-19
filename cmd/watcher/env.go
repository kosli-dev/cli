package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/merkely-development/watcher/internal/kube"
	"github.com/merkely-development/watcher/internal/requests"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const envDesc = `
Report actual deployments in an environment back to Merkely.
This allows Merkely to establish Runtime compliance of the environment.

This command lists the artifacts deployed in the k8s environment and thier digests,
before reporting them to Merkely. 
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
	// TODO remove hard coded url
	url := fmt.Sprintf("%s/api/v1/projects/%s", global.host, global.owner)
	cmd := &cobra.Command{
		Use:   "env [env-name] [flags]",
		Short: "report images data from specific namespace or entire cluster to Merkely.",
		Long:  envDesc,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("environment name is required")
			}
			if len(o.excludeNamespaces) > 0 && len(o.namespaces) > 0 {
				return fmt.Errorf("--namespace and --exclude-namespace can't be used together")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			clientset, err := kube.NewK8sClientSet(o.kubeconfig)
			if err != nil {
				log.Fatal(err)
			}
			podsData, err := kube.GetPodsData(o.namespaces, o.excludeNamespaces, clientset)
			if err != nil {
				log.Fatal(err)
			}

			requestBody := &requests.EnvRequest{
				PodsData:    podsData,
				Owner:       global.owner,
				Environment: args[0],
			}
			js, _ := json.MarshalIndent(requestBody, "", "    ")

			if global.dryRun {
				fmt.Println("############### THIS IS A DRY-RUN  ###############")
				fmt.Println(string(js))
			} else {
				fmt.Println("****** Sending a Test to the API")
				fmt.Println(string(js))
				_, err = requests.DoPost(js, url, global.apiToken)
				if err != nil {
					log.Fatal(err)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath, "kubeconfig path for the target cluster")
	cmd.Flags().StringSliceVarP(&o.namespaces, "namespace", "n", []string{}, "the comma separated list of namespaces to harvest artifacts info from. Can't be used together with --exclude-namespace.")
	cmd.Flags().StringSliceVarP(&o.excludeNamespaces, "exclude-namespace", "x", []string{}, "the comma separated list of namespaces NOT to harvest artifacts info from. Can't be used together with --namespace.")

	return cmd
}
