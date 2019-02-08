package test

import (
	"fmt"

	"github.com/reactiveops/fairwinds/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func mockContainer(name string) corev1.Container {
	c := corev1.Container{
		Name: name,
	}
	return c
}

func mockPod() corev1.PodTemplateSpec {
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
	p := mockPod()
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: p,
		},
	}
	return d
}

func SetupTestAPI() *kube.API {
	api := kube.API{
		Clientset: fake.NewSimpleClientset(),
	}
	return &api
}

func SetupAddDeploys(k *kube.API, namespace string) *kube.API {
	d1 := mockDeploy()
	_, err := k.Clientset.AppsV1().Deployments(namespace).Create(&d1)
	if err != nil {
		fmt.Println(err)
	}
	return k
}
