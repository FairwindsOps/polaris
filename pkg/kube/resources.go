package kube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	Ingresses     []v1beta1.Ingress
}

type k8sResource struct {
	Kind string `yaml:"kind"`
}

var podSpecFields = []string{"jobTemplate", "spec", "template"}

// CreateResourceProvider returns a new ResourceProvider object to interact with k8s resources
func CreateResourceProvider(ctx context.Context, directory, workload string) (*ResourceProvider, error) {
	if workload != "" {
		return CreateResourceProviderFromWorkload(ctx, workload)
	}
	if directory != "" {
		return CreateResourceProviderFromPath(directory)
	}
	return CreateResourceProviderFromCluster(ctx)
}

// CreateResourceProviderFromWorkload creates a new ResourceProvider that just contains one workload
func CreateResourceProviderFromWorkload(ctx context.Context, workload string) (*ResourceProvider, error) {
	kubeConf, configError := config.GetConfig()
	if configError != nil {
		logrus.Errorf("Error fetching KubeConfig: %v", configError)
		return nil, configError
	}
	kube, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		logrus.Errorf("Error creating Kubernetes client: %v", err)
		return nil, err
	}
	serverVersion, err := kube.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching Cluster API version: %v", err)
		return nil, err
	}
	resources := ResourceProvider{
		ServerVersion: serverVersion.Major + "." + serverVersion.Minor,
		SourceType:    "Workload",
		SourceName:    workload,
		CreationTime:  time.Now(),
		Nodes:         []corev1.Node{},
		Namespaces:    []corev1.Namespace{},
	}

	parts := strings.Split(workload, "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("Invalid workload identifier %s. Should be in format namespace/kind/version/name, e.g. nginx-ingress/Deployment.apps/v1/default-backend", workload)
	}
	namespace := parts[0]
	kind := parts[1]
	version := parts[2]
	name := parts[3]

	dynamicInterface, err := dynamic.NewForConfig(kubeConf)
	if err != nil {
		logrus.Errorf("Error connecting to dynamic interface: %v", err)
		return nil, err
	}
	groupResources, err := restmapper.GetAPIGroupResources(kube.Discovery())
	if err != nil {
		logrus.Errorf("Error getting API Group resources: %v", err)
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	obj, err := getObject(ctx, namespace, kind, version, name, &dynamicInterface, &restMapper)
	if err != nil {
		logrus.Errorf("Could not find workload %s: %v", workload, err)
		return nil, err
	}
	workloadObj, err := NewGenericWorkloadFromUnstructured(kind, obj)
	if err != nil {
		logrus.Errorf("Could not parse workload %s: %v", workload, err)
		return nil, err
	}

	resources.Controllers = []GenericWorkload{workloadObj}
	return &resources, nil
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

	if directory == "-" {
		fi, err := os.Stdin.Stat()
		if err == nil && fi.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
			if err := addResourcesFromReader(os.Stdin, &resources); err != nil {
				return nil, err
			}
			return &resources, nil
		}
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
		return addResourcesFromYaml(string(contents), &resources)
	}

	err := filepath.Walk(directory, visitFile)
	if err != nil {
		return nil, err
	}
	return &resources, nil
}

// CreateResourceProviderFromCluster creates a new ResourceProvider using live data from a cluster
func CreateResourceProviderFromCluster(ctx context.Context) (*ResourceProvider, error) {
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
	return CreateResourceProviderFromAPI(ctx, api, kubeConf.Host, &dynamicInterface)
}

// CreateResourceProviderFromAPI creates a new ResourceProvider from an existing k8s interface
func CreateResourceProviderFromAPI(ctx context.Context, kube kubernetes.Interface, clusterName string, dynamic *dynamic.Interface) (*ResourceProvider, error) {
	listOpts := metav1.ListOptions{}
	serverVersion, err := kube.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching Cluster API version: %v", err)
		return nil, err
	}

	nodes, err := kube.CoreV1().Nodes().List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Nodes: %v", err)
		return nil, err
	}
	namespaces, err := kube.CoreV1().Namespaces().List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Namespaces: %v", err)
		return nil, err
	}
	pods, err := kube.CoreV1().Pods("").List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Pods: %v", err)
		return nil, err
	}
	ingressList, err := kube.ExtensionsV1beta1().Ingresses("").List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Ingresses: %v", err)
		return nil, err
	}

	resources, err := restmapper.GetAPIGroupResources(kube.Discovery())
	if err != nil {
		logrus.Errorf("Error getting API Group resources: %v", err)
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(resources)

	objectCache := map[string]unstructured.Unstructured{}

	controllers, err := LoadControllers(ctx, pods.Items, dynamic, &restMapper, objectCache)
	if err != nil {
		logrus.Errorf("Error loading controllers from pods: %v", err)
		return nil, err
	}

	api := ResourceProvider{
		ServerVersion: serverVersion.Major + "." + serverVersion.Minor,
		SourceType:    "Cluster",
		SourceName:    clusterName,
		CreationTime:  time.Now(),
		Nodes:         nodes.Items,
		Namespaces:    namespaces.Items,
		Controllers:   controllers,
		Ingresses:     ingressList.Items,
	}
	return &api, nil
}

// LoadControllers loads a list of controllers from the kubeResources Pods
func LoadControllers(ctx context.Context, pods []corev1.Pod, dynamicClientPointer *dynamic.Interface, restMapperPointer *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) ([]GenericWorkload, error) {
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
		workload, err := NewGenericWorkload(ctx, pod, dynamicClientPointer, restMapperPointer, objectCache)
		if err != nil {
			return nil, err
		}
		interfaces = append(interfaces, workload)
	}
	return deduplicateControllers(interfaces), nil
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

func addResourcesFromReader(reader io.Reader, resources *ResourceProvider) error {
	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		logrus.Errorf("Error reading from %v: %v", reader, err)
		return err
	}
	if err := addResourcesFromYaml(string(contents), resources); err != nil {
		return err
	}
	return nil
}

func addResourcesFromYaml(contents string, resources *ResourceProvider) error {
	specs := regexp.MustCompile("[\r\n]-+[\r\n]").Split(string(contents), -1)
	for _, spec := range specs {
		if strings.TrimSpace(spec) == "" {
			continue
		}
		err := addResourceFromString(spec, resources)
		if err != nil {
			logrus.Errorf("Error parsing YAML: (%v)", err)
			return err
		}
	}
	return nil
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
		if err != nil {
			return err
		}
		workload, err := NewGenericWorkloadFromPod(pod, pod)
		if err != nil {
			return err
		}
		resources.Controllers = append(resources.Controllers, workload)
	} else if resource.Kind == "Ingress" {
		ingress := v1beta1.Ingress{}
		err = decoder.Decode(&ingress)
		resources.Ingresses = append(resources.Ingresses, ingress)
	} else {
		newController, err := GetWorkloadFromBytes(contentBytes)
		if err != nil || newController == nil {
			return err
		}
		resources.Controllers = append(resources.Controllers, *newController)
	}
	return err
}
