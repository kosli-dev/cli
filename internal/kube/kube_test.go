package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	clientset       *kubernetes.Clientset
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
	err = suite.provider.ExportKubeConfig(suite.clusterName, suite.kubeConfigPath)
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

	namespaces, err := suite.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: suite.namespacesLabel})
	require.NoErrorf(suite.T(), err, "error listing test namespaces with label %s", suite.namespacesLabel)

	for _, ns := range namespaces.Items {
		err = suite.clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{})
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
		namespaces      []string
		includePatterns []string
		excludePatterns []string
		pods            map[string][]*corev1.Pod
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
				namespaces:      []string{"empty-ns"},
				includePatterns: []string{"^empty-ns$"},
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
				includePatterns: []string{"^ns1$"},
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
				namespaces:      []string{"ns2", "ns3"},
				excludePatterns: []string{"^ns3$", "^default$", "^local-path-storage$", "^ns1$", "^kube-system$", "^empty-ns$", "^kube-node-lease$", "^kube-public$"},
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
			podsData, err := GetPodsData(t.args.includePatterns, t.args.excludePatterns, suite.clientset, logger.NewStandardLogger())
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

func (suite *KubeTestSuite) TestFilterNamespaces() {
	type args struct {
		nsList    []corev1.Namespace
		patterns  []string
		operation string
	}
	for _, t := range []struct {
		name        string
		args        args
		expectError bool
		want        []string
	}{
		{
			name: "unknown operation causes an error",
			args: args{
				nsList: []corev1.Namespace{
					{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
				},
				patterns:  []string{"^ns"},
				operation: "unknown",
			},
			expectError: true,
		},
		{
			name: "invalid patterns return error",
			args: args{
				nsList: []corev1.Namespace{
					{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
				},
				patterns:  []string{"["},
				operation: "exclude",
			},
			expectError: true,
			want:        []string{},
		},
		{
			name: "excluding when no patterns, returns all input namespaces",
			args: args{
				nsList: []corev1.Namespace{
					{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
				},
				patterns:  []string{},
				operation: "exclude",
			},
			expectError: false,
			want:        []string{"ns1", "ns2"},
		},
	} {
		suite.Run(t.name, func() {
			result, err := filterNamespaces(t.args.nsList, t.args.patterns, t.args.operation)
			if t.expectError {
				require.Error(suite.T(), err, "error was expected but got none.")
			} else {
				require.NoErrorf(suite.T(), err, "error was NOT expected but got: %v.", err)
				require.Equal(suite.T(), t.want, result, "TestFilterNamespaces: got %v -- want %v", result, t.want)
			}
		})
	}

}

// getK8sClient creates a k8s client set
func (suite *KubeTestSuite) getK8sClient(ctx context.Context) *kubernetes.Clientset {
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
	_, err := suite.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
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
	_, err := suite.clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	require.NoErrorf(suite.T(), err, "error creating pod %s", pod.Name)
	err = e2epod.WaitForPodNameRunningInNamespace(suite.clientset, pod.Name, namespace)
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
