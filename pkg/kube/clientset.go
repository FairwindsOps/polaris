package kube

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// createClientset loads kubeconfig and setups the connection to the k8s api.
func createClientset() *kubernetes.Clientset {
	kubeconfig := getKubeConfig()
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("Error:", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return clientset
}

// getKubeConfig returns a valid kubeconfig path.
func getKubeConfig() string {
	var path string

	if os.Getenv("KUBECONFIG") != "" {
		path = os.Getenv("KUBECONFIG")
		return path
	}

	if home, err := homedir.Dir(); err == nil {
		path = filepath.Join(home, ".kube", "config")
	}

	if _, err := os.Stat(path); err != nil {
		// No kubeconfig exists, therefor return an emtpy string to
		// indicate that this web server is running inside the cluster.
		return ""
	}
	return path
}

var clientset = createClientset()

// Kubernetes version 1.11 APIs

// CoreV1API exports the v1 Core API client.
var CoreV1API = clientset.CoreV1()

// AutoscalingV1API exports the v1 Autoscaling API client.
var AutoscalingV1API = clientset.AutoscalingV1()

// AppsV1API exports the v1 Apps API client.
var AppsV1API = clientset.AppsV1()
