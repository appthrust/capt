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
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
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
	// CaptMachineFinalizer allows CaptMachineReconciler to clean up resources associated with
	// CaptMachine before removing it from the apiserver.
	CaptMachineFinalizer = "captmachine.infrastructure.cluster.x-k8s.io"
)

// CaptMachineReconciler reconciles a CaptMachine object
type CaptMachineReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachines/finalizers,verbs=update

// Reconcile handles CaptMachine reconciliation
func (r *CaptMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CaptMachine instance
	machine := &infrastructurev1beta1.CaptMachine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !machine.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machine)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(machine, CaptMachineFinalizer) {
		controllerutil.AddFinalizer(machine, CaptMachineFinalizer)
		if err := r.Update(ctx, machine); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create or update WorkspaceTemplateApply
	if err := r.reconcileWorkspaceTemplateApply(ctx, machine); err != nil {
		logger.Error(err, "Failed to reconcile WorkspaceTemplateApply")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileWorkspaceTemplateApply creates or updates the WorkspaceTemplateApply for the machine
func (r *CaptMachineReconciler) reconcileWorkspaceTemplateApply(ctx context.Context, machine *infrastructurev1beta1.CaptMachine) error {
	apply := &infrastructurev1beta1.WorkspaceTemplateApply{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-machine", machine.Name),
			Namespace: machine.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, apply, func() error {
		apply.Spec.TemplateRef = machine.Spec.WorkspaceTemplateRef
		apply.Spec.Variables = map[string]string{
			"instance_type": machine.Spec.InstanceType,
			"node_group":    machine.Spec.NodeGroupRef.Name,
		}

		if machine.Spec.Labels != nil {
			apply.Spec.Variables["labels"] = fmt.Sprintf("%v", machine.Spec.Labels)
		}
		if machine.Spec.Tags != nil {
			apply.Spec.Variables["tags"] = fmt.Sprintf("%v", machine.Spec.Tags)
		}

		return controllerutil.SetControllerReference(machine, apply, r.Scheme)
	})

	if err != nil {
		return fmt.Errorf("failed to create or update WorkspaceTemplateApply: %w", err)
	}

	// Update machine status based on WorkspaceTemplateApply status
	return r.updateStatus(ctx, machine, apply)
}

// reconcileDelete handles CaptMachine deletion
func (r *CaptMachineReconciler) reconcileDelete(ctx context.Context, machine *infrastructurev1beta1.CaptMachine) (ctrl.Result, error) {
	// Delete the associated WorkspaceTemplateApply
	apply := &infrastructurev1beta1.WorkspaceTemplateApply{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-machine", machine.Name),
			Namespace: machine.Namespace,
		},
	}

	if err := r.Delete(ctx, apply); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(machine, CaptMachineFinalizer)
	if err := r.Update(ctx, machine); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// updateStatus updates CaptMachine status
func (r *CaptMachineReconciler) updateStatus(ctx context.Context, machine *infrastructurev1beta1.CaptMachine, apply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	// Update status based on WorkspaceTemplateApply status
	if apply.Status.Applied {
		machine.Status.Ready = true
		machine.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}

		// TODO: Get instance details from Terraform outputs
		// This will require changes to WorkspaceTemplateApply to expose outputs
	} else {
		machine.Status.Ready = false
	}

	// Update conditions
	for _, condition := range apply.Status.Conditions {
		if condition.Type == xpv1.TypeReady {
			machine.Status.Ready = condition.Status == corev1.ConditionTrue
		}
		if condition.Type == "Failed" && condition.Status == corev1.ConditionTrue {
			message := condition.Message
			reason := string(condition.Reason)
			machine.Status.FailureMessage = &message
			machine.Status.FailureReason = &reason
		}
	}

	return r.Status().Update(ctx, machine)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CaptMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CaptMachine{}).
		Owns(&infrastructurev1beta1.WorkspaceTemplateApply{}).
		Complete(r)
}
