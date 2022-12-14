package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/kube"
	"github.com/kosli-dev/cli/internal/requests"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const environmentReportK8SShortDesc = `Report running pods data from K8S cluster or namespace(s) to Kosli.`

const environmentReportK8SLongDesc = environmentReportK8SShortDesc + `
The reported data includes pod container images digests and creation timestamps. You can customize the scope of reporting
to include or exclude namespaces.`

const environmentReportK8SExample = `
# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config 
(with global flags defined in environment or in a config file):
export KOSLI_API_TOKEN=yourAPIToken
export KOSLI_OWNER=yourOrgName

kosli environment report k8s yourEnvironmentName

# report what is running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
    --exclude-namespace kube-system,utilities \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
	--namespace your-namespace \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a cluster using kubeconfig at a custom path:
kosli environment report k8s yourEnvironmentName \
	--kubeconfig /path/to/kube/config \
	--api-token yourAPIToken \
	--owner yourOrgName
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
		Use:     "k8s ENVIRONMENT-NAME",
		Short:   environmentReportK8SShortDesc,
		Long:    environmentReportK8SLongDesc,
		Aliases: []string{"kubernetes"},
		Example: environmentReportK8SExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath(), kubeconfigFlag)
	cmd.Flags().StringSliceVarP(&o.namespaces, "namespace", "n", []string{}, namespaceFlag)
	cmd.Flags().StringSliceVarP(&o.excludeNamespaces, "exclude-namespace", "x", []string{}, excludeNamespaceFlag)
	addDryRunFlag(cmd)
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
	podsData, err := kube.GetPodsData(o.namespaces, o.excludeNamespaces, clientset, logger)
	if err != nil {
		return err
	}

	payload := &kube.K8sEnvRequest{
		Artifacts: podsData,
		Type:      "K8S",
		Id:        o.id,
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] pods were reported to environment %s", len(payload.Artifacts), envName)
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
