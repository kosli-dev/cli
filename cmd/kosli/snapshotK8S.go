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

const snapshotK8SShortDesc = `Report a snapshot of running pods in a K8S cluster or namespace(s) to Kosli.  `

const snapshotK8SLongDesc = snapshotK8SShortDesc + `
The reported data includes pod container images digests and creation timestamps. You can customize the scope of reporting
to include or exclude namespaces.`

const snapshotK8SExample = `
# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
kosli snapshot k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config 
(with global flags defined in environment or in a config file):
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
kosli environment report k8s yourEnvironmentName \
	--kubeconfig /path/to/kube/config \
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotK8SOptions struct {
	kubeconfig        string
	namespaces        []string
	excludeNamespaces []string
}

func newSnapshotK8SCmd(out io.Writer) *cobra.Command {
	o := new(snapshotK8SOptions)
	cmd := &cobra.Command{
		Use:     "k8s ENVIRONMENT-NAME",
		Aliases: []string{"kubernetes"},
		Short:   snapshotK8SShortDesc,
		Long:    snapshotK8SLongDesc,
		Example: snapshotK8SExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return MuXRequiredFlags(cmd, []string{"namespaces", "exclude-namespaces"}, false)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "k", defaultKubeConfigPath(), kubeconfigFlag)
	cmd.Flags().StringSliceVarP(&o.namespaces, "namespaces", "n", []string{}, namespaceFlag)
	cmd.Flags().StringSliceVarP(&o.excludeNamespaces, "exclude-namespaces", "x", []string{}, excludeNamespaceFlag)
	addDryRunFlag(cmd)
	return cmd
}

func (o *snapshotK8SOptions) run(args []string) error {
	envName := args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/K8S", global.Host, global.Org, envName)
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
