package main

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/gke"
	"github.com/kosli-dev/cli/internal/kube"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotGKEShortDesc = `Report a snapshot of running pods in GKE clusters to Kosli, read from Cloud Asset Inventory.  `

const snapshotGKELongDesc = snapshotGKEShortDesc + `
Reads the pods of every GKE cluster in a Google Cloud project, folder or organization from
Cloud Asset Inventory, so it needs no kubeconfig, no Kubernetes RBAC and no network access to any
cluster control plane. The reported data matches ^kosli snapshot k8s^: container image digests,
creation timestamps and owners of the Running and Failed pods.

Skip ^--clusters^, ^--clusters-regex^ and ^--locations^ to report the pods of every cluster in
scope, and skip the namespace flags to report every namespace.`

const snapshotGKEExample = `
# report the pods of every GKE cluster in a project:
kosli snapshot gke yourEnvironmentName \
	--project yourGCPProject \
	--api-token yourAPIToken \
	--org yourOrgName
`

// gkePodLister is the seam between the command and the GCP client. Tests
// override newGKEClient with a stub that returns canned pods.
type gkePodLister interface {
	ListPods(ctx context.Context, parent string) ([]gke.Pod, error)
}

var newGKEClient = func(ctx context.Context) (gkePodLister, error) {
	return gke.New(ctx)
}

type snapshotGKEOptions struct {
	project      string
	folder       string
	organization string
	filter       gke.Filter
}

func newSnapshotGKECmd(out io.Writer) *cobra.Command {
	o := new(snapshotGKEOptions)
	cmd := &cobra.Command{
		Use:         "gke ENVIRONMENT-NAME",
		Short:       snapshotGKEShortDesc,
		Long:        snapshotGKELongDesc,
		Example:     snapshotGKEExample,
		Annotations: map[string]string{betaCLIAnnotation: ""},
		Args:        cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if err := MuXRequiredFlags(cmd, []string{"project", "folder", "organization"}, true); err != nil {
				return err
			}
			for _, pair := range [][]string{
				{"namespaces", "exclude-namespaces"},
				{"namespaces", "exclude-namespaces-regex"},
				{"namespaces-regex", "exclude-namespaces"},
				{"namespaces-regex", "exclude-namespaces-regex"},
			} {
				if err := MuXRequiredFlags(cmd, pair, false); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.project, "project", "", gkeProjectFlag)
	cmd.Flags().StringVar(&o.folder, "folder", "", gkeFolderFlag)
	cmd.Flags().StringVar(&o.organization, "organization", "", gkeOrganizationFlag)
	cmd.Flags().StringSliceVar(&o.filter.Clusters.IncludeNames, "clusters", []string{}, gkeClustersFlag)
	cmd.Flags().StringSliceVar(&o.filter.Clusters.IncludeNamesRegex, "clusters-regex", []string{}, gkeClustersRegexFlag)
	cmd.Flags().StringSliceVar(&o.filter.Locations, "locations", []string{}, gkeLocationsFlag)
	cmd.Flags().StringSliceVarP(&o.filter.Namespaces.IncludeNames, "namespaces", "n", []string{}, namespacesFlag)
	cmd.Flags().StringSliceVar(&o.filter.Namespaces.IncludeNamesRegex, "namespaces-regex", []string{}, gkeNamespacesRegexFlag)
	cmd.Flags().StringSliceVarP(&o.filter.Namespaces.ExcludeNames, "exclude-namespaces", "x", []string{}, gkeExcludeNamespacesFlag)
	cmd.Flags().StringSliceVar(&o.filter.Namespaces.ExcludeNamesRegex, "exclude-namespaces-regex", []string{}, gkeExcludeNamespacesRegexFlag)
	addDryRunFlag(cmd)
	return cmd
}

// parent is the Asset Inventory scope the pods are listed under.
func (o *snapshotGKEOptions) parent() string {
	switch {
	case o.folder != "":
		return "folders/" + o.folder
	case o.organization != "":
		return "organizations/" + o.organization
	default:
		return "projects/" + o.project
	}
}

func (o *snapshotGKEOptions) run(args []string) error {
	envName := args[0]

	if err := ensureEnvironment(envName, "K8S"); err != nil {
		return err
	}

	reportURL, err := url.JoinPath(global.Host, "api/v2/environments", global.Org, envName, "report/K8S")
	if err != nil {
		return err
	}

	ctx := context.Background()
	client, err := newGKEClient(ctx)
	if err != nil {
		return err
	}
	if closer, ok := client.(io.Closer); ok {
		defer func() { _ = closer.Close() }()
	}
	pods, err := client.ListPods(ctx, o.parent())
	if err != nil {
		return err
	}
	selected, err := o.filter.Select(pods)
	if err != nil {
		return err
	}
	podsData, err := kube.ProcessPods(selected, logger)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     reportURL,
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
