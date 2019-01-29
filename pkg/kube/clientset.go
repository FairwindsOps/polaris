package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func createClientset() *kubernetes.Clientset {
	kubeConf := config.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return clientset
}

var clientset = createClientset()

// CoreV1API exports the v1 Core API client.
var CoreV1API = clientset.CoreV1()

// AppsV1API exports the v1 Apps API client.
var AppsV1API = clientset.AppsV1()
