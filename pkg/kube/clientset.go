package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes" // Required for GKE auth.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// CreateClientset returns a new Kubernetes clientset.
func CreateClientset() *kubernetes.Clientset {
	kubeConf := config.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return clientset
}
