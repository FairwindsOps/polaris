package kube

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for other auth providers like GKE.
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ResourceProvider contains k8s resources to be audited
type ResourceProvider struct {
	ServerVersion string
	CreationTime  time.Time
	SourceName    string
	SourceType    string
	Nodes         []corev1.Node
	Deployments   []appsv1.Deployment
	StatefulSets  []appsv1.StatefulSet
	Namespaces    []corev1.Namespace
	Pods          []corev1.Pod
	LimitRanges   []corev1.LimitRange
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
		SourceType:    "Path",
		SourceName:    directory,
		Nodes:         []corev1.Node{},
		Deployments:   []appsv1.Deployment{},
		StatefulSets:  []appsv1.StatefulSet{},
		Namespaces:    []corev1.Namespace{},
		Pods:          []corev1.Pod{},
	}

	addYaml := func(contents string) error {
		return addResourceFromString(contents, &resources)
	}

	visitFile := func(path string, f os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			logrus.Errorf("Error reading file: %v", path)
			return err
		}
		specs := regexp.MustCompile("\n-+\n").Split(string(contents), -1)
		for _, spec := range specs {
			if strings.TrimSpace(spec) == "" {
				continue
			}
			err = addYaml(spec)
			if err != nil {
				logrus.Errorf("Error parsing YAML: %v", err)
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
	kubeConf, configError := config.GetConfig()
	if configError != nil {
		logrus.Errorf("Error fetching KubeConfig: %v", configError)
		return nil, configError
	}
	api, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		logrus.Errorf("Error creating Kubernetes client: %v", err)
		return nil, err
	}
	return CreateResourceProviderFromAPI(api, kubeConf.Host)
}

// CreateResourceProviderFromAPI creates a new ResourceProvider from an existing k8s interface
func CreateResourceProviderFromAPI(kube kubernetes.Interface, clusterName string) (*ResourceProvider, error) {
	listOpts := metav1.ListOptions{}
	serverVersion, err := kube.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching Cluster API version: %v", err)
		return nil, err
	}
	deploys, err := kube.AppsV1().Deployments("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Deployments: %v", err)
		return nil, err
	}
	statefulSets, err := kube.AppsV1().StatefulSets("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching StatefulSets%v", err)
		return nil, err
	}
	nodes, err := kube.CoreV1().Nodes().List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Nodes: %v", err)
		return nil, err
	}
	namespaces, err := kube.CoreV1().Namespaces().List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Namespaces: %v", err)
		return nil, err
	}
	pods, err := kube.CoreV1().Pods("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Pods: %v", err)
		return nil, err
	}
	limitRanges, err := kube.CoreV1().LimitRanges("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching LimitRanges: %v", err)
		return nil, err
	}

	api := ResourceProvider{
		ServerVersion: serverVersion.Major + "." + serverVersion.Minor,
		SourceType:    "Cluster",
		SourceName:    clusterName,
		CreationTime:  time.Now(),
		Deployments:   deploys.Items,
		StatefulSets:  statefulSets.Items,
		Nodes:         nodes.Items,
		Namespaces:    namespaces.Items,
		Pods:          pods.Items,
		LimitRanges:   limitRanges.Items,
	}
	return &api, nil
}

func addResourceFromString(contents string, resources *ResourceProvider) error {
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
		resources.Deployments = append(resources.Deployments, dep)
	} else if resource.Kind == "StatefulSet" {
		dep := appsv1.StatefulSet{}
		err = decoder.Decode(&dep)
		resources.StatefulSets = append(resources.StatefulSets, dep)
	} else if resource.Kind == "Namespace" {
		ns := corev1.Namespace{}
		err = decoder.Decode(&ns)
		resources.Namespaces = append(resources.Namespaces, ns)
	} else if resource.Kind == "Pod" {
		pod := corev1.Pod{}
		err = decoder.Decode(&pod)
		resources.Pods = append(resources.Pods, pod)
	} else if resource.Kind == "LimitRange" {
		lr := corev1.LimitRange{}
		err = decoder.Decode(&lr)
		resources.LimitRanges = append(resources.LimitRanges, lr)
	}
	if err != nil {
		logrus.Errorf("Error parsing %s: %v", resource.Kind, err)
		return err
	}
	return nil
}
