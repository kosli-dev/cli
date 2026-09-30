package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/gke"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type stubGKEPodLister struct {
	pods   []gke.Pod
	err    error
	parent *string
}

func (s stubGKEPodLister) ListPods(_ context.Context, parent string) ([]gke.Pod, error) {
	if s.parent != nil {
		*s.parent = parent
	}
	return s.pods, s.err
}

var origNewGKEClient = newGKEClient

func stubGKEPod(location, cluster, namespace, name string, phase corev1.PodPhase, digest string) gke.Pod {
	image := "europe-west1-docker.pkg.dev/p/apps/" + name
	return gke.Pod{
		Location: location,
		Cluster:  cluster,
		Pod: corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				CreationTimestamp: metav1.NewTime(time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)),
			},
			Status: corev1.PodStatus{
				Phase: phase,
				ContainerStatuses: []corev1.ContainerStatus{{
					Name:    name,
					Image:   image + ":1.0.0",
					ImageID: image + "@sha256:" + digest,
				}},
			},
		},
	}
}

// Digests are 64-char hex because the server's K8S report model rejects anything else.
func stubGKEPods() []gke.Pod {
	return []gke.Pod{
		stubGKEPod("europe-west1", "prod-eu", "payments", "api", corev1.PodRunning, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		stubGKEPod("europe-west1", "prod-eu", "kube-system", "dns", corev1.PodRunning, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		stubGKEPod("us-central1-a", "staging-us", "payments", "worker", corev1.PodFailed, "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		stubGKEPod("us-central1-a", "staging-us", "payments", "migrate", corev1.PodSucceeded, "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
	}
}

type SnapshotGKETestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *SnapshotGKETestSuite) SetupTest() {
	suite.envName = "snapshot-gke-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	newGKEClient = func(_ context.Context) (gkePodLister, error) {
		return stubGKEPodLister{pods: stubGKEPods()}, nil
	}

	CreateEnv(global.Org, suite.envName, "K8S", suite.T())
}

func (suite *SnapshotGKETestSuite) TearDownTest() {
	newGKEClient = origNewGKEClient
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "snapshot gke fails if no args are provided",
			cmd:       fmt.Sprintf(`snapshot gke --project p %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if 2 args are provided",
			cmd:       fmt.Sprintf(`snapshot gke %s xxx --project p %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if no scope flag is set",
			cmd:       fmt.Sprintf(`snapshot gke %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: at least one of --project, --folder, --organization is required\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if --project and --folder are set",
			cmd:       fmt.Sprintf(`snapshot gke %s --project p --folder 123 %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --project, --folder, --organization is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if --folder and --organization are set",
			cmd:       fmt.Sprintf(`snapshot gke %s --folder 123 --organization 456 %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --project, --folder, --organization is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if both --namespaces and --exclude-namespaces are set",
			cmd:       fmt.Sprintf(`snapshot gke %s --project p --namespaces default --exclude-namespaces default %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --namespaces, --exclude-namespaces is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if both --namespaces and --exclude-namespaces-regex are set",
			cmd:       fmt.Sprintf(`snapshot gke %s --project p --namespaces default --exclude-namespaces-regex "^kube-" %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --namespaces, --exclude-namespaces-regex is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if both --namespaces-regex and --exclude-namespaces are set",
			cmd:       fmt.Sprintf(`snapshot gke %s --project p --namespaces-regex "^default" --exclude-namespaces kube-system %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --namespaces-regex, --exclude-namespaces is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot gke fails if both --namespaces-regex and --exclude-namespaces-regex are set",
			cmd:       fmt.Sprintf(`snapshot gke %s --project p --namespaces-regex "^default" --exclude-namespaces-regex "^kube-" %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --namespaces-regex, --exclude-namespaces-regex is allowed\n",
		},
		{
			wantError:   true,
			name:        "snapshot gke fails if --clusters-regex is an invalid regex",
			cmd:         fmt.Sprintf(`snapshot gke %s --project p --clusters-regex "[invalid" --dry-run %s`, suite.envName, suite.defaultKosliArguments),
			goldenRegex: `Error: invalid --clusters-regex pattern '\[invalid': error parsing regexp: missing closing \]`,
		},
		{
			wantError:   true,
			name:        "snapshot gke fails if --namespaces-regex is an invalid regex that no pod reaches",
			cmd:         fmt.Sprintf(`snapshot gke %s --project p --clusters nope --namespaces-regex "[invalid" --dry-run %s`, suite.envName, suite.defaultKosliArguments),
			goldenRegex: `Error: invalid --namespaces-regex pattern '\[invalid'`,
		},
		{
			name:        "snapshot gke dry-runs a K8S report built from the listed pods",
			cmd:         fmt.Sprintf(`snapshot gke %s --project p --dry-run %s`, suite.envName, suite.defaultKosliArguments),
			goldenRegex: `(?s)THIS IS A DRY-RUN.*report/K8S.*"podName": "api",\s*"namespace": "payments",\s*"digests": \{\s*"europe-west1-docker.pkg.dev/p/apps/api:1.0.0": "a{64}"`,
		},
		{
			name:        "snapshot gke with --auto-environment infers the K8S type",
			cmd:         fmt.Sprintf(`snapshot gke new-gke-env-%d --project p --auto-environment --dry-run %s`, os.Getpid(), suite.defaultKosliArguments),
			goldenRegex: fmt.Sprintf("dry-run: environment new-gke-env-%d would be created with type K8S if it does not exist", os.Getpid()),
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *SnapshotGKETestSuite) runDryRun(args string) string {
	cmd := fmt.Sprintf(`snapshot gke %s --dry-run %s %s`, suite.envName, args, suite.defaultKosliArguments)
	_, combined, _, _, err := executeCommandC(cmd)
	require.NoError(suite.T(), err, "command failed: %s", combined)
	return combined
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_ReportsRunningAndFailedPods() {
	out := suite.runDryRun("--project p")
	require.Contains(suite.T(), out, `"podName": "api"`)
	require.Contains(suite.T(), out, `"podName": "dns"`)
	require.Contains(suite.T(), out, `"podName": "worker"`)
	require.NotContains(suite.T(), out, `"podName": "migrate"`)
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_ScopeFlagSetsTheParent() {
	for _, tc := range []struct {
		args string
		want string
	}{
		{args: "--project my-project", want: "projects/my-project"},
		{args: "--folder 123", want: "folders/123"},
		{args: "--organization 456", want: "organizations/456"},
	} {
		suite.Run(tc.args, func() {
			var parent string
			newGKEClient = func(_ context.Context) (gkePodLister, error) {
				return stubGKEPodLister{parent: &parent}, nil
			}
			suite.runDryRun(tc.args)
			require.Equal(suite.T(), tc.want, parent)
		})
	}
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_Filters() {
	for _, tc := range []struct {
		args string
		want []string
	}{
		{args: "--clusters prod-eu", want: []string{"api", "dns"}},
		{args: `--clusters-regex "^staging-"`, want: []string{"worker"}},
		{args: "--locations us-central1-a", want: []string{"worker"}},
		{args: "--locations us-central1", want: []string{"worker"}},
		{args: "--locations europe-west1-b", want: []string{}},
		{args: "--namespaces kube-system", want: []string{"dns"}},
		{args: `--namespaces-regex "^pay"`, want: []string{"api", "worker"}},
		{args: "--exclude-namespaces kube-system", want: []string{"api", "worker"}},
		{args: `--exclude-namespaces-regex "^kube-"`, want: []string{"api", "worker"}},
		{args: "--clusters prod-eu --exclude-namespaces kube-system", want: []string{"api"}},
	} {
		suite.Run(tc.args, func() {
			out := suite.runDryRun("--project p " + tc.args)
			for _, name := range []string{"api", "dns", "worker"} {
				podName := fmt.Sprintf(`"podName": "%s"`, name)
				if slices.Contains(tc.want, name) {
					require.Contains(suite.T(), out, podName)
				} else {
					require.NotContains(suite.T(), out, podName)
				}
			}
		})
	}
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_RejectsInvalidRegexWithoutListingPods() {
	listed := false
	newGKEClient = func(_ context.Context) (gkePodLister, error) {
		listed = true
		return stubGKEPodLister{}, nil
	}

	cmd := fmt.Sprintf(`snapshot gke %s --project p --exclude-namespaces-regex "^kube-,[invalid" --dry-run %s`, suite.envName, suite.defaultKosliArguments)
	_, combined, _, _, err := executeCommandC(cmd)

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "invalid --exclude-namespaces-regex pattern '[invalid'")
	require.False(suite.T(), listed, "an invalid pattern must be rejected before Asset Inventory is called")
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_WarnsWhenAssetInventoryReturnsNoPods() {
	newGKEClient = func(_ context.Context) (gkePodLister, error) {
		return stubGKEPodLister{}, nil
	}
	out := suite.runDryRun("--project p")
	require.Contains(suite.T(), out, "Asset Inventory returned no GKE pods in projects/p; reporting an empty snapshot to environment snapshot-gke-env")
	require.NotContains(suite.T(), out, "matched the cluster, location and namespace filters")
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_WarnsWhenFiltersSelectNoPods() {
	const warning = "none of the 4 pods in projects/p matched the cluster, location and namespace filters"
	require.Contains(suite.T(), suite.runDryRun("--project p --clusters nope"), warning)
	require.NotContains(suite.T(), suite.runDryRun("--project p"), warning)
}

// 2 Running pods and 1 Failed pod are reported; the Succeeded pod is not.
func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_HappyPathReportsToServer() {
	cmd := fmt.Sprintf(`snapshot gke %s --project p %s`, suite.envName, suite.defaultKosliArguments)
	_, combined, _, _, err := executeCommandC(cmd)

	require.NoError(suite.T(), err, "command failed: %s", combined)
	require.Contains(suite.T(), combined, fmt.Sprintf("[3] pods were reported to environment %s", suite.envName))
}

func (suite *SnapshotGKETestSuite) TestSnapshotGKECmd_PermissionDeniedReturnsFriendlyError() {
	newGKEClient = func(_ context.Context) (gkePodLister, error) {
		return stubGKEPodLister{err: status.Error(codes.PermissionDenied, "denied")}, nil
	}

	cmd := fmt.Sprintf(`snapshot gke %s --project p %s`, suite.envName, suite.defaultKosliArguments)
	_, combined, _, _, err := executeCommandC(cmd)

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `GCP permission denied: the caller needs 'cloudasset.assets.listContainerPod'`)
	require.Contains(suite.T(), combined, `"projects/p"`)
}

func TestSnapshotGKECommandTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotGKETestSuite))
}
