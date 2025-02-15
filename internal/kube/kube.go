package kube

import (
	"context"
	"fmt"
	"sync"

	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/logger"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sEnvRequest represents the PUT request body to be sent to kosli from k8s
type K8sEnvRequest struct {
	Artifacts []*PodData `json:"artifacts"`
}

// PodData represents the harvested pod data
type PodData struct {
	PodName           string                  `json:"podName"`
	Namespace         string                  `json:"namespace"`
	Digests           map[string]string       `json:"digests"`
	CreationTimestamp int64                   `json:"creationTimestamp"`
	Owners            []metav1.OwnerReference `json:"owners"`
}

type K8SConnection struct {
	*kubernetes.Clientset
}

// NewPodData creates a PodData object from a k8s pod
func NewPodData(pod *corev1.Pod) *PodData {
	digests := make(map[string]string)

	creationTimestamp := pod.GetObjectMeta().GetCreationTimestamp()
	owners := pod.GetObjectMeta().GetOwnerReferences()
	containers := pod.Status.ContainerStatuses
	for _, cs := range containers {
		digests[cs.Image] = cs.ImageID[len(cs.ImageID)-64:]
	}

	return &PodData{
		PodName:           pod.Name,
		Namespace:         pod.Namespace,
		Digests:           digests,
		CreationTimestamp: creationTimestamp.Unix(),
		Owners:            owners,
	}
}

// NewK8sClientSet creates a k8s clientset
// if the kubeconfigPath is empty, it attempts to get an in-cluster client
func NewK8sClientSet(kubeconfigPath string) (*K8SConnection, error) {
	var config *rest.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("could not build config from flags: %v ", err)
		}
	} else {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("could not build config from inside the cluster: %v ", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8SConnection{clientset}, nil
}

// GetPodsData lists pods in the target namespace(s) of a target cluster and creates a list of
// PodData objects for them
func (clientset *K8SConnection) GetPodsData(filter *filters.ResourceFilterOptions, logger *logger.Logger) ([]*PodData, error) {
	var (
		podsData = []*PodData{}
		mutex    = &sync.Mutex{}
	)

	if len(filter.IncludeNames) == 0 && len(filter.IncludeNamesRegex) == 0 &&
		len(filter.ExcludeNames) == 0 && len(filter.ExcludeNamesRegex) == 0 {
		list, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return podsData, fmt.Errorf("could not list pods on cluster scope: %v ", err)
		}
		return processPods(list), nil
	} else {
		list := &corev1.PodList{}
		filteredNamespaces, err := clientset.filterNamespaces(filter)
		if err != nil {
			return podsData, fmt.Errorf("could not filter namespaces: %v ", err)
		}

		logger.Info("scanning the following namespaces: %v ", filteredNamespaces)

		g, _ := errgroup.WithContext(context.Background())

		for _, ns := range filteredNamespaces {
			g.Go(func() error {
				pods, err := clientset.getPodsInNamespace(ns)
				if err != nil {
					return err
				}
				mutex.Lock()
				list.Items = append(list.Items, pods...)
				mutex.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return nil, err
		}

		return processPods(list), nil
	}
}

// processPods returns podData list for a list of Pods
func processPods(list *corev1.PodList) []*PodData {
	podsData := []*PodData{}
	var (
		wg    sync.WaitGroup
		mutex = &sync.Mutex{}
	)
	for _, pod := range list.Items {
		wg.Add(1)
		go func(pod corev1.Pod) {
			defer wg.Done()
			// only report running or failed pods
			if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodFailed {
				data := NewPodData(&pod)
				mutex.Lock()
				podsData = append(podsData, data)
				mutex.Unlock()
			}
		}(pod)
	}
	wg.Wait()
	return podsData
}

// filterNamespaces filters a super set of namespaces by including or excluding a subset of namespaces using regex patterns.
func (clientset *K8SConnection) filterNamespaces(filter *filters.ResourceFilterOptions) ([]string, error) {
	if len(filter.IncludeNamesRegex) == 0 && len(filter.ExcludeNamesRegex) == 0 {
		if len(filter.IncludeNames) > 0 {
			return filter.IncludeNames, nil
		}
	}
	// get all namespaces in the cluster
	nsList, err := clientset.GetClusterNamespaces()
	if err != nil {
		return nil, err
	}

	namespaces := []string{}

	if len(filter.IncludeNames) == 0 && len(filter.IncludeNamesRegex) == 0 &&
		len(filter.ExcludeNames) == 0 && len(filter.ExcludeNamesRegex) == 0 {
		for _, ns := range nsList {
			namespaces = append(namespaces, ns.Name)
		}
		return namespaces, nil
	}

	mutex := new(sync.Mutex)
	g, _ := errgroup.WithContext(context.Background())

	for _, ns := range nsList {
		g.Go(func() error {
			include, err := filter.ShouldInclude(ns.Name)
			if err != nil {
				return err
			}
			if include {
				mutex.Lock()
				namespaces = append(namespaces, ns.Name)
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return namespaces, nil
}

// getPodsInNamespace get pods in a specific namespace in a cluster
func (clientset *K8SConnection) getPodsInNamespace(namespace string) ([]corev1.Pod, error) {
	ctx := context.Background()
	podlist, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []corev1.Pod{}, fmt.Errorf("could not list pods on namespace %s: %v ", namespace, err)
	}
	return podlist.Items, nil
}

// GetClusterNamespaces gets a namespace list from the cluster
func (clientset *K8SConnection) GetClusterNamespaces() ([]corev1.Namespace, error) {
	ctx := context.Background()

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []corev1.Namespace{}, fmt.Errorf("could not list namespaces on cluster scope: %v ", err)
	}

	return namespaces.Items, nil
}
