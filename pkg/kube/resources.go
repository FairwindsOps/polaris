package kube

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // Required for GKE auth.
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ResourceProvider contains k8s resources to be audited
type ResourceProvider struct {
	ServerVersion string
	Nodes         []corev1.Node
	Deployments   []appsv1.Deployment
	Namespaces    []corev1.Namespace
	Pods          []corev1.Pod
}

type k8sResource struct {
	Kind string `yaml:"kind"`
}

// CreateResourceProvider returns a new ResourceProvider object to interact with k8s resources
func CreateResourceProvider(directory string) (*ResourceProvider, error) {
	if directory != "" {
		return CreateResourceProviderFromPath(directory)
	}
	return CreateResourceProviderFromCluster()
}

// CreateResourceProviderFromPath returns a new ResourceProvider using the YAML files in a directory
func CreateResourceProviderFromPath(directory string) (*ResourceProvider, error) {
	resources := ResourceProvider{
		ServerVersion: "unknown",
		Nodes:         []corev1.Node{},
		Deployments:   []appsv1.Deployment{},
		Namespaces:    []corev1.Namespace{},
		Pods:          []corev1.Pod{},
	}

	addYaml := func(contents string) error {
		contentBytes := []byte(contents)
		decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)
		resource := k8sResource{}
		err := decoder.Decode(&resource)
		if err != nil {
			// TODO: should we panic if the YAML is bad?
			logrus.Errorf("Invalid YAML: %s", string(contents))
			return nil
		}
		decoder = k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)
		if resource.Kind == "Deployment" {
			dep := appsv1.Deployment{}
			err = decoder.Decode(&dep)
			if err != nil {
				logrus.Errorf("Error parsing deployment %v", err)
				return err
			}
			resources.Deployments = append(resources.Deployments, dep)
		} else if resource.Kind == "Namespace" {
			ns := corev1.Namespace{}
			err = decoder.Decode(&ns)
			if err != nil {
				logrus.Errorf("Error parsing namespace %v", err)
				return err
			}
			resources.Namespaces = append(resources.Namespaces, ns)
		} else if resource.Kind == "Pod" {
			pod := corev1.Pod{}
			err = decoder.Decode(&pod)
			if err != nil {
				logrus.Errorf("Error parsing pod %v", err)
				return err
			}
			resources.Pods = append(resources.Pods, pod)
		}
		return nil
	}

	visitFile := func(path string, f os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			logrus.Errorf("Error reading file %v", path)
			return err
		}
		specs := regexp.MustCompile("\n-+\n").Split(string(contents), -1)
		for _, spec := range specs {
			err = addYaml(spec)
			if err != nil {
				logrus.Errorf("Error parsing YAML %v", err)
				return err
			}
		}
		return nil
	}

	err := filepath.Walk(directory, visitFile)
	if err != nil {
		return nil, err
	}
	return &resources, nil
}

// CreateResourceProviderFromCluster creates a new ResourceProvider using live data from a cluster
func CreateResourceProviderFromCluster() (*ResourceProvider, error) {
	kubeConf := config.GetConfigOrDie()
	api, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		logrus.Errorf("Error creating Kubernetes client %v", err)
		return nil, err
	}
	return CreateResourceProviderFromAPI(api)
}

// CreateResourceProviderFromAPI creates a new ResourceProvider from an existing k8s interface
func CreateResourceProviderFromAPI(kube kubernetes.Interface) (*ResourceProvider, error) {
	listOpts := metav1.ListOptions{}
	serverVersion, err := kube.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching Kubernetes API version %v", err)
		return nil, err
	}
	deploys, err := kube.AppsV1().Deployments("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Kubernetes Deployments %v", err)
		return nil, err
	}
	nodes, err := kube.CoreV1().Nodes().List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Kubernetes Nodes %v", err)
		return nil, err
	}
	namespaces, err := kube.CoreV1().Namespaces().List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Kubernetes Namespaces %v", err)
		return nil, err
	}
	pods, err := kube.CoreV1().Pods("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Kubernetes Pods %v", err)
		return nil, err
	}
	api := ResourceProvider{
		ServerVersion: serverVersion.Major + "." + serverVersion.Minor,
		Deployments:   deploys.Items,
		Nodes:         nodes.Items,
		Namespaces:    namespaces.Items,
		Pods:          pods.Items,
	}
	return &api, nil
}
