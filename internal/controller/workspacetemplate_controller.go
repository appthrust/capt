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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	workspaceTemplateFinalizer = "workspacetemplate.infrastructure.cluster.x-k8s.io"
)

// WorkspaceTemplateReconciler reconciles a WorkspaceTemplate object
type WorkspaceTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates/finalizers,verbs=update
//+kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
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
	if !controllerutil.ContainsFinalizer(workspaceTemplate, workspaceTemplateFinalizer) {
		controllerutil.AddFinalizer(workspaceTemplate, workspaceTemplateFinalizer)
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

	// Resolve variables
	vars, err := r.resolveVariables(ctx, workspaceTemplate)
	if err != nil {
		log.Error(err, "Failed to resolve variables")
		return ctrl.Result{}, err
	}

	// Create or update the Terraform Workspace
	workspace := &tfv1beta1.Workspace{}
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", workspaceTemplate.Name),
		Namespace: workspaceTemplate.Namespace,
	}

	if err := r.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			// Create new workspace
			if err := r.createWorkspace(ctx, workspaceTemplate, workspaceName, vars); err != nil {
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
		if err := r.updateWorkspace(ctx, workspaceTemplate, workspace, vars); err != nil {
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
			controllerutil.RemoveFinalizer(workspaceTemplate, workspaceTemplateFinalizer)
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

func (r *WorkspaceTemplateReconciler) resolveVariables(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate) ([]tfv1beta1.Var, error) {
	vars := make([]tfv1beta1.Var, 0, len(workspaceTemplate.Spec.Variables))

	for _, v := range workspaceTemplate.Spec.Variables {
		tfVar := tfv1beta1.Var{
			Key: v.Key,
		}

		if v.Value != "" {
			tfVar.Value = v.Value
		} else if v.ValueFrom != nil {
			value, err := r.resolveValueFrom(ctx, v.ValueFrom, workspaceTemplate.Namespace)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve variable %s: %w", v.Key, err)
			}
			tfVar.Value = value
		}

		vars = append(vars, tfVar)
	}

	return vars, nil
}

func (r *WorkspaceTemplateReconciler) resolveValueFrom(ctx context.Context, valueFrom *infrastructurev1beta1.VariableSource, namespace string) (string, error) {
	if valueFrom.SecretKeyRef != nil {
		secret := &corev1.Secret{}
		secretNamespace := valueFrom.SecretKeyRef.Namespace
		if secretNamespace == "" {
			secretNamespace = namespace
		}

		if err := r.Get(ctx, types.NamespacedName{
			Name:      valueFrom.SecretKeyRef.Name,
			Namespace: secretNamespace,
		}, secret); err != nil {
			return "", fmt.Errorf("failed to get secret: %w", err)
		}

		value, ok := secret.Data[valueFrom.SecretKeyRef.Key]
		if !ok {
			return "", fmt.Errorf("key %s not found in secret %s", valueFrom.SecretKeyRef.Key, valueFrom.SecretKeyRef.Name)
		}

		return string(value), nil
	}

	if valueFrom.ConfigMapKeyRef != nil {
		configMap := &corev1.ConfigMap{}
		configMapNamespace := valueFrom.ConfigMapKeyRef.Namespace
		if configMapNamespace == "" {
			configMapNamespace = namespace
		}

		if err := r.Get(ctx, types.NamespacedName{
			Name:      valueFrom.ConfigMapKeyRef.Name,
			Namespace: configMapNamespace,
		}, configMap); err != nil {
			return "", fmt.Errorf("failed to get configmap: %w", err)
		}

		value, ok := configMap.Data[valueFrom.ConfigMapKeyRef.Key]
		if !ok {
			return "", fmt.Errorf("key %s not found in configmap %s", valueFrom.ConfigMapKeyRef.Key, valueFrom.ConfigMapKeyRef.Name)
		}

		return value, nil
	}

	return "", fmt.Errorf("no valid value source specified")
}

func (r *WorkspaceTemplateReconciler) createWorkspace(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate, name types.NamespacedName, vars []tfv1beta1.Var) error {
	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: tfv1beta1.WorkspaceSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: workspaceTemplate.Spec.ProviderConfigRef,
				},
			},
			ForProvider: tfv1beta1.WorkspaceParameters{
				Module:                    workspaceTemplate.Spec.Module,
				Source:                    tfv1beta1.ModuleSource(workspaceTemplate.Spec.Source),
				EnableTerraformCLILogging: true,
				Vars:                      vars,
			},
		},
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

func (r *WorkspaceTemplateReconciler) updateWorkspace(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate, workspace *tfv1beta1.Workspace, vars []tfv1beta1.Var) error {
	workspace.Spec.ForProvider.Module = workspaceTemplate.Spec.Module
	workspace.Spec.ForProvider.Source = tfv1beta1.ModuleSource(workspaceTemplate.Spec.Source)
	workspace.Spec.ForProvider.EnableTerraformCLILogging = true
	workspace.Spec.ForProvider.Vars = vars

	// Update WriteConnectionSecretToRef if specified
	if workspaceTemplate.Spec.WriteConnectionSecretToRef != nil {
		workspace.Spec.WriteConnectionSecretToReference = workspaceTemplate.Spec.WriteConnectionSecretToRef
	}

	// Ensure ProviderConfigReference is set
	if workspace.Spec.ProviderConfigReference == nil {
		workspace.Spec.ProviderConfigReference = &xpv1.Reference{
			Name: workspaceTemplate.Spec.ProviderConfigRef,
		}
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
