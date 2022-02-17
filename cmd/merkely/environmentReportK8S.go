package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/merkely-development/reporter/internal/kube"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const environmentReportK8SDesc = `
List the artifacts deployed in the k8s environment and their digests 
and report them to Merkely. 
`

const environmentReportK8SExample = `
* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config:
merkely environment report k8s prod --api-token 1234 --owner exampleOrg --id prod-cluster

* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config 
(with global flags defined in environment or in  a config file):
merkely environment report  k8s prod

* report what's running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
merkely environment report k8s prod -x kube-system,utilities

* report what's running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
merkely environment report k8s prod -n prod-namespace

* report what's running in a cluster using kubeconfig at a custom path:
merkely environment report k8s prod -k /path/to/kube/config
`

type environmentReportK8SOptions struct {
	kubeconfig        string
	namespaces        []string
	excludeNamespaces []string
	id                string
}

func newEnvironmentReportK8SCmd(out io.Writer) *cobra.Command {
	o := new(environmentReportK8SOptions)
	cmd := &cobra.Command{
		Use:     "k8s [-n namespace | -x namespace]... [-k /path/to/kube/config] [-i infrastructure-identifier] env-name",
		Short:   "Report images data from specific namespace(s) or entire cluster to Merkely.",
		Long:    environmentReportK8SDesc,
		Aliases: []string{"kubernetes"},
		Example: environmentReportK8SExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return ErrorAfterPrintingHelp(cmd, "only env-name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return ErrorAfterPrintingHelp(cmd, "env-name argument is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath(), "The kubeconfig path for the target cluster.")
	cmd.Flags().StringSliceVarP(&o.namespaces, "namespace", "n", []string{}, "The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace.")
	cmd.Flags().StringSliceVarP(&o.excludeNamespaces, "exclude-namespace", "x", []string{}, "The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace.")
	cmd.Flags().StringVarP(&o.id, "id", "i", "", "The unique identifier of the source infrastructure of the report (e.g. the K8S cluster/namespace name). If not set, it is defaulted to environment name.")

	return cmd
}

func (o *environmentReportK8SOptions) run(args []string) error {
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

	requestBody := &kube.K8sEnvRequest{
		Artifacts: podsData,
		Type:      "K8S",
		Id:        o.id,
	}

	_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}

func defaultKubeConfigPath() string {
	if _, ok := os.LookupEnv("DEV"); ok { // used for docs generation
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
