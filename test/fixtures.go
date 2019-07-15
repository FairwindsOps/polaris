package test

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

func mockLimitRange() corev1.LimitRange {
	lr := corev1.LimitRange{
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				corev1.LimitRangeItem{
					Type:           corev1.LimitTypeContainer,
					Default:        make(map[corev1.ResourceName]resource.Quantity),
					DefaultRequest: make(map[corev1.ResourceName]resource.Quantity),
				},
			},
		},
	}
	lr.Spec.Limits[0].Default[corev1.ResourceCPU] = resource.MustParse("500m")
	lr.Spec.Limits[0].Default[corev1.ResourceMemory] = resource.MustParse("5Gi")
	lr.Spec.Limits[0].DefaultRequest[corev1.ResourceCPU] = resource.MustParse("500m")
	lr.Spec.Limits[0].DefaultRequest[corev1.ResourceMemory] = resource.MustParse("5Gi")
	return lr
}

// SetupTestAPI creates a test kube API struct.
func SetupTestAPI() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

// SetupAddControllers creates mock controllers and adds them to the test clientset.
func SetupAddControllers(k kubernetes.Interface, namespace string) {
	d1 := mockDeploy()
	_, err := k.AppsV1().Deployments(namespace).Create(&d1)
	if err != nil {
		panic(err)
	}
	s1 := mockStatefulSet()
	_, err = k.AppsV1().StatefulSets(namespace).Create(&s1)
	if err != nil {
		panic(err)
	}
}

// SetupAddLimitRanges creates mock limit ranges and adds them to the test clientset.
func SetupAddLimitRanges(k kubernetes.Interface, namespace string) {
	lr := mockLimitRange()
	_, err := k.CoreV1().LimitRanges(namespace).Create(&lr)
	if err != nil {
		panic(err)
	}
}
