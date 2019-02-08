/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package builder

import (
	"context"
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("application", func() {
	var stop chan struct{}

	BeforeEach(func() {
		stop = make(chan struct{})
		getConfig = func() (*rest.Config, error) { return cfg, nil }
		newController = controller.New
		newManager = manager.New
		getGvk = apiutil.GVKForObject
	})

	AfterEach(func() {
		close(stop)
	})

	noop := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) { return reconcile.Result{}, nil })

	Describe("New", func() {
		It("should return success if given valid objects", func() {
			instance, err := SimpleController().
				For(&appsv1.ReplicaSet{}).
				Owns(&appsv1.ReplicaSet{}).
				Build(noop)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).NotTo(BeNil())
		})

		It("should return an error if the Config is invalid", func() {
			getConfig = func() (*rest.Config, error) { return cfg, fmt.Errorf("expected error") }
			instance, err := SimpleController().
				For(&appsv1.ReplicaSet{}).
				Owns(&appsv1.ReplicaSet{}).
				Build(noop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected error"))
			Expect(instance).To(BeNil())
		})

		It("should return an error if there is no GVK for an object", func() {
			instance, err := SimpleController().
				For(&fakeType{}).
				Owns(&appsv1.ReplicaSet{}).
				Build(noop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no kind is registered for the type builder.fakeType"))
			Expect(instance).To(BeNil())

			instance, err = SimpleController().
				For(&appsv1.ReplicaSet{}).
				Owns(&fakeType{}).
				Build(noop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no kind is registered for the type builder.fakeType"))
			Expect(instance).To(BeNil())
		})

		It("should return an error if it cannot create the manager", func() {
			newManager = func(config *rest.Config, options manager.Options) (manager.Manager, error) {
				return nil, fmt.Errorf("expected error")
			}
			instance, err := SimpleController().
				For(&appsv1.ReplicaSet{}).
				Owns(&appsv1.ReplicaSet{}).
				Build(noop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected error"))
			Expect(instance).To(BeNil())
		})

		It("should return an error if it cannot create the controller", func() {
			newController = func(name string, mgr manager.Manager, options controller.Options) (
				controller.Controller, error) {
				return nil, fmt.Errorf("expected error")
			}
			instance, err := SimpleController().
				For(&appsv1.ReplicaSet{}).
				Owns(&appsv1.ReplicaSet{}).
				Build(noop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected error"))
			Expect(instance).To(BeNil())
		})
	})

	Describe("Start with SimpleController", func() {
		It("should Reconcile Owns objects", func(done Done) {
			bldr := SimpleController().
				ForType(&appsv1.Deployment{}).
				WithConfig(cfg).
				Owns(&appsv1.ReplicaSet{})
			doReconcileTest("1", stop, bldr, nil, false)

			close(done)
		}, 10)

		It("should Reconcile Owns objects with a Manager", func(done Done) {
			m, err := manager.New(cfg, manager.Options{})
			Expect(err).NotTo(HaveOccurred())

			bldr := SimpleController().
				WithManager(m).
				For(&appsv1.Deployment{}).
				Owns(&appsv1.ReplicaSet{})
			doReconcileTest("2", stop, bldr, m, false)
			close(done)
		}, 10)
	})

	Describe("Start with ControllerManagedBy", func() {
		It("should Reconcile Owns objects", func(done Done) {
			m, err := manager.New(cfg, manager.Options{})
			Expect(err).NotTo(HaveOccurred())

			bldr := ControllerManagedBy(m).
				For(&appsv1.Deployment{}).
				Owns(&appsv1.ReplicaSet{})
			doReconcileTest("3", stop, bldr, m, false)
			close(done)
		}, 10)

		It("should Reconcile Watches objects", func(done Done) {
			m, err := manager.New(cfg, manager.Options{})
			Expect(err).NotTo(HaveOccurred())

			bldr := ControllerManagedBy(m).
				For(&appsv1.Deployment{}).
				Watches( // Equivalent of Owns
					&source.Kind{Type: &appsv1.ReplicaSet{}},
					&handler.EnqueueRequestForOwner{OwnerType: &appsv1.Deployment{}, IsController: true})
			doReconcileTest("4", stop, bldr, m, true)
			close(done)
		}, 10)
	})
})

func doReconcileTest(nameSuffix string, stop chan struct{}, blder *Builder, mgr manager.Manager, complete bool) {
	deployName := "deploy-name-" + nameSuffix
	rsName := "rs-name-" + nameSuffix

	By("Creating the application")
	ch := make(chan reconcile.Request)
	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		defer GinkgoRecover()
		if !strings.HasSuffix(req.Name, nameSuffix) {
			// From different test, ignore this request.  Etcd is shared across tests.
			return reconcile.Result{}, nil
		}
		ch <- req
		return reconcile.Result{}, nil
	})

	instance := mgr
	if complete {
		err := blder.Complete(fn)
		Expect(err).NotTo(HaveOccurred())
	} else {
		var err error
		instance, err = blder.Build(fn)
		Expect(err).NotTo(HaveOccurred())
	}

	// Manager should match
	if mgr != nil {
		Expect(instance).To(Equal(mgr))
	}

	By("Starting the application")
	go func() {
		defer GinkgoRecover()
		Expect(instance.Start(stop)).NotTo(HaveOccurred())
		By("Stopping the application")
	}()

	By("Creating a Deployment")
	// Expect a Reconcile when the Deployment is managedObjects.
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      deployName,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "bar"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
	err := instance.GetClient().Create(context.TODO(), dep)
	Expect(err).NotTo(HaveOccurred())

	By("Waiting for the Deployment Reconcile")
	Expect(<-ch).To(Equal(reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: deployName}}))

	By("Creating a ReplicaSet")
	// Expect a Reconcile when an Owned object is managedObjects.
	t := true
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      rsName,
			Labels:    dep.Spec.Selector.MatchLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       deployName,
					Kind:       "Deployment",
					APIVersion: "apps/v1",
					Controller: &t,
					UID:        dep.UID,
				},
			},
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: dep.Spec.Selector,
			Template: dep.Spec.Template,
		},
	}
	err = instance.GetClient().Create(context.TODO(), rs)
	Expect(err).NotTo(HaveOccurred())

	By("Waiting for the ReplicaSet Reconcile")
	Expect(<-ch).To(Equal(reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: deployName}}))

}

var _ runtime.Object = &fakeType{}

type fakeType struct{}

func (*fakeType) GetObjectKind() schema.ObjectKind { return nil }
func (*fakeType) DeepCopyObject() runtime.Object   { return nil }
