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
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	machineFinalizerName = "captmachine.infrastructure.cluster.x-k8s.io"
)

// CaptMachineReconciler reconciles a CaptMachine object
type CaptMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachines/finalizers,verbs=update
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles the reconciliation of CaptMachine resources
func (r *CaptMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CaptMachine instance
	machine := &infrastructurev1beta1.CaptMachine{}
	if err := r.Get(ctx, req.NamespacedName, machine); err != nil {
		logger.Error(err, "unable to fetch CaptMachine")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(machine, machineFinalizerName) {
		controllerutil.AddFinalizer(machine, machineFinalizerName)
		if err := r.Update(ctx, machine); err != nil {
			logger.Error(err, "failed to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !machine.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machine)
	}

	// Handle normal reconciliation
	return r.reconcileNormal(ctx, machine)
}

func (r *CaptMachineReconciler) reconcileNormal(ctx context.Context, machine *infrastructurev1beta1.CaptMachine) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Create or update WorkspaceTemplateApply
	apply := &infrastructurev1beta1.WorkspaceTemplateApply{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-nodegroup", machine.Name),
			Namespace: machine.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, apply, func() error {
		apply.Spec.TemplateRef = infrastructurev1beta1.WorkspaceTemplateReference{
			Name:      machine.Spec.WorkspaceTemplateRef.Name,
			Namespace: machine.Spec.WorkspaceTemplateRef.Namespace,
		}

		// Set variables from machine config
		apply.Spec.Variables = map[string]string{
			"node_group_name": machine.Spec.NodeGroupConfig.Name,
			"instance_type":   machine.Spec.NodeGroupConfig.InstanceType,
			"desired_size":    strconv.Itoa(int(machine.Spec.NodeGroupConfig.Scaling.DesiredSize)),
			"min_size":        strconv.Itoa(int(machine.Spec.NodeGroupConfig.Scaling.MinSize)),
			"max_size":        strconv.Itoa(int(machine.Spec.NodeGroupConfig.Scaling.MaxSize)),
		}

		// Set owner reference
		return controllerutil.SetControllerReference(machine, apply, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "failed to create/update WorkspaceTemplateApply")
		return ctrl.Result{}, err
	}

	// Update machine status
	if err := r.updateMachineStatus(ctx, machine, apply); err != nil {
		logger.Error(err, "failed to update machine status")
		return ctrl.Result{}, err
	}

	// Requeue for periodic status updates
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *CaptMachineReconciler) reconcileDelete(ctx context.Context, machine *infrastructurev1beta1.CaptMachine) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the WorkspaceTemplateApply
	apply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	applyName := fmt.Sprintf("%s-nodegroup", machine.Name)
	err := r.Get(ctx, types.NamespacedName{
		Name:      applyName,
		Namespace: machine.Namespace,
	}, apply)
	if client.IgnoreNotFound(err) != nil {
		logger.Error(err, "failed to get WorkspaceTemplateApply")
		return ctrl.Result{}, err
	}

	// Delete the WorkspaceTemplateApply if it exists
	if err == nil {
		if err := r.Delete(ctx, apply); err != nil {
			logger.Error(err, "failed to delete WorkspaceTemplateApply")
			return ctrl.Result{}, err
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(machine, machineFinalizerName)
	if err := r.Update(ctx, machine); err != nil {
		logger.Error(err, "failed to remove finalizer")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CaptMachineReconciler) updateMachineStatus(ctx context.Context, machine *infrastructurev1beta1.CaptMachine, apply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	// Update status based on WorkspaceTemplateApply status
	machine.Status.Ready = apply.Status.Applied

	// Update current size from the desired size in variables
	if sizeStr, ok := apply.Spec.Variables["desired_size"]; ok {
		if size, err := strconv.ParseInt(sizeStr, 10, 32); err == nil {
			currentSize := int32(size)
			machine.Status.CurrentSize = &currentSize
		}
	}

	// Update last scaling time if scaling operation was performed
	if apply.Status.LastAppliedTime != nil {
		machine.Status.LastScalingTime = apply.Status.LastAppliedTime
	}

	// Update conditions
	setMachineConditions(machine, apply)

	return r.Status().Update(ctx, machine)
}

func setMachineConditions(machine *infrastructurev1beta1.CaptMachine, apply *infrastructurev1beta1.WorkspaceTemplateApply) {
	// Set Ready condition
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "NodeGroupReady",
		Message:            "Node group is ready",
	}

	if !apply.Status.Applied {
		readyCondition.Status = metav1.ConditionFalse
		readyCondition.Reason = "NodeGroupNotReady"
		readyCondition.Message = "Node group is not ready"
	}

	// Update conditions
	conditions := []metav1.Condition{readyCondition}
	machine.Status.Conditions = conditions
}

// SetupWithManager sets up the controller with the Manager.
func (r *CaptMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CaptMachine{}).
		Owns(&infrastructurev1beta1.WorkspaceTemplateApply{}).
		Complete(r)
}
