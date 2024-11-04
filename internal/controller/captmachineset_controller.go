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
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// requeuePeriod is the period for requeuing
	requeuePeriod = 30 * time.Second

	// MachineFinalizer allows CaptMachineSetReconciler to clean up resources associated with CaptMachineSet before
	// removing it from the apiserver.
	MachineFinalizer = "captmachineset.infrastructure.cluster.x-k8s.io"
)

// CaptMachineSetReconciler reconciles a CaptMachineSet object
type CaptMachineSetReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinesets/finalizers,verbs=update
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachines,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles CaptMachineSet reconciliation
func (r *CaptMachineSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CaptMachineSet instance
	machineSet := &infrastructurev1beta1.CaptMachineSet{}
	if err := r.Get(ctx, req.NamespacedName, machineSet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !machineSet.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machineSet)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(machineSet, MachineFinalizer) {
		controllerutil.AddFinalizer(machineSet, MachineFinalizer)
		if err := r.Update(ctx, machineSet); err != nil {
			return ctrl.Result{}, err
		}
	}

	// List all child Machines
	machines, err := r.listMachines(ctx, machineSet)
	if err != nil {
		logger.Error(err, "Failed to list machines")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, machineSet, machines); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	// Reconcile machines
	if err := r.reconcileMachines(ctx, machineSet, machines); err != nil {
		logger.Error(err, "Failed to reconcile machines")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles CaptMachineSet deletion
func (r *CaptMachineSetReconciler) reconcileDelete(ctx context.Context, machineSet *infrastructurev1beta1.CaptMachineSet) (ctrl.Result, error) {
	machines, err := r.listMachines(ctx, machineSet)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Delete all child machines
	for i := range machines {
		if err := r.Delete(ctx, &machines[i]); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Remove finalizer if all machines are deleted
	if len(machines) == 0 {
		controllerutil.RemoveFinalizer(machineSet, MachineFinalizer)
		return ctrl.Result{}, r.Update(ctx, machineSet)
	}

	return ctrl.Result{RequeueAfter: requeuePeriod}, nil
}

// listMachines returns all machines owned by the machine set
func (r *CaptMachineSetReconciler) listMachines(ctx context.Context, machineSet *infrastructurev1beta1.CaptMachineSet) ([]infrastructurev1beta1.CaptMachine, error) {
	var machines infrastructurev1beta1.CaptMachineList
	if err := r.List(ctx, &machines, client.InNamespace(machineSet.Namespace), client.MatchingLabels(machineSet.Spec.Selector.MatchLabels)); err != nil {
		return nil, err
	}

	// Filter out machines that don't belong to this machine set
	var owned []infrastructurev1beta1.CaptMachine
	for _, machine := range machines.Items {
		if metav1.IsControlledBy(&machine, machineSet) {
			owned = append(owned, machine)
		}
	}

	return owned, nil
}

// updateStatus updates the status of the machine set
func (r *CaptMachineSetReconciler) updateStatus(ctx context.Context, machineSet *infrastructurev1beta1.CaptMachineSet, machines []infrastructurev1beta1.CaptMachine) error {
	newStatus := infrastructurev1beta1.CaptMachineSetStatus{
		Replicas:      int32(len(machines)),
		ReadyReplicas: 0,
	}

	// Count ready replicas
	for _, machine := range machines {
		if machine.Status.Ready {
			newStatus.ReadyReplicas++
		}
	}

	// Update observed generation
	newStatus.ObservedGeneration = machineSet.Generation

	// Update status if it has changed
	if !reflect.DeepEqual(machineSet.Status, newStatus) {
		machineSet.Status = newStatus
		return r.Status().Update(ctx, machineSet)
	}

	return nil
}

// reconcileMachines reconciles the machines owned by the machine set
func (r *CaptMachineSetReconciler) reconcileMachines(ctx context.Context, machineSet *infrastructurev1beta1.CaptMachineSet, machines []infrastructurev1beta1.CaptMachine) error {
	// Get the number of desired replicas
	replicas := int32(1)
	if machineSet.Spec.Replicas != nil {
		replicas = *machineSet.Spec.Replicas
	}

	diff := len(machines) - int(replicas)

	if diff < 0 {
		// Scale up
		return r.createMachines(ctx, machineSet, -diff)
	} else if diff > 0 {
		// Scale down
		return r.deleteMachines(ctx, machines[0:diff])
	}

	return nil
}

// createMachines creates count new machines from the template
func (r *CaptMachineSetReconciler) createMachines(ctx context.Context, machineSet *infrastructurev1beta1.CaptMachineSet, count int) error {
	for i := 0; i < count; i++ {
		machine := &infrastructurev1beta1.CaptMachine{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", machineSet.Name),
				Namespace:    machineSet.Namespace,
				Labels:       machineSet.Spec.Template.ObjectMeta.Labels,
			},
			Spec: *machineSet.Spec.Template.Spec.DeepCopy(),
		}

		if err := controllerutil.SetControllerReference(machineSet, machine, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, machine); err != nil {
			return err
		}
	}

	return nil
}

// deleteMachines deletes the specified machines
func (r *CaptMachineSetReconciler) deleteMachines(ctx context.Context, machines []infrastructurev1beta1.CaptMachine) error {
	for i := range machines {
		if err := r.Delete(ctx, &machines[i]); err != nil {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CaptMachineSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CaptMachineSet{}).
		Owns(&infrastructurev1beta1.CaptMachine{}).
		Complete(r)
}
