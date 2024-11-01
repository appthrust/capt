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
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
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
	errWaitingForSecret          = "waiting for required secret"
	errGetWorkspace              = "cannot get Workspace"
	errWaitingForWorkspace       = "waiting for required workspace"

	// Event reasons
	reasonCreatedWorkspace    = "CreatedWorkspace"
	reasonWaitingForSecret    = "WaitingForSecret"
	reasonWaitingForWorkspace = "WaitingForWorkspace"
	reasonWaitingForSync      = "WaitingForSync"
	reasonWaitingForReady     = "WaitingForReady"
	reasonWorkspaceReady      = "WorkspaceReady"

	// Controller name
	controllerName = "workspacetemplateapply.infrastructure.cluster.x-k8s.io"

	// Reconciliation
	requeueAfterSecret = 30 * time.Second
	requeueAfterStatus = 10 * time.Second
)

// WorkspaceTemplateApplyGroupKind is the group and kind of the WorkspaceTemplateApply resource
var WorkspaceTemplateApplyGroupKind = schema.GroupKind{
	Group: "infrastructure.cluster.x-k8s.io",
	Kind:  "WorkspaceTemplateApply",
}

// FindStatusCondition finds the condition that matches the given type in the condition slice.
func FindStatusCondition(conditions []xpv1.Condition, conditionType xpv1.ConditionType) *xpv1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies/finalizers,verbs=update
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups=tf.crossplane.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

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

	// If already applied, check workspace status
	if cr.Status.Applied {
		return r.reconcileWorkspaceStatus(ctx, cr)
	}

	// Check if we need to wait for a secret
	if cr.Spec.WaitForSecret != nil {
		secret := &corev1.Secret{}
		err := r.client.Get(ctx, types.NamespacedName{
			Name:      cr.Spec.WaitForSecret.Name,
			Namespace: cr.Spec.WaitForSecret.Namespace,
		}, secret)
		if err != nil {
			log.Debug(errWaitingForSecret, "error", err)
			r.record.Event(cr, event.Normal(reasonWaitingForSecret, "Waiting for secret "+cr.Spec.WaitForSecret.Name))
			return ctrl.Result{RequeueAfter: requeueAfterSecret}, nil
		}
	}

	// Check if we need to wait for workspaces
	if len(cr.Spec.WaitForWorkspaces) > 0 {
		for _, workspaceRef := range cr.Spec.WaitForWorkspaces {
			workspace := &tfv1beta1.Workspace{}
			namespace := workspaceRef.Namespace
			if namespace == "" {
				namespace = cr.Namespace
			}
			err := r.client.Get(ctx, types.NamespacedName{
				Name:      workspaceRef.Name,
				Namespace: namespace,
			}, workspace)
			if err != nil {
				log.Debug(errWaitingForWorkspace, "error", err)
				r.record.Event(cr, event.Normal(reasonWaitingForWorkspace, fmt.Sprintf("Waiting for workspace %s/%s", namespace, workspaceRef.Name)))
				return ctrl.Result{RequeueAfter: requeueAfterSecret}, nil
			}

			// Check if workspace is ready
			readyCondition := FindStatusCondition(workspace.Status.Conditions, xpv1.TypeReady)
			if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
				r.record.Event(cr, event.Normal(reasonWaitingForWorkspace, fmt.Sprintf("Waiting for workspace %s/%s to be ready", namespace, workspaceRef.Name)))
				return ctrl.Result{RequeueAfter: requeueAfterSecret}, nil
			}
		}
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
	return ctrl.Result{RequeueAfter: requeueAfterStatus}, nil
}

func (r *workspaceTemplateApplyReconciler) reconcileWorkspaceStatus(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) (ctrl.Result, error) {
	workspace := &tfv1beta1.Workspace{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      cr.Status.WorkspaceName,
		Namespace: cr.Namespace,
	}, workspace); err != nil {
		r.log.Debug(errGetWorkspace, "error", err)
		return ctrl.Result{}, err
	}

	// Check if workspace is synced
	syncedCondition := FindStatusCondition(workspace.Status.Conditions, xpv1.TypeSynced)
	if syncedCondition == nil || syncedCondition.Status != corev1.ConditionTrue {
		r.record.Event(cr, event.Normal(reasonWaitingForSync, "Waiting for workspace to be synced"))
		return ctrl.Result{RequeueAfter: requeueAfterStatus}, nil
	}

	// Check if workspace is ready
	readyCondition := FindStatusCondition(workspace.Status.Conditions, xpv1.TypeReady)
	if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
		r.record.Event(cr, event.Normal(reasonWaitingForReady, "Waiting for workspace to be ready"))
		return ctrl.Result{RequeueAfter: requeueAfterStatus}, nil
	}

	// Both synced and ready are true
	r.record.Event(cr, event.Normal(reasonWorkspaceReady, "Workspace is synced and ready"))
	return ctrl.Result{}, nil
}
