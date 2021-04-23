package kube

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// PodData represents the harvested pod data
type PodData struct {
	PodName           string                  `json:"podName"`
	Namespace         string                  `json:"namespace"`
	Images            map[string]string       `json:"images"`
	CreationTimestamp metav1.Time             `json:"creationTimestamp"`
	Owners            []metav1.OwnerReference `json:"owners"`
}

// NewPodData creates a PodData object from a k8s pod
func NewPodData(pod *corev1.Pod) *PodData {
	images := make(map[string]string)

	creationTimestamp := pod.GetObjectMeta().GetCreationTimestamp()
	owners := pod.GetObjectMeta().GetOwnerReferences()
	containers := pod.Status.ContainerStatuses
	for _, cs := range containers {
		images[cs.Image] = cs.ImageID[len(cs.ImageID)-64:]
	}

	return &PodData{
		PodName:           pod.Name,
		Namespace:         pod.Namespace,
		Images:            images,
		CreationTimestamp: creationTimestamp,
		Owners:            owners,
	}
}

// NewK8sClientSet creates a k8s clientset
// if the kubeconfigPath is empty, it attempts to get an in-cluster client
func NewK8sClientSet(kubeconfigPath string) (*kubernetes.Clientset, error) {
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

	return clientset, nil
}

// GetPodsData lists pods in the target namespace(s) of a target cluster and creates a list of
// PodData objects for them
func GetPodsData(namespaces []string, excludeNamespaces []string, clientset *kubernetes.Clientset) ([]*PodData, error) {
	podsData := []*PodData{}
	ctx := context.Background()
	list := &corev1.PodList{}
	filteredNamespaces := []string{}

	var (
		wg    sync.WaitGroup
		mutex = &sync.Mutex{}
	)

	// get all namespaces in the cluster
	nsList, err := GetClusterNamespaces(clientset)
	if err != nil {
		return podsData, err
	}

	if len(excludeNamespaces) > 0 {
		filteredNamespaces, err = filterNamespaces(nsList.Items, excludeNamespaces, "exclude")
		if err != nil {
			return podsData, fmt.Errorf("could not filter namespaces: %v ", err)
		}
	} else if len(namespaces) == 0 {
		var err error
		list, err = clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return podsData, fmt.Errorf("could not list pods on cluster scope: %v ", err)
		}
	} else {
		filteredNamespaces, err = filterNamespaces(nsList.Items, namespaces, "include")
		if err != nil {
			return podsData, fmt.Errorf("could not filter namespaces: %v ", err)
		}
	}

	log.Printf("scraping the following namespaces: %v \n", filteredNamespaces)

	// run concurrently
	errs := make(chan error, 1) // Buffered only for the first error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure it's called to release resources even if no errors

	for _, ns := range filteredNamespaces {
		wg.Add(1)
		go func(ns string) {
			defer wg.Done()
			// Check if any error occurred in any other gorouties:
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
			}

			pods, err := getPodsInNamespace(ns, clientset)
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

	podsData = processPods(list)

	return podsData, nil
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
func filterNamespaces(nsList []corev1.Namespace, patterns []string, operation string) ([]string, error) {
	result := []string{}
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

			// if exclude nothing, then add all namespaces to result
			if len(patterns) == 0 && operation == "exclude" {
				mutex.Lock()
				result = append(result, ns)
				mutex.Unlock()
			}
			for _, p := range patterns {
				r, err := regexp.Compile(p)
				if err != nil {
					select {
					case errs <- fmt.Errorf("failed to compile regex pattern %v : %v", p, err):
					default:
					}
					cancel() // send cancel signal to goroutines
					return
				}
				match := r.MatchString(ns)
				switch operation {
				case "include":
					if match {
						mutex.Lock()
						result = append(result, ns)
						mutex.Unlock()
					}
				case "exclude":
					if !match {
						mutex.Lock()
						result = append(result, ns)
						mutex.Unlock()
					}

				default:
					select {
					case errs <- fmt.Errorf("unsupported operation %s", operation):
					default:
					}
					cancel() // send cancel signal to goroutines
					return

				}
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
func getPodsInNamespace(namespace string, clientset *kubernetes.Clientset) ([]corev1.Pod, error) {
	ctx := context.Background()
	podlist, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []corev1.Pod{}, fmt.Errorf("could not list pods on namespace %s: %v ", namespace, err)
	}
	return podlist.Items, nil
}

// GetClusterNamespaces gets a namespace list from the cluster
func GetClusterNamespaces(clientset *kubernetes.Clientset) (*corev1.NamespaceList, error) {
	ctx := context.Background()

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return namespaces, fmt.Errorf("could not list namespaces on cluster scope: %v ", err)
	}

	return namespaces, nil
}
