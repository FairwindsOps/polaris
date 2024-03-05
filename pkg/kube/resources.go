// Copyright 2022 FairwindsOps, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fairwindsops/controller-utils/pkg/controller"
	conf "github.com/fairwindsops/polaris/pkg/config"

	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for other auth providers like GKE.
	"k8s.io/client-go/rest"
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
	Resources     resourceKindMap
}

type resourceKindMap map[string][]GenericResource

func (rkm resourceKindMap) addResource(r GenericResource) {
	gvk := r.Resource.GroupVersionKind()
	var key string
	if gvk.Group != "" {
		key = gvk.Group + "/" + gvk.Kind
	} else {
		key = gvk.Kind
	}
	rkm[key] = append(rkm[key], r)
}

func (rkm resourceKindMap) addResources(rs []GenericResource) {
	for _, r := range rs {
		rkm.addResource(r)
	}
}

func (rkm resourceKindMap) GetLength() int {
	total := 0
	for _, rs := range rkm {
		total += len(rs)
	}
	return total
}

func (rkm resourceKindMap) GetNumberOfControllers() int {
	total := 0
	for _, rs := range rkm {
		for _, r := range rs {
			if r.PodSpec != nil {
				total++
			}
		}
	}
	return total
}

// This is here for backward compatibility reasons
func maybeTransformKindIntoGroupKind(k string) string {
	if k == "Ingress" {
		return "networking.k8s.io/Ingress"
	} else if k == "PodDisruptionBudget" {
		return "policy/PodDisruptionBudget"
	}
	return k
}

func parseGroupKind(gk string) schema.GroupKind {
	i := strings.Index(gk, "/")
	if i == -1 {
		return schema.GroupKind{Kind: gk}
	}

	group := gk[:i]
	kind := gk[i+1:]
	return schema.GroupKind{Group: group, Kind: kind}
}

func newResourceProvider(version, sourceType, sourceName string) ResourceProvider {
	return ResourceProvider{
		ServerVersion: version,
		SourceType:    sourceType,
		SourceName:    sourceName,
		CreationTime:  time.Now(),
		Nodes:         make([]corev1.Node, 0),
		Namespaces:    make([]corev1.Namespace, 0),
		Resources:     make(map[string][]GenericResource),
	}
}

type k8sResource struct {
	Kind string `yaml:"kind"`
}

var podSpecFields = []string{"jobTemplate", "spec", "template"}

// CreateResourceProvider returns a new ResourceProvider object to interact with k8s resources
func CreateResourceProvider(ctx context.Context, directory, workload string, c conf.Configuration) (*ResourceProvider, error) {
	if workload != "" {
		return CreateResourceProviderFromResource(ctx, workload)
	}
	if directory != "" {
		return CreateResourceProviderFromPath(directory)
	}
	return CreateResourceProviderFromCluster(ctx, c)
}

// CreateResourceProviderFromResource creates a new ResourceProvider that just contains one workload
func CreateResourceProviderFromResource(ctx context.Context, workload string) (*ResourceProvider, error) {
	dynamicClient, restMapper, clientSet, _, err := GetKubeClient(ctx, "")
	if err != nil {
		return nil, err
	}
	serverVersion, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("Error fetching Cluster API version: %w", err)
	}
	resources := newResourceProvider(serverVersion.Major+"."+serverVersion.Minor, "Resource", workload)

	parts := strings.Split(workload, "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("Invalid workload identifier %s. Should be in format namespace/kind/version/name, e.g. nginx-ingress/Deployment.apps/v1/default-backend", workload)
	}
	namespace := parts[0]
	kind := parts[1]
	version := parts[2]
	name := parts[3]

	obj, err := getObject(ctx, namespace, kind, version, name, dynamicClient, restMapper)
	if err != nil {
		return nil, fmt.Errorf("Could not find workload %s: %w", workload, err)
	}
	workloadObj, err := NewGenericResourceFromUnstructured(*obj, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not parse workload %s: %w", workload, err)
	}
	resources.Resources.addResource(workloadObj)
	return &resources, nil
}

// CreateResourceProviderFromPath returns a new ResourceProvider using the YAML files in a directory
func CreateResourceProviderFromPath(directory string) (*ResourceProvider, error) {
	resources := newResourceProvider("unknown", "Path", directory)

	if directory == "-" {
		fi, err := os.Stdin.Stat()
		if err == nil && fi.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
			if err := resources.addResourcesFromReader(os.Stdin); err != nil {
				return nil, err
			}
			return &resources, nil
		}
	}

	visitFile := func(path string, f os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			logrus.Errorf("Error reading file: %v", path)
			return err
		}
		err = resources.addResourcesFromYaml(string(contents))
		if err != nil {
			logrus.Warnf("Skipping %s: cannot add resource from YAML: %v", path, err)
		}
		return nil
	}

	err := filepath.Walk(directory, visitFile)
	if err != nil {
		return nil, err
	}
	return &resources, nil
}

// CreateResourceProviderFromYaml returns a new ResourceProvider using the yaml
func CreateResourceProviderFromYaml(yamlContent string) *ResourceProvider {
	resources := newResourceProvider("unknown", "Content", "unknown")
	resources.addResourcesFromYaml(string(yamlContent))
	return &resources
}

// CreateResourceProviderFromCluster creates a new ResourceProvider using live data from a cluster
func CreateResourceProviderFromCluster(ctx context.Context, c conf.Configuration) (*ResourceProvider, error) {
	dynamicClient, _, clientSet, clusterHost, err := GetKubeClient(ctx, c.KubeContext)
	if err != nil {
		return nil, err
	}
	return CreateResourceProviderFromAPI(ctx, clientSet, clusterHost, dynamicClient, c)
}

func GetKubeClient(ctx context.Context, kubeContext string) (dynamic.Interface, meta.RESTMapper, kubernetes.Interface, string, error) {
	var kubeConf *rest.Config
	var err error
	if len(kubeContext) > 0 {
		kubeConf, err = config.GetConfigWithContext(kubeContext)
	} else {
		kubeConf, err = config.GetConfig()
	}
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("Error fetching KubeConfig: %v", err)
	}
	clientSet, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("Error creating Kubernetes client: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(kubeConf)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("Error connecting to dynamic interface: %v", err)
	}
	resources, err := restmapper.GetAPIGroupResources(clientSet.Discovery())
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("Error getting API Group resources: %v", err)
	}
	return dynamicClient, restmapper.NewDiscoveryRESTMapper(resources), clientSet, kubeConf.Host, nil
}

// CreateResourceProviderFromAPI creates a new ResourceProvider from an existing k8s interface
func CreateResourceProviderFromAPI(ctx context.Context, kube kubernetes.Interface, clusterName string, dynamic dynamic.Interface, c conf.Configuration) (*ResourceProvider, error) {
	listOpts := metav1.ListOptions{}
	serverVersion, err := kube.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching Cluster API version: %v", err)
		return nil, err
	}

	sourceType := "Cluster"
	if c.Namespace != "" {
		logrus.Debug("namespace is specififed in config, setting source type to ClusterNamespace")
		sourceType = "ClusterNamespace"
	}
	provider := newResourceProvider(serverVersion.Major+"."+serverVersion.Minor, sourceType, clusterName)

	logrus.Info("Loading nodes")
	nodes, err := kube.CoreV1().Nodes().List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Nodes: %v", err)
		return nil, err
	}

	logrus.Info("Loading namespaces")
	var namespaces *corev1.NamespaceList
	if c.Namespace != "" {
		ns, err := kube.CoreV1().Namespaces().Get(ctx, c.Namespace, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		namespaces = &corev1.NamespaceList{
			Items: []corev1.Namespace{*ns},
		}
	} else {
		nsList, err := kube.CoreV1().Namespaces().List(ctx, listOpts)
		if err != nil {
			logrus.Errorf("Error fetching Namespaces: %v", err)
			return nil, err
		}
		namespaces = nsList
	}
	logrus.Info("Setting up restmapper")
	resources, err := restmapper.GetAPIGroupResources(kube.Discovery())
	if err != nil {
		logrus.Errorf("Error getting API Group resources: %v", err)
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(resources)
	allChecks := []conf.SchemaCheck{}
	for _, check := range c.CustomChecks {
		allChecks = append(allChecks, check)
	}
	for _, check := range conf.BuiltInChecks {
		allChecks = append(allChecks, check)
	}

	var additionalKinds []conf.TargetKind
	for _, check := range allChecks {
		neededKinds := []conf.TargetKind{check.Target}
		for key := range check.AdditionalSchemas {
			neededKinds = append(neededKinds, conf.TargetKind(key))
		}
		for key := range check.AdditionalSchemaStrings {
			neededKinds = append(neededKinds, conf.TargetKind(key))
		}
		for _, kind := range neededKinds {
			if !funk.Contains(conf.HandledTargets, kind) && !funk.Contains(additionalKinds, kind) {
				additionalKinds = append(additionalKinds, kind)
			}
		}
	}

	var kubernetesResources []GenericResource
	for _, kind := range additionalKinds {
		groupKind := parseGroupKind(maybeTransformKindIntoGroupKind(string(kind)))
		mapping, err := restMapper.RESTMapping(groupKind)
		if err != nil {
			logrus.Warnf("Error retrieving mapping of Kind %s because of error: %v", kind, err)
			return nil, err
		}
		if c.Namespace != "" && mapping.Scope.Name() != meta.RESTScopeNameNamespace {
			logrus.Infof("Skipping %s because of auditing specific namespace", mapping.GroupVersionKind)
			continue
		}

		logrus.Info("Loading " + kind)
		objects, err := dynamic.Resource(mapping.Resource).Namespace(c.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logrus.Warnf("Error retrieving parent object API %s and Kind %s because of error: %v", mapping.Resource.Version, mapping.Resource.Resource, err)
			return nil, err
		}
		for _, obj := range objects.Items {
			res, err := NewGenericResourceFromUnstructured(obj, nil)
			if err != nil {
				return nil, err
			}
			kubernetesResources = append(kubernetesResources, res)
		}
	}
	logrus.Info("Loading controllers")
	client := controller.Client{
		Context:    ctx,
		Dynamic:    dynamic,
		RESTMapper: restMapper,
	}
	topControllers, err := client.GetAllTopControllersSummary("")
	if err != nil {
		return nil, fmt.Errorf("error while getting all TopControllers: %v", err)
	}
	for _, workload := range topControllers {
		topController := workload.TopController
		workloadObj, err := NewGenericResourceFromUnstructured(topController, nil)
		if err != nil {
			return nil, fmt.Errorf("could not parse workload %v: %w", workload, err)
		}
		kubernetesResources = append(kubernetesResources, workloadObj)
	}

	provider.Nodes = nodes.Items
	provider.Namespaces = namespaces.Items
	provider.Resources.addResources(kubernetesResources)
	logrus.Info("Done loading Kubernetes resources")
	return &provider, nil
}

func (resources *ResourceProvider) addResourcesFromReader(reader io.Reader) error {
	contents, err := io.ReadAll(reader)
	if err != nil {
		logrus.Errorf("Error reading from %v: %v", reader, err)
		return err
	}
	if err := resources.addResourcesFromYaml(string(contents)); err != nil {
		return err
	}
	return nil
}

func (resources *ResourceProvider) addResourcesFromYaml(contents string) error {
	specs := regexp.MustCompile("[\r\n]-+[\r\n]").Split(string(contents), -1)
	for _, spec := range specs {
		if strings.TrimSpace(spec) == "" {
			continue
		}
		err := resources.addResourceFromString(spec)
		if err != nil {
			logrus.Errorf("Error parsing YAML: (%v)", err)
			return err
		}
	}
	return nil
}

func (resources *ResourceProvider) addResourceFromString(contents string) error {
	contentBytes := []byte(contents)
	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)
	resource := k8sResource{}
	err := decoder.Decode(&resource)
	decoder = k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)

	if err != nil {
		return err
	}
	if resource.Kind == "Namespace" {
		ns := corev1.Namespace{}
		err = decoder.Decode(&ns)
		resources.Namespaces = append(resources.Namespaces, ns)
	}

	if resource.Kind == "Pod" {
		pod := corev1.Pod{}
		err = decoder.Decode(&pod)
		if err != nil {
			return err
		}
		workload, err := NewGenericResourceFromPod(pod, pod)
		if err != nil {
			return err
		}
		workload.OriginalObjectYAML = contentBytes
		resources.Resources.addResource(workload)
	} else {
		newResource, err := NewGenericResourceFromBytes(contentBytes)
		if err != nil {
			return err
		}
		resources.Resources.addResource(newResource)
	}
	return err
}

// SerializePodSpec converts a typed PodSpec into a map[string]interface{}
func SerializePodSpec(pod *corev1.PodSpec) (map[string]interface{}, error) {
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	podMap := make(map[string]interface{})
	err = json.Unmarshal(podJSON, &podMap)
	if err != nil {
		return nil, err
	}
	return podMap, nil
}

// SerializePod converts a typed Pod into a map[string]interface{}
func SerializePod(pod *corev1.Pod) (map[string]interface{}, error) {
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	podMap := make(map[string]interface{})
	err = json.Unmarshal(podJSON, &podMap)
	if err != nil {
		return nil, err
	}
	return podMap, nil
}

// SerializeContainer converts a typed Container into a map[string]interface{}
func SerializeContainer(container *corev1.Container) (map[string]interface{}, error) {
	containerJSON, err := json.Marshal(container)
	if err != nil {
		return nil, err
	}
	containerMap := make(map[string]interface{})
	err = json.Unmarshal(containerJSON, &containerMap)
	if err != nil {
		return nil, err
	}
	return containerMap, nil
}
