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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	workspaceTemplateFinalizerV2 = "workspacetemplate.v2.infrastructure.cluster.x-k8s.io"
)

// WorkspaceTemplateReconciler reconciles a WorkspaceTemplate object
type WorkspaceTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates/finalizers,verbs=update
//+kubebuilder:rbac:groups=tf.crossplane.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *WorkspaceTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the WorkspaceTemplate instance
	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	if err := r.Get(ctx, req.NamespacedName, workspaceTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch WorkspaceTemplate")
		return ctrl.Result{}, err
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(workspaceTemplate, workspaceTemplateFinalizerV2) {
		controllerutil.AddFinalizer(workspaceTemplate, workspaceTemplateFinalizerV2)
		if err := r.Update(ctx, workspaceTemplate); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !workspaceTemplate.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, workspaceTemplate)
	}

	// Handle normal reconciliation
	return r.reconcileNormal(ctx, workspaceTemplate)
}

func (r *WorkspaceTemplateReconciler) reconcileNormal(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Create or update the Terraform Workspace
	workspace := &tfv1beta1.Workspace{}
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", workspaceTemplate.Name),
		Namespace: workspaceTemplate.Namespace,
	}

	if err := r.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			// Create new workspace
			if err := r.createWorkspace(ctx, workspaceTemplate, workspaceName); err != nil {
				log.Error(err, "Failed to create Terraform Workspace")
				return ctrl.Result{}, err
			}
			log.Info("Created Terraform Workspace", "workspace", workspaceName)
		} else {
			log.Error(err, "Failed to get Terraform Workspace")
			return ctrl.Result{}, err
		}
	} else {
		// Update existing workspace
		if err := r.updateWorkspace(ctx, workspaceTemplate, workspace); err != nil {
			log.Error(err, "Failed to update Terraform Workspace")
			return ctrl.Result{}, err
		}
		log.Info("Updated Terraform Workspace", "workspace", workspaceName)
	}

	// Update status
	workspaceTemplate.Status.WorkspaceName = workspaceName.Name

	return ctrl.Result{}, nil
}

func (r *WorkspaceTemplateReconciler) reconcileDelete(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Delete the associated Terraform Workspace
	workspace := &tfv1beta1.Workspace{}
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", workspaceTemplate.Name),
		Namespace: workspaceTemplate.Namespace,
	}

	if err := r.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			// Workspace is already gone, remove finalizer
			controllerutil.RemoveFinalizer(workspaceTemplate, workspaceTemplateFinalizerV2)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Delete the workspace
	if err := r.Delete(ctx, workspace); err != nil {
		log.Error(err, "Failed to delete Terraform Workspace")
		return ctrl.Result{}, err
	}

	log.Info("Deleted Terraform Workspace", "workspace", workspaceName)
	return ctrl.Result{}, nil
}

func (r *WorkspaceTemplateReconciler) createWorkspace(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate, name types.NamespacedName) error {
	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: workspaceTemplate.Spec.Template.Spec,
	}

	// Set WriteConnectionSecretToRef if specified
	if workspaceTemplate.Spec.WriteConnectionSecretToRef != nil {
		workspace.Spec.WriteConnectionSecretToReference = workspaceTemplate.Spec.WriteConnectionSecretToRef
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(workspaceTemplate, workspace, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	return r.Create(ctx, workspace)
}

func (r *WorkspaceTemplateReconciler) updateWorkspace(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate, workspace *tfv1beta1.Workspace) error {
	// Update spec with template spec
	workspace.Spec = workspaceTemplate.Spec.Template.Spec

	// Update WriteConnectionSecretToRef if specified
	if workspaceTemplate.Spec.WriteConnectionSecretToRef != nil {
		workspace.Spec.WriteConnectionSecretToReference = workspaceTemplate.Spec.WriteConnectionSecretToRef
	}

	return r.Update(ctx, workspace)
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.WorkspaceTemplate{}).
		Owns(&tfv1beta1.Workspace{}).
		Complete(r)
}
