package kube

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // Required for GKE auth.
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// API is a wrapper for the clientset and methods to interact with the Kubernetes API.
type API struct {
	Clientset kubernetes.Interface
}

// GetDeploys gets all the deployments in the k8s cluster.
func (api *API) GetDeploys() (*appsv1.DeploymentList, error) {
	return api.Clientset.AppsV1().Deployments("").List(metav1.ListOptions{})
}

// CreateKubeAPI returns a new KubeAPI object to interact with the cluster API with.
func CreateKubeAPI() (*API, error) {
	kubeConf := config.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	// return clientset, nil
	api := API{
		Clientset: clientset,
	}
	return &api, nil
}
