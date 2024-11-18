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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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
	errWaitingForSecrets         = "waiting for required secrets"
	errGetWorkspace              = "cannot get Workspace"
	errWaitingForWorkspace       = "waiting for required workspace"
	errDeleteWorkspace           = "cannot delete Workspace"

	// Event reasons
	reasonCreatedWorkspace    = "CreatedWorkspace"
	reasonRetainedWorkspace   = "RetainedWorkspace"
	reasonDeletedWorkspace    = "DeletedWorkspace"
	reasonWaitingForSecrets   = "WaitingForSecrets"
	reasonWaitingForWorkspace = "WaitingForWorkspace"
	reasonWaitingForSync      = "WaitingForSync"
	reasonWaitingForReady     = "WaitingForReady"
	reasonWorkspaceReady      = "WorkspaceReady"

	// Controller name
	controllerName = "workspacetemplateapply.infrastructure.cluster.x-k8s.io"

	// Reconciliation
	requeueAfterSecret = 30 * time.Second
	requeueAfterStatus = 10 * time.Second

	// Variables
	workspaceNameVar = "${WORKSPACE_NAME}"

	// Finalizer
	workspaceTemplateApplyFinalizer = "infrastructure.cluster.x-k8s.io/finalizer"

	// Suffixes
	applySuffix = "-apply"
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

// generateWorkspaceName generates a consistent workspace name from a WorkspaceTemplateApply name
func generateWorkspaceName(applyName string) string {
	// Remove "-apply" suffix if present
	name := strings.TrimSuffix(applyName, applySuffix)
	return name
}

// waitForDependentWorkspaces checks if all dependent workspaces are ready
func (r *workspaceTemplateApplyReconciler) waitForDependentWorkspaces(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) error {
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
			r.log.Debug(errWaitingForWorkspace, "error", err)
			r.record.Event(cr, event.Normal(reasonWaitingForWorkspace,
				fmt.Sprintf("Waiting for workspace %s/%s", namespace, workspaceRef.Name)))
			return fmt.Errorf("%s: %w", errWaitingForWorkspace, err)
		}

		// Check if workspace is ready
		readyCondition := FindStatusCondition(workspace.Status.Conditions, xpv1.TypeReady)
		if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
			r.record.Event(cr, event.Normal(reasonWaitingForWorkspace,
				fmt.Sprintf("Waiting for workspace %s/%s to be ready", namespace, workspaceRef.Name)))
			return fmt.Errorf("%s: workspace %s/%s is not ready", errWaitingForWorkspace, namespace, workspaceRef.Name)
		}
	}
	return nil
}

// waitForRequiredSecrets checks if all required secrets exist
func (r *workspaceTemplateApplyReconciler) waitForRequiredSecrets(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) error {
	if len(cr.Spec.WaitForSecrets) == 0 {
		return nil
	}

	var missingSecrets []string
	for _, secretRef := range cr.Spec.WaitForSecrets {
		secret := &corev1.Secret{}
		namespace := secretRef.Namespace
		if namespace == "" {
			namespace = cr.Namespace
		}

		err := r.client.Get(ctx, types.NamespacedName{
			Name:      secretRef.Name,
			Namespace: namespace,
		}, secret)
		if err != nil {
			missingSecrets = append(missingSecrets, fmt.Sprintf("%s/%s", namespace, secretRef.Name))
		}
	}

	if len(missingSecrets) > 0 {
		message := fmt.Sprintf("Waiting for secrets: %s", strings.Join(missingSecrets, ", "))
		r.log.Debug(errWaitingForSecrets, "missing", strings.Join(missingSecrets, ", "))
		r.record.Event(cr, event.Normal(reasonWaitingForSecrets, message))
		return fmt.Errorf("%s: %s", errWaitingForSecrets, message)
	}

	return nil
}

// replaceTemplateVariables replaces template variables with their values
func replaceTemplateVariables(template *v1beta1.WorkspaceTemplate, cr *v1beta1.WorkspaceTemplateApply) (*v1beta1.WorkspaceTemplate, error) {
	// Create a deep copy of the template to avoid modifying the original
	templateCopy := template.DeepCopy()

	// Convert the template spec to JSON for easier manipulation
	specJSON, err := json.Marshal(templateCopy.Spec.Template.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template spec: %w", err)
	}

	// Generate workspace name
	workspaceName := generateWorkspaceName(cr.Name)

	// Replace workspace name variable with the generated name
	specStr := string(specJSON)
	specStr = strings.ReplaceAll(specStr, workspaceNameVar, workspaceName)

	// Replace other variables from cr.Spec.Variables
	for key, value := range cr.Spec.Variables {
		specStr = strings.ReplaceAll(specStr, fmt.Sprintf("${%s}", key), value)
	}

	// Unmarshal back to the template spec
	if err := json.Unmarshal([]byte(specStr), &templateCopy.Spec.Template.Spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template spec: %w", err)
	}

	return templateCopy, nil
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies/finalizers,verbs=update
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces;workspaces/status,verbs=get;list;watch;create;update;patch;delete
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
		return r.reconcileDelete(ctx, cr)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(cr, workspaceTemplateApplyFinalizer) {
		controllerutil.AddFinalizer(cr, workspaceTemplateApplyFinalizer)
		if err := r.client.Update(ctx, cr); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if we need to wait for secrets first
	if err := r.waitForRequiredSecrets(ctx, cr); err != nil {
		return ctrl.Result{RequeueAfter: requeueAfterSecret}, nil
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

	// Check if we need to wait for workspaces
	if len(cr.Spec.WaitForWorkspaces) > 0 {
		if err := r.waitForDependentWorkspaces(ctx, cr); err != nil {
			return ctrl.Result{RequeueAfter: requeueAfterSecret}, nil
		}
	}

	// Generate workspace name
	workspaceName := generateWorkspaceName(cr.Name)

	// Replace template variables
	template, err := replaceTemplateVariables(template, cr)
	if err != nil {
		log.Debug("Failed to replace template variables", "error", err)
		return ctrl.Result{}, err
	}

	// Create Workspace from template
	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspaceName,
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

func (r *workspaceTemplateApplyReconciler) reconcileDelete(ctx context.Context, cr *v1beta1.WorkspaceTemplateApply) (ctrl.Result, error) {
	log := r.log.WithValues("request", cr.Name)
	log.Debug("Reconciling deletion")

	// If RetainWorkspaceOnDelete is true, remove finalizer and skip workspace deletion
	if cr.Spec.RetainWorkspaceOnDelete {
		log.Info("RetainWorkspaceOnDelete is true, skipping workspace deletion",
			"workspaceName", cr.Status.WorkspaceName)
		r.record.Event(cr, event.Normal(reasonRetainedWorkspace,
			fmt.Sprintf("Retained workspace %s as specified", cr.Status.WorkspaceName)))
		controllerutil.RemoveFinalizer(cr, workspaceTemplateApplyFinalizer)
		if err := r.client.Update(ctx, cr); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Delete associated workspace if it exists
	if cr.Status.WorkspaceName != "" {
		workspace := &tfv1beta1.Workspace{}
		err := r.client.Get(ctx, types.NamespacedName{
			Name:      cr.Status.WorkspaceName,
			Namespace: cr.Namespace,
		}, workspace)

		if err == nil {
			// Check if workspace is already being deleted
			if workspace.DeletionTimestamp != nil {
				// Workspace is being deleted, wait for it
				log.Info("Waiting for workspace to be deleted", "workspaceName", workspace.Name)
				return ctrl.Result{RequeueAfter: requeueAfterStatus}, nil
			}

			// Workspace exists and not being deleted, delete it
			if err := r.client.Delete(ctx, workspace); err != nil {
				log.Debug(errDeleteWorkspace, "error", err)
				return ctrl.Result{}, fmt.Errorf("%s: %w", errDeleteWorkspace, err)
			}
			log.Info("Initiated workspace deletion", "workspaceName", workspace.Name)
			r.record.Event(cr, event.Normal(reasonDeletedWorkspace,
				fmt.Sprintf("Initiated deletion of workspace %s", cr.Status.WorkspaceName)))
			return ctrl.Result{RequeueAfter: requeueAfterStatus}, nil
		} else if !apierrors.IsNotFound(err) {
			// Error other than NotFound
			return ctrl.Result{}, err
		}
	}

	// Workspace doesn't exist or is fully deleted, remove finalizer
	controllerutil.RemoveFinalizer(cr, workspaceTemplateApplyFinalizer)
	if err := r.client.Update(ctx, cr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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

	// Copy conditions from workspace to WorkspaceTemplateApply
	cr.Status.Conditions = workspace.Status.Conditions

	// Update status
	if err := r.client.Status().Update(ctx, cr); err != nil {
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
