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

package controlplane

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

// CAPTControlPlaneReconciler reconciles a CAPTControlPlane object
type CAPTControlPlaneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters,verbs=get;list;watch

// Reconcile handles the reconciliation of CAPTControlPlane resources
func (r *CAPTControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CAPTControlPlane instance
	controlPlane := &controlplanev1beta1.CAPTControlPlane{}
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		logger.Error(err, "Failed to get CAPTControlPlane")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !controlPlane.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, controlPlane)
	}

	// Handle normal reconciliation
	return r.reconcileNormal(ctx, controlPlane)
}

func (r *CAPTControlPlaneReconciler) reconcileNormal(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the referenced WorkspaceTemplate
	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	templateNamespacedName := types.NamespacedName{
		Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
		Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
	}
	if err := r.Get(ctx, templateNamespacedName, workspaceTemplate); err != nil {
		logger.Error(err, "Failed to get WorkspaceTemplate")
		return ctrl.Result{}, err
	}

	// Get the CAPTCluster instance
	captCluster := &infrastructurev1beta1.CAPTCluster{}
	if err := r.Get(ctx, types.NamespacedName{Name: controlPlane.Name, Namespace: controlPlane.Namespace}, captCluster); err != nil {
		logger.Error(err, "Failed to get CAPTCluster")
		return ctrl.Result{}, err
	}

	// Check if VPC is ready
	if !captCluster.Status.Ready {
		logger.Info("Waiting for VPC to be ready")
		return ctrl.Result{Requeue: true}, nil
	}

	// Create or update WorkspaceTemplateApply
	workspaceApply, err := r.reconcileWorkspaceTemplateApply(ctx, controlPlane, workspaceTemplate)
	if err != nil {
		logger.Error(err, "Failed to reconcile WorkspaceTemplateApply")
		return ctrl.Result{}, err
	}

	// Update status based on WorkspaceTemplateApply status
	if err := r.updateStatus(ctx, controlPlane, workspaceApply); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CAPTControlPlaneReconciler) reconcileWorkspaceTemplateApply(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	_ *infrastructurev1beta1.WorkspaceTemplate,
) (*infrastructurev1beta1.WorkspaceTemplateApply, error) {
	// Create WorkspaceTemplateApply name based on controlPlane name
	applyName := fmt.Sprintf("%s-apply", controlPlane.Name)

	// Prepare WorkspaceTemplateApply
	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	workspaceApply.Name = applyName
	workspaceApply.Namespace = controlPlane.Namespace

	// Set variables based on CAPTControlPlane spec
	variables := map[string]string{
		"cluster_name":       controlPlane.Name,
		"kubernetes_version": controlPlane.Spec.Version,
	}

	// Add additional configuration if specified
	if controlPlane.Spec.ControlPlaneConfig != nil {
		if controlPlane.Spec.ControlPlaneConfig.EndpointAccess != nil {
			variables["endpoint_public_access"] = fmt.Sprintf("%v", controlPlane.Spec.ControlPlaneConfig.EndpointAccess.Public)
			variables["endpoint_private_access"] = fmt.Sprintf("%v", controlPlane.Spec.ControlPlaneConfig.EndpointAccess.Private)
		}
	}

	// Add additional tags if specified
	if len(controlPlane.Spec.AdditionalTags) > 0 {
		for k, v := range controlPlane.Spec.AdditionalTags {
			variables[fmt.Sprintf("tags_%s", k)] = v
		}
	}

	// Convert WorkspaceTemplateReference
	templateRef := infrastructurev1beta1.WorkspaceTemplateReference{
		Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
		Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
	}

	// Set template reference and variables
	workspaceApply.Spec.TemplateRef = templateRef
	workspaceApply.Spec.Variables = variables

	// Set wait for VPC workspace
	workspaceApply.Spec.WaitForWorkspaces = []infrastructurev1beta1.WorkspaceReference{
		{
			Name:      fmt.Sprintf("%s-vpc", controlPlane.Name),
			Namespace: controlPlane.Namespace,
		},
	}

	// Create or update the WorkspaceTemplateApply
	err := r.Get(ctx, types.NamespacedName{Name: applyName, Namespace: controlPlane.Namespace}, workspaceApply)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		// Create new WorkspaceTemplateApply
		if err := r.Create(ctx, workspaceApply); err != nil {
			return nil, err
		}
	} else {
		// Update existing WorkspaceTemplateApply
		if err := r.Update(ctx, workspaceApply); err != nil {
			return nil, err
		}
	}

	return workspaceApply, nil
}

func (r *CAPTControlPlaneReconciler) updateStatus(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	// Update status based on WorkspaceTemplateApply status
	if workspaceApply.Status.Applied {
		controlPlane.Status.Ready = true
		controlPlane.Status.Initialized = true
		if controlPlane.Status.WorkspaceTemplateStatus == nil {
			controlPlane.Status.WorkspaceTemplateStatus = &controlplanev1beta1.WorkspaceTemplateStatus{}
		}
		controlPlane.Status.WorkspaceTemplateStatus.Ready = true
		controlPlane.Status.WorkspaceTemplateStatus.State = workspaceApply.Status.WorkspaceName
		if workspaceApply.Status.LastAppliedTime != nil {
			controlPlane.Status.WorkspaceTemplateStatus.LastAppliedRevision = workspaceApply.Status.LastAppliedTime.String()
		}
	}

	return r.Status().Update(ctx, controlPlane)
}

func (r *CAPTControlPlaneReconciler) reconcileDelete(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling deletion of CAPTControlPlane")

	// Find and delete associated WorkspaceTemplateApply
	applyName := fmt.Sprintf("%s-apply", controlPlane.Name)
	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      applyName,
		Namespace: controlPlane.Namespace,
	}, workspaceApply)

	if err == nil {
		// WorkspaceTemplateApply exists, delete it
		if err := r.Delete(ctx, workspaceApply); err != nil {
			logger.Error(err, "Failed to delete WorkspaceTemplateApply")
			return ctrl.Result{}, err
		}
		logger.Info("Successfully deleted WorkspaceTemplateApply")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CAPTControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1beta1.CAPTControlPlane{}).
		Complete(r)
}
