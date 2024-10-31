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

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/appthrust/capt/api/v1beta1"
)

const (
	// Errors
	errNotWorkspaceTemplateApply = "managed resource is not a WorkspaceTemplateApply custom resource"
	errTrackPCUsage              = "cannot track ProviderConfig usage"
	errGetPC                     = "cannot get ProviderConfig"
	errGetCreds                  = "cannot get credentials"
	errGetTemplate               = "cannot get WorkspaceTemplate"
	errCreateWorkspace           = "cannot create Workspace"

	// Event reasons
	reasonCreatedWorkspace = "CreatedWorkspace"

	// Controller name
	controllerName = "workspacetemplateapply.infrastructure.appthrust.dev"
)

// WorkspaceTemplateApplyGroupKind is the group and kind of the WorkspaceTemplateApply resource
var WorkspaceTemplateApplyGroupKind = schema.GroupKind{
	Group: "infrastructure.appthrust.dev",
	Kind:  "WorkspaceTemplateApply",
}

// SetupWorkspaceTemplateApply adds a controller that reconciles WorkspaceTemplateApplies.
func SetupWorkspaceTemplateApply(mgr ctrl.Manager, l logging.Logger) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&v1beta1.WorkspaceTemplateApply{}).
		Complete(&workspaceTemplateApplyReconciler{
			client: mgr.GetClient(),
			log:    l,
			record: event.NewAPIRecorder(mgr.GetEventRecorderFor(controllerName)),
		})
}

type workspaceTemplateApplyReconciler struct {
	client client.Client
	log    logging.Logger
	record event.Recorder
}

func (r *workspaceTemplateApplyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	cr := &v1beta1.WorkspaceTemplateApply{}
	if err := r.client.Get(ctx, req.NamespacedName, cr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if meta.WasDeleted(cr) {
		return ctrl.Result{}, nil
	}

	// Get the referenced WorkspaceTemplate
	template := &v1beta1.WorkspaceTemplate{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      cr.Spec.TemplateRef.Name,
		Namespace: cr.Spec.TemplateRef.Namespace,
	}, template); err != nil {
		log.Debug(errGetTemplate, "error", err)
		return ctrl.Result{}, err
	}

	// If already applied, skip
	if cr.Status.Applied {
		return ctrl.Result{}, nil
	}

	// Create Workspace from template
	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-workspace",
			Namespace: cr.Namespace,
		},
		Spec: template.Spec.Template.Spec,
	}

	// Set connection secret if specified
	if cr.Spec.WriteConnectionSecretToRef != nil {
		workspace.Spec.WriteConnectionSecretToReference = cr.Spec.WriteConnectionSecretToRef
	}

	if err := r.client.Create(ctx, workspace); err != nil {
		log.Debug(errCreateWorkspace, "error", err)
		return ctrl.Result{}, err
	}

	// Update status
	cr.Status.WorkspaceName = workspace.GetName()
	cr.Status.Applied = true
	now := metav1.Now()
	cr.Status.LastAppliedTime = &now

	if err := r.client.Status().Update(ctx, cr); err != nil {
		return ctrl.Result{}, err
	}

	r.record.Event(cr, event.Normal(reasonCreatedWorkspace, "Created Workspace from template"))
	return ctrl.Result{}, nil
}
