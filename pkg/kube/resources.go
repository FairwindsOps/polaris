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
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for other auth providers like GKE.
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ResourceProvider contains k8s resources to be audited
type ResourceProvider struct {
	ServerVersion string
	CreationTime  time.Time
	SourceName    string
	SourceType    string
	Nodes         []corev1.Node
	Namespaces    []corev1.Namespace
	Pods          []corev1.Pod
	DynamicClient *dynamic.Interface
	RestMapper    *meta.RESTMapper
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
				logrus.Errorf("Error parsing YAML: (%v)", err)
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
	dynamicInterface, err := dynamic.NewForConfig(kubeConf)
	if err != nil {
		logrus.Errorf("Error connecting to dynamic interface: %v", err)
		return nil, err
	}
	return CreateResourceProviderFromAPI(api, kubeConf.Host, &dynamicInterface)
}

// CreateResourceProviderFromAPI creates a new ResourceProvider from an existing k8s interface
func CreateResourceProviderFromAPI(kube kubernetes.Interface, clusterName string, dynamic *dynamic.Interface) (*ResourceProvider, error) {
	listOpts := metav1.ListOptions{}
	serverVersion, err := kube.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching Cluster API version: %v", err)
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

	resources, err := restmapper.GetAPIGroupResources(kube.Discovery())
	if err != nil {
		logrus.Errorf("Error getting API Group resources: %v", err)
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(resources)

	api := ResourceProvider{
		ServerVersion: serverVersion.Major + "." + serverVersion.Minor,
		SourceType:    "Cluster",
		SourceName:    clusterName,
		CreationTime:  time.Now(),
		Nodes:         nodes.Items,
		Namespaces:    namespaces.Items,
		Pods:          pods.Items,
		DynamicClient: dynamic,
		RestMapper:    &restMapper,
	}
	return &api, nil
}

func getPodSpec(yaml map[string]interface{}) interface{} {
	allowedChildren := []string{"jobTemplate", "spec", "template"}
	for _, child := range allowedChildren {
		if childYaml, ok := yaml[child]; ok {
			return getPodSpec(childYaml.(map[string]interface{}))
		}
	}
	return yaml
}

func addResourceFromString(contents string, resources *ResourceProvider) error {
	contentBytes := []byte(contents)
	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)
	resource := k8sResource{}
	err := decoder.Decode(&resource)
	decoder = k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)

	if err != nil {
		logrus.Errorf("Invalid YAML: %s", string(contents))
		return err
	}
	if resource.Kind == "Namespace" {
		ns := corev1.Namespace{}
		err = decoder.Decode(&ns)
		resources.Namespaces = append(resources.Namespaces, ns)
	} else if resource.Kind == "Pod" {
		pod := corev1.Pod{}
		err = decoder.Decode(&pod)
		resources.Pods = append(resources.Pods, pod)
	} else {
		yamlNode := make(map[string]interface{})
		err = yaml.Unmarshal(contentBytes, &yamlNode)
		if err != nil {
			logrus.Errorf("Invalid YAML: %s", string(contents))
			return err
		}
		finalDoc := make(map[string]interface{})
		finalDoc["metadata"] = yamlNode["metadata"]
		finalDoc["apiVersion"] = "v1"
		finalDoc["kind"] = "Pod"
		finalDoc["spec"] = getPodSpec(yamlNode)
		marshelledYaml, err := yaml.Marshal(finalDoc)
		if err != nil {
			logrus.Errorf("Could not marshell yaml: %v", err)
			return err
		}
		decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(marshelledYaml), 1000)
		pod := corev1.Pod{}
		err = decoder.Decode(&pod)
		resources.Pods = append(resources.Pods, pod)
	}
	return err
}
