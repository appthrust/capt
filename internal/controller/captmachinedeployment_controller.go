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
	// DeploymentFinalizer allows CaptMachineDeploymentReconciler to clean up resources associated with
	// CaptMachineDeployment before removing it from the apiserver.
	DeploymentFinalizer = "captmachinedeployment.infrastructure.cluster.x-k8s.io"

	// DefaultDeploymentUniqueLabelKey is the default key of the selector that is added
	// to existing MachineSets to prevent the existing MachineSets from selecting new machines.
	DefaultDeploymentUniqueLabelKey = "capt-deployment-hash"

	// DefaultRollingUpdateMaxUnavailable is the default value of MaxUnavailable for RollingUpdate strategy.
	DefaultRollingUpdateMaxUnavailable = 0

	// DefaultRollingUpdateMaxSurge is the default value of MaxSurge for RollingUpdate strategy.
	DefaultRollingUpdateMaxSurge = 1

	// DefaultRevisionHistoryLimit is the default value of RevisionHistoryLimit.
	DefaultRevisionHistoryLimit = 10

	// DefaultProgressDeadlineSeconds is the default value of ProgressDeadlineSeconds.
	DefaultProgressDeadlineSeconds = 600
)

// CaptMachineDeploymentReconciler reconciles a CaptMachineDeployment object
type CaptMachineDeploymentReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinedeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinedeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinedeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinesets,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles CaptMachineDeployment reconciliation
func (r *CaptMachineDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CaptMachineDeployment instance
	deployment := &infrastructurev1beta1.CaptMachineDeployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !deployment.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, deployment)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(deployment, DeploymentFinalizer) {
		controllerutil.AddFinalizer(deployment, DeploymentFinalizer)
		if err := r.Update(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	// List all MachineSets owned by this deployment
	machineSets, err := r.listMachineSets(ctx, deployment)
	if err != nil {
		logger.Error(err, "Failed to list machine sets")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, deployment, machineSets); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	// Check if the deployment is paused
	if deployment.Spec.Paused {
		return ctrl.Result{}, nil
	}

	// Check progress deadline
	if deployment.Spec.ProgressDeadlineSeconds != nil {
		if err := r.checkProgressDeadline(deployment); err != nil {
			logger.Error(err, "Failed to check progress deadline")
			return ctrl.Result{}, err
		}
	}

	// Reconcile MachineSets
	if err := r.reconcileMachineSets(ctx, deployment, machineSets); err != nil {
		logger.Error(err, "Failed to reconcile machine sets")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles CaptMachineDeployment deletion
func (r *CaptMachineDeploymentReconciler) reconcileDelete(ctx context.Context, deployment *infrastructurev1beta1.CaptMachineDeployment) (ctrl.Result, error) {
	machineSets, err := r.listMachineSets(ctx, deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Delete all MachineSets
	for i := range machineSets {
		if err := r.Delete(ctx, &machineSets[i]); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Remove finalizer if all MachineSets are deleted
	if len(machineSets) == 0 {
		controllerutil.RemoveFinalizer(deployment, DeploymentFinalizer)
		return ctrl.Result{}, r.Update(ctx, deployment)
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// listMachineSets returns all MachineSets owned by the deployment
func (r *CaptMachineDeploymentReconciler) listMachineSets(ctx context.Context, deployment *infrastructurev1beta1.CaptMachineDeployment) ([]infrastructurev1beta1.CaptMachineSet, error) {
	var machineSets infrastructurev1beta1.CaptMachineSetList
	if err := r.List(ctx, &machineSets, client.InNamespace(deployment.Namespace), client.MatchingLabels(deployment.Spec.Selector.MatchLabels)); err != nil {
		return nil, err
	}

	// Filter out MachineSets that don't belong to this deployment
	var owned []infrastructurev1beta1.CaptMachineSet
	for _, machineSet := range machineSets.Items {
		if metav1.IsControlledBy(&machineSet, deployment) {
			owned = append(owned, machineSet)
		}
	}

	return owned, nil
}

// updateStatus updates the status of the deployment
func (r *CaptMachineDeploymentReconciler) updateStatus(ctx context.Context, deployment *infrastructurev1beta1.CaptMachineDeployment, machineSets []infrastructurev1beta1.CaptMachineSet) error {
	newStatus := infrastructurev1beta1.CaptMachineDeploymentStatus{
		ObservedGeneration: deployment.Generation,
	}

	// Calculate replicas
	for _, machineSet := range machineSets {
		newStatus.Replicas += machineSet.Status.Replicas
		newStatus.UpdatedReplicas += machineSet.Status.ReadyReplicas
		newStatus.AvailableReplicas += machineSet.Status.ReadyReplicas
	}

	// Update status if it has changed
	if !reflect.DeepEqual(deployment.Status, newStatus) {
		deployment.Status = newStatus
		return r.Status().Update(ctx, deployment)
	}

	return nil
}

// checkProgressDeadline checks if the deployment has exceeded its progress deadline
func (r *CaptMachineDeploymentReconciler) checkProgressDeadline(deployment *infrastructurev1beta1.CaptMachineDeployment) error {
	if deployment.Status.ObservedGeneration < deployment.Generation {
		// No deadline exceeded because we haven't observed the latest revision yet
		return nil
	}

	deadline := DefaultProgressDeadlineSeconds
	if deployment.Spec.ProgressDeadlineSeconds != nil {
		deadline = int(*deployment.Spec.ProgressDeadlineSeconds)
	}

	// Check if the deployment is making progress
	if deployment.Status.UpdatedReplicas < deployment.Status.Replicas {
		// Calculate the time elapsed since the deployment started
		// TODO: Store the deployment start time in the status
		return fmt.Errorf("deployment %s/%s has not made progress within %d seconds", deployment.Namespace, deployment.Name, deadline)
	}

	return nil
}

// reconcileMachineSets reconciles the MachineSets owned by the deployment
func (r *CaptMachineDeploymentReconciler) reconcileMachineSets(ctx context.Context, deployment *infrastructurev1beta1.CaptMachineDeployment, machineSets []infrastructurev1beta1.CaptMachineSet) error {
	if len(machineSets) == 0 {
		// Create new MachineSet
		return r.createInitialMachineSet(ctx, deployment)
	}

	// Handle rolling update if needed
	if deployment.Spec.Strategy != nil && deployment.Spec.Strategy.Type == "RollingUpdate" {
		return r.rolloutRolling(ctx, deployment, machineSets)
	}

	// Default to recreate strategy
	return r.rolloutRecreate(ctx, deployment, machineSets)
}

// createInitialMachineSet creates the initial MachineSet for the deployment
func (r *CaptMachineDeploymentReconciler) createInitialMachineSet(ctx context.Context, deployment *infrastructurev1beta1.CaptMachineDeployment) error {
	machineSet := &infrastructurev1beta1.CaptMachineSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", deployment.Name),
			Namespace:    deployment.Namespace,
			Labels:       deployment.Spec.Template.ObjectMeta.Labels,
		},
		Spec: infrastructurev1beta1.CaptMachineSetSpec{
			Replicas: deployment.Spec.Replicas,
			Selector: deployment.Spec.Selector,
			Template: deployment.Spec.Template,
		},
	}

	if err := controllerutil.SetControllerReference(deployment, machineSet, r.Scheme); err != nil {
		return err
	}

	return r.Create(ctx, machineSet)
}

// rolloutRolling implements the rolling update strategy
func (r *CaptMachineDeploymentReconciler) rolloutRolling(ctx context.Context, deployment *infrastructurev1beta1.CaptMachineDeployment, machineSets []infrastructurev1beta1.CaptMachineSet) error {
	// TODO: Implement rolling update logic
	return nil
}

// rolloutRecreate implements the recreate strategy
func (r *CaptMachineDeploymentReconciler) rolloutRecreate(ctx context.Context, deployment *infrastructurev1beta1.CaptMachineDeployment, machineSets []infrastructurev1beta1.CaptMachineSet) error {
	// Delete old MachineSets
	for i := range machineSets {
		if err := r.Delete(ctx, &machineSets[i]); err != nil {
			return err
		}
	}

	// Create new MachineSet
	return r.createInitialMachineSet(ctx, deployment)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CaptMachineDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CaptMachineDeployment{}).
		Owns(&infrastructurev1beta1.CaptMachineSet{}).
		Complete(r)
}
