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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

// CaptMachineTemplateReconciler reconciles a CaptMachineTemplate object
type CaptMachineTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinetemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinetemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch

// Reconcile handles CaptMachineTemplate reconciliation
func (r *CaptMachineTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling CaptMachineTemplate", "name", req.Name, "namespace", req.Namespace)

	// Fetch the CaptMachineTemplate instance
	machineTemplate := &infrastructurev1beta1.CaptMachineTemplate{}
	if err := r.Get(ctx, req.NamespacedName, machineTemplate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the WorkspaceTemplate exists
	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	workspaceTemplateRef := machineTemplate.Spec.Template.Spec.WorkspaceTemplateRef
	if err := r.Get(ctx, client.ObjectKey{
		Name:      workspaceTemplateRef.Name,
		Namespace: workspaceTemplateRef.Namespace,
	}, workspaceTemplate); err != nil {
		log.Error(err, "Failed to get referenced WorkspaceTemplate",
			"name", workspaceTemplateRef.Name,
			"namespace", workspaceTemplateRef.Namespace)
		return ctrl.Result{}, fmt.Errorf("failed to get referenced WorkspaceTemplate: %w", err)
	}

	// The CaptMachineTemplate is immutable after creation
	// It serves as a template for creating CaptMachines
	// No additional reconciliation is needed

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CaptMachineTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CaptMachineTemplate{}).
		Complete(r)
}
