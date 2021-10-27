package kube

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
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
	suite.tmpDir, err = ioutil.TempDir("", "testDir")
	require.NoError(suite.T(), err, "error creating a temporary test directory")
	suite.kubeConfigPath = filepath.Join(suite.tmpDir, "kubeconfig")
	err = suite.provider.ExportKubeConfig(suite.clusterName, suite.kubeConfigPath)
	require.NoError(suite.T(), err, "exporting kubeconfig failed")
	ctx := context.Background()
	suite.clientset = suite.GetK8sClient(ctx)
	framework.ExpectNoError(framework.WaitForAllNodesSchedulable(suite.clientset, framework.TestContext.NodeSchedulableTimeout))
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
		namespaces []string
		pods       map[string][]*corev1.Pod
	}
	for _, t := range []struct {
		name string
		args args
		want []*comparablePodData
	}{
		{
			name: "an empty namespace has nothing to report.",
			args: args{
				namespaces: []string{"ns1"},
				pods: map[string][]*corev1.Pod{
					"ns1": {suite.GetPodPayload("pod1", []string{"nginx"})},
				},
			},
			want: []*comparablePodData{
				// {
				// 	podName:   "",
				// 	namespace: "ns1",
				// 	digests: map[string]string{
				// 		"": "",
				// 	},
				// },
			},
		},
	} {
		suite.Run(t.name, func() {
			// create namespaces
			for _, ns := range t.args.namespaces {
				fmt.Printf("creating ns %s \n", ns)
				suite.CreateNamespace(ns)
			}
			// create pods
			for ns, pods := range t.args.pods {
				for _, pod := range pods {
					fmt.Printf("creating pod %s \n", pod.Name)
					suite.CreatePod(ns, pod)
				}
			}
			// Get pods data
			podsData, err := GetPodsData(t.args.namespaces, []string{}, suite.clientset)
			require.NoErrorf(suite.T(), err, "error getting pods data for test %s", t.name)
			actual := []*comparablePodData{}
			for _, pd := range podsData {
				actual = append(actual, &comparablePodData{
					podName:   pd.PodName,
					namespace: pd.Namespace,
					digests:   pd.Digests,
				})
			}
			fmt.Printf("actual %v \n", actual)
			require.Equal(suite.T(), t.want, actual, fmt.Sprintf("want: %v -- got: %v", t.want, actual))

		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKubeTestSuite(t *testing.T) {
	suite.Run(t, new(KubeTestSuite))
}

// GetK8sClient creates a k8s client set
func (suite *KubeTestSuite) GetK8sClient(ctx context.Context) *kubernetes.Clientset {
	clientset, err := NewK8sClientSet(suite.kubeConfigPath)
	require.NoErrorf(suite.T(), err, "error creating k8s client set for kubeconfig %s", suite.kubeConfigPath)
	return clientset
}

// CreateNamespace creates a namespace in the suite KIND cluster
func (suite *KubeTestSuite) CreateNamespace(name string) {
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

// GetPodPayload creates a k8s Pod struct
func (suite *KubeTestSuite) GetPodPayload(name string, images []string) *corev1.Pod {
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

// CreatePod creates a pod in the suite KIND cluster
func (suite *KubeTestSuite) CreatePod(namespace string, pod *corev1.Pod) {
	ctx := context.Background()
	_, err := suite.clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	require.NoErrorf(suite.T(), err, "error creating pod %s", pod.Name)
}
