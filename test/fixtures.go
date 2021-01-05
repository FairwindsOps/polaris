package test

import (
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func newUnstructured(apiVersion, kind, namespace, name string, spec interface{}) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
			"spec": spec,
		},
	}
}

// MockContainer creates a container object
func MockContainer(name string) corev1.Container {
	c := corev1.Container{
		Name: name,
	}
	return c
}

// MockPod creates a pod object.
func MockPod() corev1.Pod {
	c1 := MockContainer("test")
	p := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				c1,
			},
		},
	}
	return p
}

// MockNakedPod creates a pod object.
func MockNakedPod() corev1.Pod {
	return corev1.Pod{
		Spec: MockPod().Spec,
	}
}

// MockIngress creates an ingress object
func MockIngress() extv1beta1.Ingress {
	return extv1beta1.Ingress{
		Spec: extv1beta1.IngressSpec{},
	}
}

// MockController creates a mock controller and pod
func MockController(apiVersion, kind, namespace, name string, spec interface{}, podSpec corev1.PodSpec) (unstructured.Unstructured, corev1.Pod) {
	d := newUnstructured(apiVersion, kind, namespace, name, spec)
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-12345",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: apiVersion,
				Kind:       kind,
				Name:       name,
			}},
		},
		Spec: podSpec,
	}
	return d, pod
}

// MockControllerWithNormalSpec mocks a controller with podspec at spec.template.spec
func MockControllerWithNormalSpec(apiVersion, kind, namespace, name string) (unstructured.Unstructured, corev1.Pod) {
	p := MockPod()
	b, err := json.Marshal(p.Spec)
	if err != nil {
		panic(err)
	}
	pSpec := map[string]interface{}{}
	err = json.Unmarshal(b, &pSpec)
	if err != nil {
		panic(err)
	}
	spec := map[string]interface{}{
		"template": map[string]interface{}{
			"spec": pSpec,
		},
	}
	return MockController(apiVersion, kind, namespace, name, spec, p.Spec)
}

// MockDeploy creates a Deployment object.
func MockDeploy(namespace, name string) (unstructured.Unstructured, corev1.Pod) {
	return MockControllerWithNormalSpec("apps/v1", "Deployment", namespace, name)
}

// MockStatefulSet creates a StatefulSet object.
func MockStatefulSet(namespace, name string) (unstructured.Unstructured, corev1.Pod) {
	return MockControllerWithNormalSpec("apps/v1", "StatefulSet", namespace, name)
}

// MockDaemonSet creates a DaemonSet object.
func MockDaemonSet(namespace, name string) (unstructured.Unstructured, corev1.Pod) {
	return MockControllerWithNormalSpec("apps/v1", "DaemonSet", namespace, name)
}

// MockJob creates a Job object.
func MockJob(namespace, name string) (unstructured.Unstructured, corev1.Pod) {
	return MockControllerWithNormalSpec("batch/v1", "Job", namespace, name)
}

// MockCronJob creates a CronJob object.
func MockCronJob(namespace, name string) (unstructured.Unstructured, corev1.Pod) {
	p := MockPod()
	b, err := json.Marshal(p.Spec)
	if err != nil {
		panic(err)
	}
	pSpec := map[string]interface{}{}
	err = json.Unmarshal(b, &pSpec)
	if err != nil {
		panic(err)
	}
	spec := map[string]interface{}{
		"job_template": map[string]interface{}{
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": pSpec,
				},
			},
		},
	}
	return MockController("batch/v1beta1", "CronJob", namespace, name, spec, p.Spec)
}

// MockReplicationController creates a ReplicationController object.
func MockReplicationController(namespace, name string) (unstructured.Unstructured, corev1.Pod) {
	return MockControllerWithNormalSpec("core/v1", "ReplicationController", namespace, name)
}

// SetupTestAPI creates a test kube API struct.
func SetupTestAPI(objects ...runtime.Object) (kubernetes.Interface, dynamic.Interface) {
	scheme := runtime.NewScheme()
	appsv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	fake.AddToScheme(scheme)
	dynamicClient := dynamicFake.NewSimpleDynamicClient(scheme, objects...)
	k := fake.NewSimpleClientset(objects...)
	k.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: corev1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
				{Name: "replicationcontrollers", Namespaced: true, Kind: "ReplicationController"},
			},
		},
		{
			GroupVersion: appsv1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "deployments", Namespaced: true, Kind: "Deployment"},
				{Name: "daemonsets", Namespaced: true, Kind: "DaemonSet"},
				{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet"},
			},
		},
		{
			GroupVersion: batchv1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "jobs", Namespaced: true, Kind: "Job"},
			},
		},
		{
			GroupVersion: batchv1beta1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "cronjobs", Namespaced: true, Kind: "CronJob"},
			},
		},
		{
			GroupVersion: appsv1beta2.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "deployments", Namespaced: true, Kind: "Deployment"},
				{Name: "deployments/scale", Namespaced: true, Kind: "Scale", Group: "apps", Version: "v1beta2"},
			},
		},
		{
			GroupVersion: appsv1beta1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet"},
				{Name: "statefulsets/scale", Namespaced: true, Kind: "Scale", Group: "apps", Version: "v1beta1"},
			},
		},
	}
	return k, dynamicClient
}

// GetMockControllers returns mocked controllers for 5 major controller types
func GetMockControllers(namespace string) []runtime.Object {
	deploy, deployPod := MockDeploy(namespace, "deploy")
	statefulset, statefulsetPod := MockStatefulSet(namespace, "statefulset")
	daemonset, daemonsetPod := MockDaemonSet(namespace, "daemonset")
	job, jobPod := MockJob(namespace, "job")
	cronjob, cronjobPod := MockCronJob(namespace, "cronjob")
	return []runtime.Object{
		&deploy, &deployPod,
		&daemonset, &daemonsetPod,
		&statefulset, &statefulsetPod,
		&cronjob, &cronjobPod,
		&job, &jobPod,
	}
}
