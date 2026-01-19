package kube

import (
	"context"
	"fmt"
	"sync"

	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	clientQPS   = 50
	clientBurst = 100
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
func NewPodData(pod *corev1.Pod, logger *logger.Logger) (*PodData, error) {
	digests := make(map[string]string)

	creationTimestamp := pod.GetObjectMeta().GetCreationTimestamp()
	owners := pod.GetObjectMeta().GetOwnerReferences()
	containers := pod.Status.ContainerStatuses

	for _, cs := range containers {
		if cs.ImageID == "" {
			switch pod.Status.Phase {
			case corev1.PodFailed:
				// skip failed pods
				logger.Warn("skipping failed pod %s in namespace %s as it has containers without image IDs", pod.Name, pod.Namespace)
				return nil, nil
			case corev1.PodRunning:
				// fail
				return nil, fmt.Errorf("pod %s in namespace %s has containers without image IDs", pod.Name, pod.Namespace)
			}
		} else {
			digests[cs.Image] = cs.ImageID[len(cs.ImageID)-64:]
		}
	}

	return &PodData{
		PodName:           pod.Name,
		Namespace:         pod.Namespace,
		Digests:           digests,
		CreationTimestamp: creationTimestamp.Unix(),
		Owners:            owners,
	}, nil
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

	// set the QPS and burst for the config to control the rate of requests to the API server
	//  defaults are 5 QPS and 10 burst which is too low for large clusters
	config.QPS = clientQPS
	config.Burst = clientBurst

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
		wg       sync.WaitGroup
		mutex    = &sync.Mutex{}
	)

	if len(filter.IncludeNames) == 0 && len(filter.IncludeNamesRegex) == 0 &&
		len(filter.ExcludeNames) == 0 && len(filter.ExcludeNamesRegex) == 0 {
		list, err := clientset.Clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return podsData, fmt.Errorf("could not list pods on cluster scope: %v ", err)
		}
		return processPods(list, logger)
	} else {
		list := &corev1.PodList{}
		filteredNamespaces, err := clientset.filterNamespaces(filter)
		if err != nil {
			return podsData, fmt.Errorf("could not filter namespaces: %v ", err)
		}

		logger.Info("scanning the following namespaces: %v ", filteredNamespaces)

		// run concurrently
		errs := make(chan error, 1) // Buffered only for the first error
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // Make sure it's called to release resources even if no errors

		// semaphore to limit the number of concurrent requests to the API server
		sem := make(chan struct{}, (clientBurst/2)-1) // max concurrent requests: slightly lower than client's QPS to avoid throttling

		for _, ns := range filteredNamespaces {
			wg.Add(1)
			go func(ns string) {
				defer wg.Done()
				sem <- struct{}{}        // acquire the semaphore
				defer func() { <-sem }() // release the semaphore
				// Check if any error occurred in any other gorouties:
				select {
				case <-ctx.Done():
					return // Error somewhere, terminate
				default: // Default is must to avoid blocking
				}

				pods, err := clientset.getPodsInNamespace(ns)
				if err != nil {
					// Non-blocking send of error
					select {
					case errs <- err:
					default:
					}
					cancel() // send cancel signal to goroutines
					return
				}
				mutex.Lock()
				list.Items = append(list.Items, pods...)
				mutex.Unlock()

			}(ns)
		}

		wg.Wait()
		// Return (first) error, if any:
		if ctx.Err() != nil {
			return podsData, <-errs
		}

		return processPods(list, logger)
	}
}

// processPods returns podData list for a list of Pods
func processPods(list *corev1.PodList, logger *logger.Logger) ([]*PodData, error) {
	podsData := []*PodData{}
	var (
		wg    sync.WaitGroup
		mutex = &sync.Mutex{}
	)

	errs := make(chan error, 1) // Buffered only for the first error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure it's called to release resources even if no errors

	for _, pod := range list.Items {
		wg.Add(1)
		go func(pod corev1.Pod) {
			defer wg.Done()

			// Check if any error occurred in any other goroutines:
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
			}

			// only report running or failed pods
			if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodFailed {
				data, err := NewPodData(&pod, logger)
				if err != nil {
					// Non-blocking send of error
					select {
					case errs <- err:
					default:
					}
					cancel() // send cancel signal to goroutines
					return
				}
				// Only append if data is not nil (NewPodData returns nil for skipped pods)
				if data != nil {
					mutex.Lock()
					podsData = append(podsData, data)
					mutex.Unlock()
				}
			}
		}(pod)
	}
	wg.Wait()
	// Return (first) error, if any:
	if ctx.Err() != nil {
		return podsData, <-errs
	}
	return podsData, nil
}

// filterNamespaces filters a super set of namespaces by including or excluding a subset of namespaces using regex patterns.
func (clientset *K8SConnection) filterNamespaces(filter *filters.ResourceFilterOptions) ([]string, error) {
	if len(filter.IncludeNamesRegex) == 0 && len(filter.ExcludeNamesRegex) == 0 {
		if len(filter.IncludeNames) > 0 {
			return filter.IncludeNames, nil
		}
	}
	result := []string{}
	// get all namespaces in the cluster
	nsList, err := clientset.GetClusterNamespaces()
	if err != nil {
		return result, err
	}

	if len(filter.IncludeNames) == 0 && len(filter.IncludeNamesRegex) == 0 &&
		len(filter.ExcludeNames) == 0 && len(filter.ExcludeNamesRegex) == 0 {
		for _, ns := range nsList {
			result = append(result, ns.Name)
		}
		return result, nil
	}

	var (
		wg    sync.WaitGroup
		mutex = &sync.Mutex{}
	)

	errs := make(chan error, 1) // Buffered only for the first error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure it's called to release resources even if no errors

	for _, ns := range nsList {
		wg.Add(1)
		go func(ns string) {
			defer wg.Done()

			// Check if any error occurred in any other gorouties:
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
			}

			include, err := filter.ShouldInclude(ns)
			if err != nil {
				select {
				case errs <- err:
				default:
				}
				cancel() // send cancel signal to goroutines
				return
			}
			if include {
				mutex.Lock()
				result = append(result, ns)
				mutex.Unlock()
			}
		}(ns.Name)
	}
	wg.Wait()
	if ctx.Err() != nil {
		return result, <-errs
	}
	return result, nil
}

// getPodsInNamespace get pods in a specific namespace in a cluster
func (clientset *K8SConnection) getPodsInNamespace(namespace string) ([]corev1.Pod, error) {
	ctx := context.Background()
	podlist, err := clientset.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []corev1.Pod{}, fmt.Errorf("could not list pods on namespace %s: %v ", namespace, err)
	}
	return podlist.Items, nil
}

// GetClusterNamespaces gets a namespace list from the cluster
func (clientset *K8SConnection) GetClusterNamespaces() ([]corev1.Namespace, error) {
	ctx := context.Background()

	namespaces, err := clientset.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []corev1.Namespace{}, fmt.Errorf("could not list namespaces on cluster scope: %v ", err)
	}

	return namespaces.Items, nil
}
