package test

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func mockContainer(name string) corev1.Container {
	c := corev1.Container{
		Name: name,
	}
	return c
}

// MockPod creates a pod object.
func MockPod() corev1.PodTemplateSpec {
	c1 := mockContainer("test")
	p := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				c1,
			},
		},
	}
	return p
}

func mockDeploy() appsv1.Deployment {
	p := MockPod()
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: p,
		},
	}
	return d
}

func mockStatefulSet() appsv1.StatefulSet {
	p := MockPod()
	s := appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Template: p,
		},
	}
	return s
}

// SetupTestAPI creates a test kube API struct.
func SetupTestAPI() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

// SetupAddControllers creates mock controllers and adds them to the test clientset.
func SetupAddControllers(k kubernetes.Interface, namespace string) kubernetes.Interface {
	d1 := mockDeploy()
	_, err := k.AppsV1().Deployments(namespace).Create(&d1)
	if err != nil {
		fmt.Println(err)
	}
	s1 := mockStatefulSet()
	_, err = k.AppsV1().StatefulSets(namespace).Create(&s1)
	if err != nil {
		fmt.Println(err)
	}
	return k
}
