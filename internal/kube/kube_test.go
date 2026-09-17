package kube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	"sigs.k8s.io/kind/pkg/cluster"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type KubeTestSuite struct {
	suite.Suite
	tmpDir          string
	clusterName     string
	provider        *cluster.Provider
	kubeConfigPath  string
	namespacesLabel string
	clientset       *K8SConnection
}

// create a KIND cluster and a tmp dir before the suite execution
func (suite *KubeTestSuite) SetupSuite() {
	suite.clusterName = "test"
	suite.namespacesLabel = fmt.Sprintf("suite=%s", suite.clusterName)
	createOptions := cluster.CreateWithWaitForReady(300 * time.Second)
	suite.provider = cluster.NewProvider(cluster.ProviderWithDocker())
	err := suite.provider.Create(suite.clusterName, createOptions)
	require.NoError(suite.T(), err, "creating test k8s cluster failed")
	suite.tmpDir, err = os.MkdirTemp("", "testDir")
	require.NoError(suite.T(), err, "error creating a temporary test directory")
	suite.kubeConfigPath = filepath.Join(suite.tmpDir, "kubeconfig")
	err = suite.provider.ExportKubeConfig(suite.clusterName, suite.kubeConfigPath, false) // false = don't use internal cluster IPs
	require.NoError(suite.T(), err, "exporting kubeconfig failed")
	ctx := context.Background()
	suite.clientset = suite.getK8sClient(ctx)
}

// delete the KIND cluster and the tmp dir after the suite execution
func (suite *KubeTestSuite) TearDownSuite() {
	err := suite.provider.Delete(suite.clusterName, suite.kubeConfigPath)
	require.NoError(suite.T(), err, "deleting KIND cluster failed")
	err = os.RemoveAll(suite.tmpDir)
	require.NoErrorf(suite.T(), err, "error cleaning up the temporary test directory %s", suite.tmpDir)
}

func (suite *KubeTestSuite) AfterTest(_, _ string) {
	ctx := context.Background()

	namespaces, err := suite.clientset.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: suite.namespacesLabel})
	require.NoErrorf(suite.T(), err, "error listing test namespaces with label %s", suite.namespacesLabel)

	for _, ns := range namespaces.Items {
		err = suite.clientset.Clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		require.NoErrorf(suite.T(), err, "error deleting namespace %s", ns.Name)
	}
}

func (suite *KubeTestSuite) TestGetPodsData() {
	type comparablePodData struct {
		podName   string
		namespace string
		digests   map[string]string
	}
	type args struct {
		namespaces []string
		filter     *filters.ResourceFilterOptions
		pods       map[string][]*corev1.Pod
	}
	for _, t := range []struct {
		name            string
		args            args
		want            []*comparablePodData
		wantSubset      []*comparablePodData
		wantMinimumPods int
	}{
		{
			name: "an empty namespace has nothing to report.",
			args: args{
				filter: &filters.ResourceFilterOptions{
					IncludeNamesRegex: []string{"^empty-ns$"},
				},
				namespaces: []string{"empty-ns"},
			},
			want: []*comparablePodData{},
		},
		{
			name: "a namespace with one pod reports one pod data.",
			args: args{
				namespaces: []string{"ns1"},
				pods: map[string][]*corev1.Pod{
					"ns1": {suite.getPodPayload("pod1", []string{"nginx:1.21.3"})},
				},
				filter: &filters.ResourceFilterOptions{
					IncludeNamesRegex: []string{"^ns1$"},
				},
			},
			want: []*comparablePodData{
				{
					podName:   "pod1",
					namespace: "ns1",
					digests: map[string]string{
						"docker.io/library/nginx:1.21.3": "644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36",
					},
				},
			},
		},
		{
			name: "excluding a namespace works.",
			args: args{
				namespaces: []string{"ns2", "ns3"},
				filter: &filters.ResourceFilterOptions{
					ExcludeNamesRegex: []string{"^ns3$", "^default$", "^local-path-storage$", "^ns1$", "^kube-system$", "^empty-ns$", "^kube-node-lease$", "^kube-public$"},
				},
				pods: map[string][]*corev1.Pod{
					"ns2": {suite.getPodPayload("nginx1", []string{"nginx:1.21.3"})},
					"ns3": {suite.getPodPayload("nginx2", []string{"nginx:1.21.0"})},
				},
			},
			want: []*comparablePodData{
				{
					podName:   "nginx1",
					namespace: "ns2",
					digests: map[string]string{
						"docker.io/library/nginx:1.21.3": "644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36",
					},
				},
			},
		},
		{
			name: "not excluding nor including namespaces reports all cluster pods.",
			args: args{
				namespaces: []string{"ns4"},
				pods: map[string][]*corev1.Pod{
					"ns4": {suite.getPodPayload("nginx1", []string{"nginx:1.21.3"})},
				},
				filter: &filters.ResourceFilterOptions{},
			},
			wantSubset: []*comparablePodData{
				{
					podName:   "nginx1",
					namespace: "ns4",
					digests: map[string]string{
						"docker.io/library/nginx:1.21.3": "644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36",
					},
				},
			},
			wantMinimumPods: 2,
		},
	} {
		suite.Run(t.name, func() {
			// create namespaces
			for _, ns := range t.args.namespaces {
				suite.createNamespace(ns)
			}
			// create pods
			for ns, pods := range t.args.pods {
				for _, pod := range pods {
					suite.createPod(ns, pod)
				}
			}
			// Get pods data
			podsData, err := suite.clientset.GetPodsData(t.args.filter, logger.NewStandardLogger())
			require.NoErrorf(suite.T(), err, "error getting pods data for test %s", t.name)
			actual := []*comparablePodData{}
			for _, pd := range podsData {
				actual = append(actual, &comparablePodData{
					podName:   pd.PodName,
					namespace: pd.Namespace,
					digests:   pd.Digests,
				})
			}
			if len(t.want) > 0 {
				require.Equal(suite.T(), t.want, actual, fmt.Sprintf("want: %v -- got: %v", t.want, actual))
			} else if len(t.wantSubset) > 0 {
				require.Subset(suite.T(), actual, t.wantSubset)
				require.GreaterOrEqual(suite.T(), len(actual), t.wantMinimumPods)
			}
		})
	}
}

func (suite *KubeTestSuite) TestGetPodsDataWithThrottling() {
	// create a large number of pods
	for i := 0; i < 200; i++ {
		suite.createNamespace(fmt.Sprintf("ns-%d", i))
	}
	// Get pods data with timeout check
	startTime := time.Now()
	_, err := suite.clientset.GetPodsData(&filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^ns-.*"}}, logger.NewStandardLogger())
	duration := time.Since(startTime)
	require.NoErrorf(suite.T(), err, "error getting pods data for test GetPodsDataWithThrottling")
	require.LessOrEqual(suite.T(), duration, 5*time.Second, "GetPodsData should complete within 5 seconds, but took %v", duration)
}

func (suite *KubeTestSuite) TestFilterNamespaces() {
	type args struct {
		namespaces []string
		filter     *filters.ResourceFilterOptions
	}
	for _, t := range []struct {
		name         string
		args         args
		expectError  bool
		want         []string
		wantFiltered []string
	}{
		{
			// Creates no namespaces: the invalid pattern is reported as soon as filtering
			// reaches it, and the cluster's own namespaces (default, kube-system, ...)
			// are enough to reach it. Any created here would be a create and a delete
			// against a real cluster for nothing.
			name: "invalid regex patterns return error",
			args: args{
				filter: &filters.ResourceFilterOptions{
					IncludeNamesRegex: []string{"["},
				},
			},
			expectError: true,
		},
		{
			name: "namespaces matching the include regex patterns are returned",
			args: args{
				namespaces: []string{"filter-inc-a1", "filter-inc-a2", "filter-inc-b1"},
				filter: &filters.ResourceFilterOptions{
					IncludeNamesRegex: []string{"^filter-inc-a.*$"},
				},
			},
			want: []string{"filter-inc-a1", "filter-inc-a2"},
		},
		{
			name: "namespaces matching the exclude regex patterns are filtered out",
			args: args{
				namespaces: []string{"filter-exc-a1", "filter-exc-a2", "filter-exc-b1"},
				filter: &filters.ResourceFilterOptions{
					ExcludeNamesRegex: []string{"^filter-exc-a.*$"},
				},
			},
			wantFiltered: []string{"filter-exc-a1", "filter-exc-a2"},
		},
		{
			// An invalid pattern is only reported once matching a name reaches it, and a
			// name listed in ExcludeNames is settled before that. On this path the
			// short-circuit never saves the snapshot though: the cluster's own
			// namespaces (default, kube-system, ...) are not in ExcludeNames, so one of
			// them always reaches the pattern.
			name: "an invalid exclude pattern is reported despite the excluded literal name",
			args: args{
				namespaces: []string{"filter-lazy-a1"},
				filter: &filters.ResourceFilterOptions{
					ExcludeNames:      []string{"filter-lazy-a1"},
					ExcludeNamesRegex: []string{"["},
				},
			},
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			// namespace names must not be shared with another test method: AfterTest
			// only asks for deletion, and a namespace lingers in Terminating for a
			// while after that, so re-creating one by the same name fails with
			// "object is being deleted". Hence the filter- prefix here.
			for _, ns := range t.args.namespaces {
				suite.createNamespace(ns)
			}
			result, err := suite.clientset.filterNamespaces(t.args.filter)
			if t.expectError {
				require.Error(suite.T(), err, "error was expected but got none.")
				return
			}
			require.NoErrorf(suite.T(), err, "error was NOT expected but got: %v.", err)
			if len(t.wantFiltered) > 0 {
				// every namespace this case created and did not ask to be filtered out
				// has to survive, so the case describes its own expectation instead of
				// naming a survivor twice
				survivors := []string{}
				for _, ns := range t.args.namespaces {
					if !slices.Contains(t.wantFiltered, ns) {
						survivors = append(survivors, ns)
					}
				}
				require.Subset(suite.T(), result, survivors,
					"TestFilterNamespaces: %v should contain every non-excluded namespace %v", result, survivors)
				for _, ns := range t.wantFiltered {
					require.NotContains(suite.T(), result, ns,
						"TestFilterNamespaces: %s should have been filtered out of %v", ns, result)
				}
				return
			}
			// filterNamespaces returns the namespaces in the order the cluster listed
			// them, which the goroutine-per-namespace fan-out it replaced could not
			// guarantee. Build the expectation in that same order so the assertion
			// pins the ordering and not just the set.
			nsList, err := suite.clientset.GetClusterNamespaces()
			require.NoErrorf(suite.T(), err, "error listing cluster namespaces")
			wantInClusterOrder := []string{}
			for _, ns := range nsList {
				if slices.Contains(t.want, ns.Name) {
					wantInClusterOrder = append(wantInClusterOrder, ns.Name)
				}
			}
			require.ElementsMatch(suite.T(), t.want, wantInClusterOrder,
				"TestFilterNamespaces: %v missing from the cluster listing, the ordering assertion would be vacuous", t.want)
			require.Equal(suite.T(), wantInClusterOrder, result, "TestFilterNamespaces: got %v -- want %v", result, wantInClusterOrder)
		})
	}

}

// getK8sClient creates a k8s client set
func (suite *KubeTestSuite) getK8sClient(ctx context.Context) *K8SConnection {
	clientset, err := NewK8sClientSet(suite.kubeConfigPath)
	require.NoErrorf(suite.T(), err, "error creating k8s client set for kubeconfig %s", suite.kubeConfigPath)
	return clientset
}

// createNamespace creates a namespace in the suite KIND cluster
func (suite *KubeTestSuite) createNamespace(name string) {
	ctx := context.Background()
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"suite": suite.clusterName,
			},
		},
	}
	_, err := suite.clientset.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	require.NoErrorf(suite.T(), err, "error creating namespace %s", name)
}

// getPodPayload creates a k8s Pod struct
func (suite *KubeTestSuite) getPodPayload(name string, images []string) *corev1.Pod {
	podContainers := []corev1.Container{}
	for i, image := range images {
		podContainers = append(podContainers, corev1.Container{
			Name:  fmt.Sprintf("container-%d", i),
			Image: image,
		})
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"suite": suite.clusterName,
			},
		},
		Spec: corev1.PodSpec{
			Containers: podContainers,
		},
	}
}

// createPod creates a pod in the suite KIND cluster
func (suite *KubeTestSuite) createPod(namespace string, pod *corev1.Pod) {
	ctx := context.Background()
	_, err := suite.clientset.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	require.NoErrorf(suite.T(), err, "error creating pod %s", pod.Name)
	err = e2epod.WaitForPodNameRunningInNamespace(ctx, suite.clientset, pod.Name, namespace)
	require.NoErrorf(suite.T(), err, "error waiting for pod %s to be running in namespace %s", pod.Name, namespace)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKubeTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	suite.Run(t, new(KubeTestSuite))
}

// TestProcessPodsWithFailedPodsWithoutImageIDs tests that failed pods without image IDs
// are skipped and do not result in nil entries in the returned slice
// This relates to issue https://github.com/kosli-dev/server/issues/4448
func TestProcessPodsWithFailedPodsWithoutImageIDs(t *testing.T) {
	testLogger := logger.NewStandardLogger()

	// Create a mix of pods: running, failed with image IDs, and failed without image IDs
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "running-pod",
					Namespace: "test-ns",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "container1",
							Image:   "nginx:1.21.3",
							ImageID: "docker-pullable://nginx@sha256:644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36",
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed-pod-without-imageid",
					Namespace: "test-ns",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "container1",
							Image:   "nginx:1.21.3",
							ImageID: "", // Empty ImageID - should be skipped
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "another-running-pod",
					Namespace: "test-ns",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "container1",
							Image:   "nginx:1.21.0",
							ImageID: "docker-pullable://nginx@sha256:123a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36",
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "another-failed-pod-without-imageid",
					Namespace: "test-ns",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:    "container1",
							Image:   "busybox:latest",
							ImageID: "", // Empty ImageID - should be skipped
						},
					},
				},
			},
		},
	}

	result, err := processPods(pods, testLogger)
	require.NoError(t, err, "processPods should not return an error")

	// We should only get 2 pods (the two running ones), not 4
	require.Equal(t, 2, len(result), "Expected only running pods to be included")

	// Verify no nil entries
	for i, podData := range result {
		require.NotNil(t, podData, "Pod data at index %d should not be nil", i)
	}

	// Verify the correct pods are included
	podNames := make([]string, len(result))
	for i, podData := range result {
		podNames[i] = podData.PodName
	}
	require.Contains(t, podNames, "running-pod")
	require.Contains(t, podNames, "another-running-pod")
	require.NotContains(t, podNames, "failed-pod-without-imageid")
	require.NotContains(t, podNames, "another-failed-pod-without-imageid")
}

// containerStatus is the (image, imageID) pair kubelet reports per container. An empty
// imageID models a container the kubelet has not reported an image for yet.
type containerStatus struct{ image, imageID string }

// podWithStatuses builds a pod in the given phase with one container status per
// containerStatus, in order.
func podWithStatuses(name string, phase corev1.PodPhase, statuses ...containerStatus) corev1.Pod {
	containerStatuses := make([]corev1.ContainerStatus, 0, len(statuses))
	for i, s := range statuses {
		containerStatuses = append(containerStatuses, corev1.ContainerStatus{
			Name:    fmt.Sprintf("container-%d", i),
			Image:   s.image,
			ImageID: s.imageID,
		})
	}
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-ns"},
		Status:     corev1.PodStatus{Phase: phase, ContainerStatuses: containerStatuses},
	}
}

// podWithSpec builds a Running pod whose containers have a spec image as well as a
// status, which is what a digest-pinned pod looks like. specImages are matched to
// statuses by position, as container-N.
func podWithSpec(name string, specImages []string, statuses ...containerStatus) corev1.Pod {
	pod := podWithStatuses(name, corev1.PodRunning, statuses...)
	containers := make([]corev1.Container, 0, len(specImages))
	for i, image := range specImages {
		containers = append(containers, corev1.Container{
			Name:  fmt.Sprintf("container-%d", i),
			Image: image,
		})
	}
	pod.Spec = corev1.PodSpec{Containers: containers}
	return pod
}

const (
	nginxImageID   = "docker-pullable://nginx@sha256:644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36"
	busyboxImageID = "docker-pullable://busybox@sha256:123a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36"
)

// TestNewPodData covers containers without an image ID, which is a transient state on
// a healthy cluster (image still pulling, kubelet status not yet populated) and must
// never abort the snapshot. See https://github.com/kosli-dev/cli/issues/1194
func TestNewPodData(t *testing.T) {
	const nginxSha = "644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36"
	for _, tc := range []struct {
		name        string
		pod         corev1.Pod
		wantDigests map[string]string
		wantSkipped bool
		wantWarning string
	}{
		{
			name:        "a Running pod whose only container has no image ID is skipped, not an error",
			pod:         podWithStatuses("pod", corev1.PodRunning, containerStatus{"nginx:1.21.3", ""}),
			wantSkipped: true,
			wantWarning: "skipping Running pod pod in namespace test-ns as none of its containers has a usable image ID: container-0 (empty image ID)",
		},
		{
			name:        "a Failed pod whose only container has no image ID is skipped",
			pod:         podWithStatuses("pod", corev1.PodFailed, containerStatus{"nginx:1.21.3", ""}),
			wantSkipped: true,
			wantWarning: "skipping Failed pod pod in namespace test-ns",
		},
		{
			name:        "a Running pod whose only container has a malformed image ID is skipped, not an error",
			pod:         podWithStatuses("pod", corev1.PodRunning, containerStatus{"nginx:1.21.3", "sha256:abc"}),
			wantSkipped: true,
			wantWarning: "skipping Running pod pod in namespace test-ns",
		},
		{
			name:        "a Running pod with no container statuses yet is skipped",
			pod:         podWithStatuses("pod", corev1.PodRunning),
			wantSkipped: true,
			wantWarning: "skipping Running pod pod in namespace test-ns as none of its containers has a usable image ID\n",
		},
		{
			name: "a Running pod is reported with the digests of the containers that have an image ID",
			pod: podWithStatuses("pod", corev1.PodRunning,
				containerStatus{"nginx:1.21.3", nginxImageID},
				containerStatus{"busybox:latest", ""}),
			wantDigests: map[string]string{"nginx:1.21.3": nginxSha},
			wantWarning: "Running pod pod in namespace test-ns has containers without a usable image ID, reporting it without them: container-1 (empty image ID)",
		},
		{
			name: "a Failed pod is reported with the digests of the containers that have an image ID",
			pod: podWithStatuses("pod", corev1.PodFailed,
				containerStatus{"busybox:latest", ""},
				containerStatus{"nginx:1.21.3", nginxImageID}),
			wantDigests: map[string]string{"nginx:1.21.3": nginxSha},
			wantWarning: "reporting it without them: container-0 (empty image ID)",
		},
		{
			name:        "a pod whose containers all have an image ID is reported without a warning",
			pod:         podWithStatuses("pod", corev1.PodRunning, containerStatus{"nginx:1.21.3", nginxImageID}),
			wantDigests: map[string]string{"nginx:1.21.3": nginxSha},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var warnings bytes.Buffer
			got, err := NewPodData(&tc.pod, logger.NewLogger(io.Discard, &warnings, false))
			require.NoError(t, err)
			if tc.wantWarning == "" {
				require.Empty(t, warnings.String())
			} else {
				require.Contains(t, warnings.String(), tc.wantWarning)
			}
			if tc.wantSkipped {
				require.Nil(t, got, "expected the pod to be skipped")
				return
			}
			require.NotNil(t, got, "expected the pod to be reported")
			require.Equal(t, tc.wantDigests, got.Digests)
		})
	}
}

// TestNewPodDataArtifactName covers the name an artifact is reported under when the
// runtime has no tagged name for the image, which happens whenever a digest-pinned
// image is pulled on a node that does not already hold it under a tag.
// See https://github.com/kosli-dev/cli/issues/1203
func TestNewPodDataArtifactName(t *testing.T) {
	const (
		nginxSha    = "644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36"
		busyboxSha  = "123a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36"
		imageID     = "sha256:8dd77ef2d82eade8dcf2c08ea032bd9cba04c9d28ace2ccf08ad6804c27bf14f"
		imageIDSha  = "8dd77ef2d82eade8dcf2c08ea032bd9cba04c9d28ace2ccf08ad6804c27bf14f"
		nginxDigest = "nginx:1.25@sha256:644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36"
	)
	for _, tc := range []struct {
		name        string
		pod         corev1.Pod
		wantDigests map[string]string
	}{
		{
			name:        "the runtime's name is used when it has one",
			pod:         podWithSpec("pod", []string{nginxDigest}, containerStatus{"docker.io/library/nginx:1.25", nginxImageID}),
			wantDigests: map[string]string{"docker.io/library/nginx:1.25": nginxSha},
		},
		{
			name:        "a bare image ID falls back to the spec image, normalized to its tag",
			pod:         podWithSpec("pod", []string{nginxDigest}, containerStatus{imageID, nginxImageID}),
			wantDigests: map[string]string{"docker.io/library/nginx:1.25": nginxSha},
		},
		{
			name:        "the fallback name matches what the runtime reports for the same image",
			pod:         podWithSpec("pod", []string{"nginx:1.25"}, containerStatus{imageID, nginxImageID}),
			wantDigests: map[string]string{"docker.io/library/nginx:1.25": nginxSha},
		},
		{
			name:        "a spec image with no tag falls back to the image ID, not to :latest",
			pod:         podWithSpec("pod", []string{"nginx@sha256:" + nginxSha}, containerStatus{imageID, nginxImageID}),
			wantDigests: map[string]string{"nginx@sha256:" + nginxSha: nginxSha},
		},
		{
			name:        "a bare image ID with no spec containers at all falls back to the image ID",
			pod:         podWithStatuses("pod", corev1.PodRunning, containerStatus{imageID, nginxImageID}),
			wantDigests: map[string]string{"nginx@sha256:" + nginxSha: nginxSha},
		},
		{
			name: "a spec container whose name does not match is not used",
			pod: func() corev1.Pod {
				pod := podWithStatuses("pod", corev1.PodRunning, containerStatus{imageID, nginxImageID})
				pod.Spec = corev1.PodSpec{Containers: []corev1.Container{
					{Name: "sidecar", Image: "busybox:1.36"},
				}}
				return pod
			}(),
			wantDigests: map[string]string{"nginx@sha256:" + nginxSha: nginxSha},
		},
		{
			name:        "a spec image that is itself a bare digest falls back to the image ID",
			pod:         podWithSpec("pod", []string{imageID}, containerStatus{imageID, nginxImageID}),
			wantDigests: map[string]string{"nginx@sha256:" + nginxSha: nginxSha},
		},
		{
			// a bare-digest ImageID, so this reaches the last fallback rather than
			// returning at the isNamed(imageID) check above it
			name:        "a container status with no name and no named image ID is reported under the digest",
			pod:         podWithStatuses("pod", corev1.PodRunning, containerStatus{"", imageID}),
			wantDigests: map[string]string{imageID: imageIDSha},
		},
		{
			name: "two digest-pinned containers keep distinct names",
			pod: podWithSpec("pod",
				[]string{nginxDigest, "busybox:1.36@sha256:" + busyboxSha},
				containerStatus{imageID, nginxImageID},
				containerStatus{"sha256:99aa70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36", busyboxImageID}),
			wantDigests: map[string]string{
				"docker.io/library/nginx:1.25":   nginxSha,
				"docker.io/library/busybox:1.36": busyboxSha,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var warnings bytes.Buffer
			got, err := NewPodData(&tc.pod, logger.NewLogger(io.Discard, &warnings, false))
			require.NoError(t, err)
			require.NotNil(t, got, "expected the pod to be reported")
			require.Equal(t, tc.wantDigests, got.Digests)
			require.Empty(t, warnings.String())
		})
	}
}

// TestProcessPodsWithRunningPodWithoutImageID checks that one Running pod with an empty
// image ID does not abort the snapshot for every other pod.
// See https://github.com/kosli-dev/cli/issues/1194
func TestProcessPodsWithRunningPodWithoutImageID(t *testing.T) {
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			podWithStatuses("running-pod", corev1.PodRunning, containerStatus{"nginx:1.21.3", nginxImageID}),
			podWithStatuses("running-pod-without-imageid", corev1.PodRunning, containerStatus{"nginx:1.21.3", ""}),
			podWithStatuses("another-running-pod", corev1.PodRunning, containerStatus{"busybox:latest", busyboxImageID}),
		},
	}

	result, err := processPods(pods, logger.NewStandardLogger())
	require.NoError(t, err, "a Running pod without an image ID must not abort the snapshot")

	podNames := []string{}
	for _, podData := range result {
		require.NotNil(t, podData)
		podNames = append(podNames, podData.PodName)
	}
	require.ElementsMatch(t, []string{"running-pod", "another-running-pod"}, podNames)
}

func TestImageFingerprint(t *testing.T) {
	const sha = "644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36"
	for _, tc := range []struct {
		name    string
		imageID string
		want    string
	}{
		{name: "dockershim reference with digest", imageID: "docker-pullable://nginx@sha256:" + sha, want: sha},
		{name: "dockershim bare digest", imageID: "docker://sha256:" + sha, want: sha},
		{name: "containerd reference with digest", imageID: "docker.io/library/nginx@sha256:" + sha, want: sha},
		{name: "reference with tag and digest", imageID: "ghcr.io/org/app:v1@sha256:" + sha, want: sha},
		{name: "bare digest of a locally loaded image", imageID: "sha256:" + sha, want: sha},
		{name: "empty", imageID: ""},
		{name: "too short to hold a digest", imageID: "sha256:abc"},
		{name: "reference without a digest", imageID: "nginx:1.21.3"},
		{name: "sha512 digest", imageID: "sha512:" + sha + sha},
		{name: "upper-case hex", imageID: "sha256:" + strings.ToUpper(sha)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := imageFingerprint(tc.imageID)
			if tc.want == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
