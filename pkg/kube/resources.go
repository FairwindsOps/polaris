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
	Controllers   []GenericWorkload
}

type k8sResource struct {
	Kind string `yaml:"kind"`
}

var podSpecFields = []string{"jobTemplate", "spec", "template"}

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
		Controllers:   []GenericWorkload{},
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
		Controllers:   LoadControllers(pods.Items, dynamic, &restMapper),
	}
	return &api, nil
}

// LoadControllers loads a list of controllers from the kubeResources Pods
func LoadControllers(pods []corev1.Pod, dynamicClientPointer *dynamic.Interface, restMapperPointer *meta.RESTMapper) []GenericWorkload {
	interfaces := []GenericWorkload{}
	deduped := map[string]corev1.Pod{}
	for _, pod := range pods {
		owners := pod.ObjectMeta.OwnerReferences
		if len(owners) == 0 {
			deduped[pod.ObjectMeta.Namespace+"/Pod/"+pod.ObjectMeta.Name] = pod
			continue
		}
		deduped[pod.ObjectMeta.Namespace+"/"+owners[0].Kind+"/"+owners[0].Name] = pod
	}
	for _, pod := range deduped {
		interfaces = append(interfaces, NewGenericWorkload(pod, dynamicClientPointer, restMapperPointer))
	}
	return deduplicateControllers(interfaces)
}

// Because the controllers with an Owner take on the name of the Owner, this eliminates any duplicates.
// In cases like CronJobs older children can hang around, so this takes the most recent.
func deduplicateControllers(inputControllers []GenericWorkload) []GenericWorkload {
	controllerMap := make(map[string]GenericWorkload)
	for _, controller := range inputControllers {
		key := controller.ObjectMeta.GetNamespace() + "/" + controller.Kind + "/" + controller.ObjectMeta.GetName()
		oldController, ok := controllerMap[key]
		if !ok || controller.ObjectMeta.GetCreationTimestamp().Time.After(oldController.ObjectMeta.GetCreationTimestamp().Time) {
			controllerMap[key] = controller
		}
	}
	results := make([]GenericWorkload, 0)
	for _, controller := range controllerMap {
		results = append(results, controller)
	}
	return results
}

// GetPodSpec looks inside arbitrary YAML for a PodSpec
func GetPodSpec(yaml map[string]interface{}) interface{} {
	for _, child := range podSpecFields {
		if childYaml, ok := yaml[child]; ok {
			return GetPodSpec(childYaml.(map[string]interface{}))
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
		resources.Controllers = append(resources.Controllers, NewGenericWorkload(pod, nil, nil))
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
		finalDoc["spec"] = GetPodSpec(yamlNode)
		marshaledYaml, err := yaml.Marshal(finalDoc)
		if err != nil {
			logrus.Errorf("Could not marshal yaml: %v", err)
			return err
		}
		decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(marshaledYaml), 1000)
		pod := corev1.Pod{}
		err = decoder.Decode(&pod)
		newController := NewGenericWorkload(pod, nil, nil)
		newController.Kind = resource.Kind
		resources.Controllers = append(resources.Controllers, newController)
	}
	return err
}
