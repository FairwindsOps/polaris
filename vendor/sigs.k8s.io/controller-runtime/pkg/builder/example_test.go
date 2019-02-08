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

package builder_test

import (
	"context"
	"fmt"
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// This example creates a simple application ControllerManagedBy that is configured for ReplicaSets and Pods.
//
// * Create a new application for ReplicaSets that manages Pods owned by the ReplicaSet and calls into
// ReplicaSetReconciler.
//
// * Start the application.
func ExampleBuilder() {
	var log = logf.Log.WithName("builder-examples")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	_, err = builder.
		ControllerManagedBy(mgr).  // Create the ControllerManagedBy
		For(&appsv1.ReplicaSet{}). // ReplicaSet is the Application API
		Owns(&corev1.Pod{}).       // ReplicaSet owns Pods created by it
		Build(&ReplicaSetReconciler{})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

// ReplicaSetReconciler is a simple ControllerManagedBy example implementation.
type ReplicaSetReconciler struct {
	client.Client
}

// Implement the business logic:
// This function will be called when there is a change to a ReplicaSet or a Pod with an OwnerReference
// to a ReplicaSet.
//
// * Read the ReplicaSet
// * Read the Pods
// * Set a Label on the ReplicaSet with the Pod count
func (a *ReplicaSetReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	// Read the ReplicaSet
	rs := &appsv1.ReplicaSet{}
	err := a.Get(context.TODO(), req.NamespacedName, rs)
	if err != nil {
		return reconcile.Result{}, err
	}

	// List the Pods matching the PodTemplate Labels
	pods := &corev1.PodList{}
	err = a.List(context.TODO(), client.InNamespace(req.Namespace).MatchingLabels(rs.Spec.Template.Labels), pods)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Update the ReplicaSet
	rs.Labels["pod-count"] = fmt.Sprintf("%v", len(pods.Items))
	err = a.Update(context.TODO(), rs)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (a *ReplicaSetReconciler) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}
