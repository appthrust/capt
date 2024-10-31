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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

func (r *WorkspaceTemplateReconciler) reconcileNormal(_ context.Context, _ *infrastructurev1beta1.WorkspaceTemplate) (ctrl.Result, error) {
	// Template validation logic can be added here
	// For now, we just ensure the template exists and is valid
	return ctrl.Result{}, nil
}

func (r *WorkspaceTemplateReconciler) reconcileDelete(ctx context.Context, workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate) (ctrl.Result, error) {
	// Remove finalizer
	controllerutil.RemoveFinalizer(workspaceTemplate, workspaceTemplateFinalizerV2)
	if err := r.Update(ctx, workspaceTemplate); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.WorkspaceTemplate{}).
		Complete(r)
}
