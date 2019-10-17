package test

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// MockContainer creates a container object
func MockContainer(name string) corev1.Container {
	c := corev1.Container{
		Name: name,
	}
	return c
}

// MockPod creates a pod object.
func MockPod() corev1.PodTemplateSpec {
	c1 := MockContainer("test")
	p := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				c1,
			},
		},
	}
	return p
}

// MockDeploy creates a Deployment object.
func MockDeploy() appsv1.Deployment {
	p := MockPod()
	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: p,
		},
	}
	return d
}

// MockStatefulSet creates a StatefulSet object.
func MockStatefulSet() appsv1.StatefulSet {
	p := MockPod()
	s := appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Template: p,
		},
	}
	return s
}

// MockDaemonSet creates a DaemonSet object.
func MockDaemonSet() appsv1.DaemonSet {
	return appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: MockPod(),
		},
	}
}

// MockJob creates a Job object.
func MockJob() batchv1.Job {
	return batchv1.Job{
		Spec: batchv1.JobSpec{
			Template: MockPod(),
		},
	}
}

// MockCronJob creates a CronJob object.
func MockCronJob() batchv1beta1.CronJob {
	return batchv1beta1.CronJob{
		Spec: batchv1beta1.CronJobSpec{
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: MockPod(),
				},
			},
		},
	}
}

// MockReplicationController creates a ReplicationController object.
func MockReplicationController() corev1.ReplicationController {
	p := MockPod()
	return corev1.ReplicationController{
		Spec: corev1.ReplicationControllerSpec{
			Template: &p,
		},
	}
}

// SetupTestAPI creates a test kube API struct.
func SetupTestAPI() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

// SetupAddControllers creates mock controllers and adds them to the test clientset.
func SetupAddControllers(k kubernetes.Interface, namespace string) kubernetes.Interface {
	d1 := MockDeploy()
	if _, err := k.AppsV1().Deployments(namespace).Create(&d1); err != nil {
		fmt.Println(err)
	}

	s1 := MockStatefulSet()
	if _, err := k.AppsV1().StatefulSets(namespace).Create(&s1); err != nil {
		fmt.Println(err)
	}

	ds1 := MockDaemonSet()
	if _, err := k.AppsV1().DaemonSets(namespace).Create(&ds1); err != nil {
		fmt.Println(err)
	}

	j1 := MockJob()
	if _, err := k.BatchV1().Jobs(namespace).Create(&j1); err != nil {
		fmt.Println(err)
	}

	cj1 := MockCronJob()
	if _, err := k.BatchV1beta1().CronJobs(namespace).Create(&cj1); err != nil {
		fmt.Println(err)
	}

	rc1 := MockReplicationController()
	if _, err := k.CoreV1().ReplicationControllers(namespace).Create(&rc1); err != nil {
		fmt.Println(err)
	}

	return k
}
