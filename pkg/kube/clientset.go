package kube

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // Required for GKE auth.
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ResourceProvider contains k8s resources to be audited
type ResourceProvider struct {
	ServerVersion string
	Deployments   []appsv1.Deployment
	Nodes         []corev1.Node
	Namespaces    []corev1.Namespace
	Pods          map[string][]corev1.Pod
}

// CreateKubeResourceProvider returns a new ResourceProvider object to interact with k8s resources
func CreateKubeResourceProvider() (*ResourceProvider, error) {
	kubeConf := config.GetConfigOrDie()
	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	listOpts := metav1.ListOptions{}
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	deploys, err := clientset.AppsV1().Deployments("").List(listOpts)
	if err != nil {
		return nil, err
	}
	nodes, err := clientset.CoreV1().Nodes().List(listOpts)
	if err != nil {
		return nil, err
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(listOpts)
	if err != nil {
		return nil, err
	}
	podsByNamespace := map[string][]corev1.Pod{}
	for _, ns := range namespaces.Items {
		pods, err := clientset.CoreV1().Pods(ns.Name).List(listOpts)
		if err != nil {
			return nil, err
		}
		podsByNamespace[ns.Name] = pods.Items
	}
	api := ResourceProvider{
		ServerVersion: serverVersion.Major + "." + serverVersion.Minor,
		Deployments:   deploys.Items,
		Nodes:         nodes.Items,
		Namespaces:    namespaces.Items,
		Pods:          podsByNamespace,
	}
	return &api, nil
}
