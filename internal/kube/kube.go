package kube

import (
	"context"
	"fmt"

	"github.com/merkely-development/watcher/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// PodData represents the harvested pod data
type PodData struct {
	Name              string                  `json:"name"`
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
		Name:              pod.Name,
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

// GetPodsData lists pods in the target namespace of a target cluster and creates a list of
// PodData objects for them
func GetPodsData(namespaces []string, excludeNamespace []string, clientset *kubernetes.Clientset) ([]*PodData, error) {
	podsData := []*PodData{}
	ctx := context.Background()
	list := &corev1.PodList{}

	if len(excludeNamespace) > 0 {
		nsList, err := GetClusterNamespaces(clientset)
		if err != nil {
			return podsData, err
		}

		for _, ns := range nsList.Items {
			if !utils.Contains(excludeNamespace, ns.Name) {
				pods, err := getPodsInNamespace(ns.Name, clientset)
				if err != nil {
					return podsData, err
				}
				list.Items = append(list.Items, pods...)
			}
		}

	} else if len(namespaces) == 0 {
		var err error
		list, err = clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return podsData, fmt.Errorf("could not list pods on cluster scope: %v ", err)
		}
	}

	for _, ns := range namespaces {

		pods, err := getPodsInNamespace(ns, clientset)
		if err != nil {
			return podsData, err
		}
		list.Items = append(list.Items, pods...)
	}

	for _, pod := range list.Items {
		podsData = append(podsData, NewPodData(&pod))
	}

	return podsData, nil
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
