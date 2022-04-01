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
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func newUnstructured(apiVersion, kind, namespace, name string, spec map[string]interface{}) unstructured.Unstructured {
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
func MockController(apiVersion, kind, namespace, name string, spec map[string]interface{}, podSpec corev1.PodSpec, dest interface{}) corev1.Pod {
	unst := newUnstructured(apiVersion, kind, namespace, name, spec)
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
	b, err := unst.MarshalJSON()
	if err != nil {
		panic(err)
	}
	json.Unmarshal(b, &dest)
	return pod
}

// MockControllerWithNormalSpec mocks a controller with podspec at spec.template.spec
func MockControllerWithNormalSpec(apiVersion, kind, namespace, name string, dest interface{}) corev1.Pod {
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
	return MockController(apiVersion, kind, namespace, name, spec, p.Spec, dest)
}

// MockDeploy creates a Deployment object.
func MockDeploy(namespace, name string) (appsv1.Deployment, corev1.Pod) {
	d := appsv1.Deployment{}
	pod := MockControllerWithNormalSpec("apps/v1", "Deployment", namespace, name, &d)
	return d, pod
}

// MockStatefulSet creates a StatefulSet object.
func MockStatefulSet(namespace, name string) (appsv1.StatefulSet, corev1.Pod) {
	s := appsv1.StatefulSet{}
	pod := MockControllerWithNormalSpec("apps/v1", "StatefulSet", namespace, name, &s)
	return s, pod
}

// MockDaemonSet creates a DaemonSet object.
func MockDaemonSet(namespace, name string) (appsv1.DaemonSet, corev1.Pod) {
	d := appsv1.DaemonSet{}
	pod := MockControllerWithNormalSpec("apps/v1", "DaemonSet", namespace, name, &d)
	return d, pod
}

// MockJob creates a Job object.
func MockJob(namespace, name string) (batchv1.Job, corev1.Pod) {
	j := batchv1.Job{}
	pod := MockControllerWithNormalSpec("batch/v1", "Job", namespace, name, &j)
	return j, pod
}

// MockCronJob creates a CronJob object.
func MockCronJob(namespace, name string) (batchv1beta1.CronJob, corev1.Pod) {
	cj := batchv1beta1.CronJob{}
	p := MockPod()
	spec := map[string]interface{}{}
	pod := MockController("batch/v1beta1", "CronJob", namespace, name, spec, p.Spec, &cj)
	cj.Spec.JobTemplate.Spec.Template.Spec = pod.Spec

	return cj, pod
}

// MockReplicationController creates a ReplicationController object.
func MockReplicationController(namespace, name string) (corev1.ReplicationController, corev1.Pod) {
	rc := corev1.ReplicationController{}
	pod := MockControllerWithNormalSpec("core/v1", "ReplicationController", namespace, name, &rc)
	return rc, pod
}

// SetupTestAPI creates a test kube API struct.
func SetupTestAPI(objects ...runtime.Object) (kubernetes.Interface, dynamic.Interface) {
	scheme := runtime.NewScheme()
	appsv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	policyv1beta1.AddToScheme(scheme)
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
		{
			GroupVersion: "networking.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{Name: "ingresses", Namespaced: true, Kind: "Ingress", Version: "v1"},
			},
		},
		{
			GroupVersion: policyv1beta1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "poddisruptionbudgets", Namespaced: true, Kind: "PodDisruptionBudget", Version: "v1"},
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
