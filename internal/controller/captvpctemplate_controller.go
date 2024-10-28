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
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CAPTVPCTemplateReconciler reconciles a CAPTVPCTemplate object
type CAPTVPCTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.capt.labthrust.io,resources=captvpctemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.capt.labthrust.io,resources=captvpctemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.capt.labthrust.io,resources=captvpctemplates/finalizers,verbs=update
//+kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CAPTVPCTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CAPTVPCTemplate instance
	vpcTemplate := &infrastructurev1beta1.CAPTVPCTemplate{}
	if err := r.Get(ctx, req.NamespacedName, vpcTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch CAPTVPCTemplate")
		return ctrl.Result{}, err
	}

	// Generate workspace name
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-vpc-workspace", vpcTemplate.Name),
		Namespace: vpcTemplate.Namespace,
	}

	// Check if workspace exists
	workspace := &tfv1beta1.Workspace{}
	if err := r.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			// Create new workspace
			if err := r.createWorkspace(ctx, vpcTemplate, workspaceName); err != nil {
				log.Error(err, "Failed to create workspace")
				return ctrl.Result{}, err
			}
			log.Info("Created new workspace", "workspace", workspaceName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get workspace")
		return ctrl.Result{}, err
	}

	// Update existing workspace
	if err := r.updateWorkspace(ctx, vpcTemplate, workspace); err != nil {
		log.Error(err, "Failed to update workspace")
		return ctrl.Result{}, err
	}
	log.Info("Updated workspace", "workspace", workspaceName)

	// Update status
	vpcTemplate.Status.WorkspaceName = workspaceName.Name
	if err := r.Status().Update(ctx, vpcTemplate); err != nil {
		log.Error(err, "Failed to update CAPTVPCTemplate status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CAPTVPCTemplateReconciler) createWorkspace(ctx context.Context, vpcTemplate *infrastructurev1beta1.CAPTVPCTemplate, name types.NamespacedName) error {
	hclCode, err := generateVPCWorkspaceModule(vpcTemplate)
	if err != nil {
		return fmt.Errorf("failed to generate VPC Terraform code: %w", err)
	}

	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: tfv1beta1.WorkspaceSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: "aws-provider-for-eks",
				},
			},
			ForProvider: tfv1beta1.WorkspaceParameters{
				Module:                    hclCode,
				Source:                    tfv1beta1.ModuleSourceInline,
				EnableTerraformCLILogging: true,
				Vars: []tfv1beta1.Var{
					{
						Key:   "name",
						Value: vpcTemplate.Name,
					},
				},
			},
		},
	}

	// Set WriteConnectionSecretToRef if specified
	if vpcTemplate.Spec.WriteConnectionSecretToRef != nil {
		workspace.Spec.WriteConnectionSecretToReference = vpcTemplate.Spec.WriteConnectionSecretToRef
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(vpcTemplate, workspace, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	return r.Create(ctx, workspace)
}

func (r *CAPTVPCTemplateReconciler) updateWorkspace(ctx context.Context, vpcTemplate *infrastructurev1beta1.CAPTVPCTemplate, workspace *tfv1beta1.Workspace) error {
	hclCode, err := generateVPCWorkspaceModule(vpcTemplate)
	if err != nil {
		return fmt.Errorf("failed to generate VPC Terraform code: %w", err)
	}

	workspace.Spec.ForProvider.Module = hclCode
	workspace.Spec.ForProvider.Vars = []tfv1beta1.Var{
		{
			Key:   "name",
			Value: vpcTemplate.Name,
		},
	}

	// Update WriteConnectionSecretToRef if specified
	if vpcTemplate.Spec.WriteConnectionSecretToRef != nil {
		workspace.Spec.WriteConnectionSecretToReference = vpcTemplate.Spec.WriteConnectionSecretToRef
	}

	// Ensure ProviderConfigReference is set
	if workspace.Spec.ProviderConfigReference == nil {
		workspace.Spec.ProviderConfigReference = &xpv1.Reference{
			Name: "aws-provider-for-eks",
		}
	}

	return r.Update(ctx, workspace)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CAPTVPCTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CAPTVPCTemplate{}).
		Complete(r)
}
