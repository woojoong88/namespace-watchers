/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const (
	namespaceWatcherFinalizer = "microsoft.com/namespace-watcher"
	EnvKeyExcludedNamespaces  = "EXCLUDED_NAMESPACES"
)

var (
	// predicateFuncs is used to filter events
	predicateFuncs = predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// case if namespace is being deleted
			if e.ObjectOld.GetDeletionTimestamp().IsZero() && !e.ObjectNew.GetDeletionTimestamp().IsZero() {
				return true
			}

			// case if finalizer is added to namespace
			if controllerutil.ContainsFinalizer(e.ObjectNew, namespaceWatcherFinalizer) &&
				!controllerutil.ContainsFinalizer(e.ObjectOld, namespaceWatcherFinalizer) {
				return true
			}

			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	ExcludedNamespaces []string
	StartTime          time.Time
}

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=namespaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Namespace object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// get namespace
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		// ignore not found error - the namespace might have been deleted right before reconcile loop
		if err = client.IgnoreNotFound(err); err != nil {
			log.Error(err, " failed to get namespace")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// skip namespaces defined in ExcludedNamespaces
	if r.ExcludedNamespaces != nil {
		for _, excludedNs := range r.ExcludedNamespaces {
			if ns.Name == excludedNs {
				return ctrl.Result{}, nil
			}
		}
	}

	// case 1: namespace is being created
	// add finalizer to namespace if the namespace is being created
	// finalizer is used for avoiding the race condition among multiple controllers
	if !controllerutil.ContainsFinalizer(ns, namespaceWatcherFinalizer) && ns.ObjectMeta.DeletionTimestamp.IsZero() && r.isCreatedAfterStartTime(ns) {
		log.Info(fmt.Sprintf("Namespace %s is created - %+v", ns.Name, ns.GetObjectMeta()))

		// add finalizer to namespace
		controllerutil.AddFinalizer(ns, namespaceWatcherFinalizer)
		if err := r.Update(ctx, ns); err != nil {
			log.Error(err, " failed to update namespace to add finalizer")
			return ctrl.Result{}, err // will retry
		}
		return ctrl.Result{}, nil
	}

	// case 2: namespace is being deleted
	// remove finalizer from namespace if the namespace is being deleted
	if !ns.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info(fmt.Sprintf("Namespace %s is being deleted - %+v", ns.Name, ns.GetObjectMeta()))

		// remove finalizer from namespace before delete
		if controllerutil.ContainsFinalizer(ns, namespaceWatcherFinalizer) {
			controllerutil.RemoveFinalizer(ns, namespaceWatcherFinalizer)
			if err := r.Update(ctx, ns); err != nil {
				log.Error(err, " failed to update namespace to remove finalizer")
				return ctrl.Result{}, err // will retry
			}
		}
		// terminate reconcile loop for case 2
		return ctrl.Result{}, nil
	}

	// then, create a pod in the namespace for case 1
	// terminate reconcile loop for case 2 if there is no error; otherwise, will retry
	return r.createEchoPod(ctx, ns)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set start time to check if the namespace is created after the controller starts
	r.StartTime = time.Now()

	// run controller to watch for namespace events
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}, builder.WithPredicates(predicateFuncs)).
		Complete(r)
}

// isCreatedAfterStartTime checks if the namespace is created after the controller starts
func (r *NamespaceReconciler) isCreatedAfterStartTime(namespace *corev1.Namespace) bool {
	return namespace.CreationTimestamp.Time.After(r.StartTime)
}

// createEchoPod creates a pod in the namespace with echo message
func (r *NamespaceReconciler) createEchoPod(ctx context.Context, namespace *corev1.Namespace) (ctrl.Result, error) {

	// skip if namespace is created before the controller starts
	if !r.isCreatedAfterStartTime(namespace) {
		return ctrl.Result{}, nil
	}

	log := log.FromContext(ctx)

	// echo message
	echoMsg := fmt.Sprintf("echo namespace: %s", namespace.Name)

	// create pod object
	pod := &corev1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      "echo-pod",
			Namespace: namespace.Name,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "echo",
					Image: "busybox:stable",
					Args:  []string{"sh", "-c", echoMsg},
				},
			},
		},
	}

	log.Info(fmt.Sprintf("Creating pod %s in namespace %s: spec %+v", pod.Name, pod.Namespace, pod.Spec))

	// check if pod already exists
	existingPod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Namespace: pod.Namespace, Name: pod.Name}, existingPod)
	if err == nil {
		log.Info("Pod already exists", "Pod", pod.Name, "Namespace", pod.Namespace)
		return reconcile.Result{}, nil
	}

	// failed to get pod to check if it already exists
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err // will retry
	}

	// create pod
	if err := r.Create(ctx, pod); err != nil {
		log.Error(err, " failed to create pod")
		return ctrl.Result{}, err // will retry
	}

	return ctrl.Result{}, nil
}
