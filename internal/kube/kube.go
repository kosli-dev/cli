package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// PodData represents the harvested pod data
type PodData struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Images    map[string]string `json:"images"`
}

// NewPodData creates a PodData object from a k8s pod
func NewPodData(pod *corev1.Pod) *PodData {
	images := make(map[string]string)

	containers := pod.Status.ContainerStatuses
	for _, cs := range containers {
		images[cs.Image] = cs.ImageID[len(cs.ImageID)-64:]
	}

	return &PodData{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Images:    images,
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
func GetPodsData(namespace string, clientset *kubernetes.Clientset) ([]*PodData, error) {
	podsData := []*PodData{}
	ctx := context.Background()
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return podsData, fmt.Errorf("could not list pods: %v ", err)
	}

	for _, pod := range list.Items {
		podsData = append(podsData, NewPodData(&pod))
	}

	return podsData, nil
}
