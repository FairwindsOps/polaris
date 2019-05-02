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

// SetupTestAPI creates a test kube API struct.
func SetupTestAPI() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

// SetupAddDeploys creates a mock deployment and adds it to the test clientset.
func SetupAddDeploys(k kubernetes.Interface, namespace string) kubernetes.Interface {
	d1 := mockDeploy()
	_, err := k.AppsV1().Deployments(namespace).Create(&d1)
	if err != nil {
		fmt.Println(err)
	}
	return k
}
