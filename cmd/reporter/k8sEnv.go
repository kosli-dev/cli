package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/merkely-development/reporter/internal/kube"
	"github.com/merkely-development/reporter/internal/requests"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const k8sEnvDesc = `
List the artifacts deployed in the k8s environment and their digests
and report them to Merkely.
`

const k8sEnvExample = `
* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config:
merkely report env k8s prod --api-token 1234 --owner exampleOrg --id prod-cluster

* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config
(with global flags defined in environment or in  a config file):
merkely report env k8s prod

* report what's running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
merkely report env k8s prod -x kube-system,utilities

* report what's running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
merkely report env k8s prod -n prod-namespace

* report what's running in a cluster using kubeconfig at a custom path:
merkely report env k8s prod -k /path/to/kube/config
`

type k8sEnvOptions struct {
	kubeconfig        string
	namespaces        []string
	excludeNamespaces []string
	id                string
}

func newK8sEnvCmd(out io.Writer) *cobra.Command {
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

	o := new(k8sEnvOptions)
	cmd := &cobra.Command{
		Use:               "k8s [-n namespace | -x namespace]... [-k /path/to/kube/config] [-i infrastructure-identifier] env-name",
		Short:             "Report images data from specific namespace(s) or entire cluster to Merkely.",
		Long:              k8sEnvDesc,
		Aliases:           []string{"kubernetes"},
		Example:           k8sEnvExample,
		DisableAutoGenTag: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only environment name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("environment name is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(o.excludeNamespaces) > 0 && len(o.namespaces) > 0 {
				return fmt.Errorf("--namespace and --exclude-namespace can't be used together. This can also happen if you set one of the two options in a config file or env var and the other on the command line")
			}
			envName := args[0]
			if o.id == "" {
				o.id = envName
			}
			url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
			clientset, err := kube.NewK8sClientSet(o.kubeconfig)
			if err != nil {
				return err
			}
			podsData, err := kube.GetPodsData(o.namespaces, o.excludeNamespaces, clientset, log)
			if err != nil {
				return err
			}

			requestBody := &requests.K8sEnvRequest{
				Artifacts: podsData,
				Type:      "K8S",
				Id:        o.id,
			}

			_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
				global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
			return err
		},
	}

	cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath, "The kubeconfig path for the target cluster.")
	cmd.Flags().StringSliceVarP(&o.namespaces, "namespace", "n", []string{}, "The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace.")
	cmd.Flags().StringSliceVarP(&o.excludeNamespaces, "exclude-namespace", "x", []string{}, "The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace.")
	cmd.Flags().StringVarP(&o.id, "id", "i", "", "The unique identifier of the source infrastructure of the report (e.g. the K8S cluster/namespace name). If not set, it is defaulted to environment name.")

	return cmd
}
