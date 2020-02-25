package kube

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for other auth providers like GKE.
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ResourceProvider contains k8s resources to be audited
type ResourceProvider struct {
	ServerVersion          string
	CreationTime           time.Time
	SourceName             string
	SourceType             string
	Nodes                  []corev1.Node
	Deployments            []appsv1.Deployment
	StatefulSets           []appsv1.StatefulSet
	DaemonSets             []appsv1.DaemonSet
	Jobs                   []batchv1.Job
	CronJobs               []batchv1beta1.CronJob
	ReplicationControllers []corev1.ReplicationController
	Namespaces             []corev1.Namespace
	Pods                   []corev1.Pod
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
		ServerVersion:          "unknown",
		SourceType:             "Path",
		SourceName:             directory,
		Nodes:                  []corev1.Node{},
		Deployments:            []appsv1.Deployment{},
		StatefulSets:           []appsv1.StatefulSet{},
		DaemonSets:             []appsv1.DaemonSet{},
		Jobs:                   []batchv1.Job{},
		CronJobs:               []batchv1beta1.CronJob{},
		ReplicationControllers: []corev1.ReplicationController{},
		Namespaces:             []corev1.Namespace{},
		Pods:                   []corev1.Pod{},
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
	deploys, err := getDeployments(kube)
	if err != nil {
		return nil, err
	}
	statefulSets, err := getStatefulSets(kube)
	if err != nil {
		return nil, err
	}
	daemonSets, err := kube.AppsV1().DaemonSets("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching DaemonSets: %v", err)
		return nil, err
	}
	jobs, err := kube.BatchV1().Jobs("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Jobs: %v", err)
		return nil, err
	}
	cronJobs, err := kube.BatchV1beta1().CronJobs("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching CronJobs: %v", err)
		return nil, err
	}
	replicationControllers, err := kube.CoreV1().ReplicationControllers("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching ReplicationControllers: %v", err)
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

	api := ResourceProvider{
		ServerVersion:          serverVersion.Major + "." + serverVersion.Minor,
		SourceType:             "Cluster",
		SourceName:             clusterName,
		CreationTime:           time.Now(),
		Deployments:            deploys,
		StatefulSets:           statefulSets,
		DaemonSets:             daemonSets.Items,
		Jobs:                   jobs.Items,
		CronJobs:               cronJobs.Items,
		ReplicationControllers: replicationControllers.Items,
		Nodes:                  nodes.Items,
		Namespaces:             namespaces.Items,
		Pods:                   pods.Items,
	}
	return &api, nil
}

func addResourceFromString(contents string, resources *ResourceProvider) error {
	contentBytes := []byte(contents)
	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)
	resource := k8sResource{}
	err := decoder.Decode(&resource)
	if err != nil {
		logrus.Errorf("Invalid YAML: %s", string(contents))
		return err
	}
	decoder = k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(contentBytes), 1000)
	if resource.Kind == "Deployment" {
		controller := appsv1.Deployment{}
		err = decoder.Decode(&controller)
		resources.Deployments = append(resources.Deployments, controller)
	} else if resource.Kind == "StatefulSet" {
		controller := appsv1.StatefulSet{}
		err = decoder.Decode(&controller)
		resources.StatefulSets = append(resources.StatefulSets, controller)
	} else if resource.Kind == "DaemonSet" {
		controller := appsv1.DaemonSet{}
		err = decoder.Decode(&controller)
		resources.DaemonSets = append(resources.DaemonSets, controller)
	} else if resource.Kind == "Job" {
		controller := batchv1.Job{}
		err = decoder.Decode(&controller)
		resources.Jobs = append(resources.Jobs, controller)
	} else if resource.Kind == "CronJob" {
		controller := batchv1beta1.CronJob{}
		err = decoder.Decode(&controller)
		resources.CronJobs = append(resources.CronJobs, controller)
	} else if resource.Kind == "ReplicationController" {
		controller := corev1.ReplicationController{}
		err = decoder.Decode(&controller)
		resources.ReplicationControllers = append(resources.ReplicationControllers, controller)
	} else if resource.Kind == "Namespace" {
		ns := corev1.Namespace{}
		err = decoder.Decode(&ns)
		resources.Namespaces = append(resources.Namespaces, ns)
	} else if resource.Kind == "Pod" {
		pod := corev1.Pod{}
		err = decoder.Decode(&pod)
		resources.Pods = append(resources.Pods, pod)
	}
	if err != nil {
		logrus.Errorf("Error parsing %s: %v", resource.Kind, err)
		return err
	}
	return nil
}

func getDeployments(kube kubernetes.Interface) ([]appsv1.Deployment, error) {
	listOpts := metav1.ListOptions{}
	deployList, err := kube.AppsV1().Deployments("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Deployments: %v", err)
		return nil, err
	}
	deploys := deployList.Items

	oldDeploys := make([]interface{}, 0)
	deploysV1B1, err := kube.AppsV1beta1().Deployments("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Deployments v1beta1: %v", err)
		return nil, err
	}
	for _, oldDeploy := range deploysV1B1.Items {
		oldDeploys = append(oldDeploys, oldDeploy)
	}
	deploysV1B2, err := kube.AppsV1beta2().Deployments("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching Deployments v1beta2: %v", err)
		return nil, err
	}
	for _, oldDeploy := range deploysV1B2.Items {
		oldDeploys = append(oldDeploys, oldDeploy)
	}

	for _, oldDeploy := range oldDeploys {
		str, err := json.Marshal(oldDeploy)
		if err != nil {
			logrus.Errorf("Error marshaling old deployment version: %v", err)
			return nil, err
		}
		deploy := appsv1.Deployment{}
		err = json.Unmarshal(str, &deploy)
		if err != nil {
			logrus.Errorf("Error unmarshaling old deployment version: %v", err)
			return nil, err
		}
		deploys = append(deploys, deploy)
	}
	return deploys, nil
}

func getStatefulSets(kube kubernetes.Interface) ([]appsv1.StatefulSet, error) {
	listOpts := metav1.ListOptions{}
	controllerList, err := kube.AppsV1().StatefulSets("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching StatefulSets: %v", err)
		return nil, err
	}
	controllers := controllerList.Items

	oldControllers := make([]interface{}, 0)
	controllersV1B1, err := kube.AppsV1beta1().StatefulSets("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching StatefulSets v1beta1: %v", err)
		return nil, err
	}
	for _, oldController := range controllersV1B1.Items {
		oldControllers = append(oldControllers, oldController)
	}
	controllersV1B2, err := kube.AppsV1beta2().StatefulSets("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching StatefulSets v1beta2: %v", err)
		return nil, err
	}
	for _, oldController := range controllersV1B2.Items {
		oldControllers = append(oldControllers, oldController)
	}

	for _, oldController := range oldControllers {
		str, err := json.Marshal(oldController)
		if err != nil {
			logrus.Errorf("Error marshaling old StatefulSet version: %v", err)
			return nil, err
		}
		controller := appsv1.StatefulSet{}
		err = json.Unmarshal(str, &controller)
		if err != nil {
			logrus.Errorf("Error unmarshaling old StatefulSet version: %v", err)
			return nil, err
		}
		controllers = append(controllers, controller)
	}
	return controllers, nil
}

func getDaemonSets(kube kubernetes.Interface) ([]appsv1.DaemonSet, error) {
	listOpts := metav1.ListOptions{}
	controllerList, err := kube.AppsV1().DaemonSets("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching DaemonSets: %v", err)
		return nil, err
	}
	controllers := controllerList.Items

	controllersV1B2, err := kube.AppsV1beta2().DaemonSets("").List(listOpts)
	if err != nil {
		logrus.Errorf("Error fetching DaemonSets v1beta2: %v", err)
		return nil, err
	}

	for _, oldController := range controllersV1B2.Items {
		str, err := json.Marshal(oldController)
		if err != nil {
			logrus.Errorf("Error marshaling old DaemonSet version: %v", err)
			return nil, err
		}
		controller := appsv1.DaemonSet{}
		err = json.Unmarshal(str, &controller)
		if err != nil {
			logrus.Errorf("Error unmarshaling old DaemonSet version: %v", err)
			return nil, err
		}
		controllers = append(controllers, controller)
	}
	return controllers, nil
}
