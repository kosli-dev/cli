package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

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

GCP authentication uses Application Default Credentials. On a developer machine, run
^gcloud auth application-default login^; in GCE/GKE/Cloud Run the metadata server / Workload
Identity is used automatically. The Cloud Asset API (^cloudasset.googleapis.com^) must be enabled
in the quota project of the caller's credentials.

The caller needs ^cloudasset.assets.listContainerPod^ and ^serviceusage.services.use^ on the
project, folder or organization. Grant them through a custom role for least privilege:
^roles/cloudasset.viewer^ also works, but it can list every asset type, including ^k8s.io/Secret^.

Asset Inventory is eventually consistent, so a snapshot can lag behind recent pod changes.

Skip ^--clusters^, ^--clusters-regex^ and ^--locations^ to report the pods of every cluster in
scope, and skip the namespace flags to report every namespace. Filters are case-sensitive.
With ^--folder^ or ^--organization^, ^--clusters^ and ^--clusters-regex^ match cluster names in
every project under the scope.

All selected clusters report to one environment, and the report does not carry the cluster
name: pods with the same namespace and name in two clusters (e.g. StatefulSet pods) are
indistinguishable. Snapshot such clusters to separate environments.`

const snapshotGKEExample = `
# report the pods of every GKE cluster in a project:
kosli snapshot gke yourEnvironmentName \
	--project yourGCPProject \
	--api-token yourAPIToken \
	--org yourOrgName

# report the pods of every GKE cluster in all projects under a folder:
kosli snapshot gke yourEnvironmentName \
	--folder yourGCPFolderID \
	--api-token yourAPIToken \
	--org yourOrgName

# report the pods of one cluster, excluding system namespaces:
kosli snapshot gke yourEnvironmentName \
	--project yourGCPProject \
	--clusters yourClusterName \
	--locations europe-west1 \
	--exclude-namespaces-regex "^kube-,^gke-" \
	--api-token yourAPIToken \
	--org yourOrgName
`

// gkePodLister lets tests replace the Asset Inventory client through newGKEClient.
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
			return o.validateRegexFlags()
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

// validateRegexFlags rejects invalid patterns up front, since Select reports only those a pod reaches.
func (o *snapshotGKEOptions) validateRegexFlags() error {
	for _, f := range []struct {
		name     string
		patterns []string
	}{
		{"clusters-regex", o.filter.Clusters.IncludeNamesRegex},
		{"namespaces-regex", o.filter.Namespaces.IncludeNamesRegex},
		{"exclude-namespaces-regex", o.filter.Namespaces.ExcludeNamesRegex},
	} {
		for _, pattern := range f.patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid --%s pattern '%s': %v", f.name, pattern, err)
			}
		}
	}
	return nil
}

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
	parent := o.parent()
	pods, err := client.ListPods(ctx, parent)
	if err != nil {
		return gke.Classify(err, parent)
	}
	selected, err := o.filter.Select(pods)
	if err != nil {
		return err
	}
	switch {
	case len(pods) == 0:
		logger.Warn("Asset Inventory returned no GKE pods in %s; reporting an empty snapshot to environment %s", parent, envName)
	case len(selected) == 0:
		logger.Warn("none of the %d pods in %s matched the cluster, location and namespace filters; reporting an empty snapshot to environment %s", len(pods), parent, envName)
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
