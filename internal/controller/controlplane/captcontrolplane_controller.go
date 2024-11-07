package controlplane

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capierrors "sigs.k8s.io/cluster-api/errors"
)

const (
	requeueInterval = 10 * time.Second
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
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces,verbs=get;list;watch

func (r *CAPTControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CAPTControlPlane")

	// Fetch the CAPTControlPlane instance
	controlPlane := &controlplanev1beta1.CAPTControlPlane{}
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		logger.Error(err, "Failed to get CAPTControlPlane")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get owner Cluster
	cluster := &clusterv1.Cluster{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: controlPlane.Namespace, Name: controlPlane.Name}, cluster); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get owner Cluster")
			return ctrl.Result{}, err
		}
		// Cluster not found, could be a standalone CAPTControlPlane
		cluster = nil
		return ctrl.Result{}, nil
	}

	// Handle deletion
	if !controlPlane.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, controlPlane)
	}

	// Set owner reference if not already set
	if err := r.setOwnerReference(ctx, controlPlane, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Handle normal reconciliation
	return r.reconcileNormal(ctx, controlPlane, cluster)
}

func (r *CAPTControlPlaneReconciler) setOwnerReference(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) error {
	if cluster == nil {
		return nil
	}

	// Check if owner reference is already set
	for _, ref := range controlPlane.OwnerReferences {
		if ref.Kind == "Cluster" && ref.APIVersion == clusterv1.GroupVersion.String() {
			return nil
		}
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(cluster, controlPlane, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %v", err)
	}

	return r.Update(ctx, controlPlane)
}

func (r *CAPTControlPlaneReconciler) reconcileNormal(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling normal state")

	// Get the referenced WorkspaceTemplate
	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	templateNamespacedName := types.NamespacedName{
		Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
		Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
	}
	if err := r.Get(ctx, templateNamespacedName, workspaceTemplate); err != nil {
		logger.Error(err, "Failed to get WorkspaceTemplate")
		return r.setFailedStatus(ctx, controlPlane, cluster, controlplanev1beta1.ReasonFailed, fmt.Sprintf("Failed to get WorkspaceTemplate: %v", err))
	}

	// Try to find existing WorkspaceTemplateApply
	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	var applyName string

	if controlPlane.Spec.WorkspaceTemplateApplyName != "" {
		// Use the name from spec if it exists
		applyName = controlPlane.Spec.WorkspaceTemplateApplyName
		logger.Info("Using existing WorkspaceTemplateApply name from spec", "name", applyName)
	} else {
		// Create a new name
		applyName = fmt.Sprintf("%s-eks-controlplane-apply", controlPlane.Name)
		logger.Info("Creating new WorkspaceTemplateApply", "name", applyName)
	}

	// Get or create WorkspaceTemplateApply
	err := r.Get(ctx, types.NamespacedName{Name: applyName, Namespace: controlPlane.Namespace}, workspaceApply)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		// Create new WorkspaceTemplateApply
		workspaceApply, err = r.createWorkspaceTemplateApply(ctx, controlPlane, workspaceTemplate, applyName)
		if err != nil {
			logger.Error(err, "Failed to create WorkspaceTemplateApply")
			return r.setFailedStatus(ctx, controlPlane, cluster, controlplanev1beta1.ReasonFailed, fmt.Sprintf("Failed to create WorkspaceTemplateApply: %v", err))
		}
	} else {
		// Update existing WorkspaceTemplateApply
		workspaceApply, err = r.updateWorkspaceTemplateApply(ctx, controlPlane, workspaceTemplate, workspaceApply)
		if err != nil {
			logger.Error(err, "Failed to update WorkspaceTemplateApply")
			return r.setFailedStatus(ctx, controlPlane, cluster, controlplanev1beta1.ReasonFailed, fmt.Sprintf("Failed to update WorkspaceTemplateApply: %v", err))
		}
	}

	// Reconcile secrets and endpoint
	if err := r.reconcileSecrets(ctx, controlPlane, cluster, workspaceApply); err != nil {
		logger.Error(err, "Failed to reconcile secrets")
		return ctrl.Result{}, err
	}

	// Update status
	return r.updateStatus(ctx, controlPlane, workspaceApply, cluster)
}

func (r *CAPTControlPlaneReconciler) createWorkspaceTemplateApply(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate,
	applyName string,
) (*infrastructurev1beta1.WorkspaceTemplateApply, error) {
	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{
		ObjectMeta: metav1.ObjectMeta{
			Name:      applyName,
			Namespace: controlPlane.Namespace,
		},
	}

	if err := ctrl.SetControllerReference(controlPlane, workspaceApply, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %w", err)
	}

	workspaceApply.Spec = r.generateWorkspaceTemplateApplySpec(controlPlane)

	if err := r.Create(ctx, workspaceApply); err != nil {
		return nil, err
	}

	return workspaceApply, nil
}

func (r *CAPTControlPlaneReconciler) updateWorkspaceTemplateApply(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	workspaceTemplate *infrastructurev1beta1.WorkspaceTemplate,
	workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply,
) (*infrastructurev1beta1.WorkspaceTemplateApply, error) {
	workspaceApply.Spec = r.generateWorkspaceTemplateApplySpec(controlPlane)

	if err := r.Update(ctx, workspaceApply); err != nil {
		return nil, err
	}

	return workspaceApply, nil
}

func (r *CAPTControlPlaneReconciler) generateWorkspaceTemplateApplySpec(controlPlane *controlplanev1beta1.CAPTControlPlane) infrastructurev1beta1.WorkspaceTemplateApplySpec {
	variables := map[string]string{
		"cluster_name":       controlPlane.Name,
		"kubernetes_version": controlPlane.Spec.Version,
	}

	if controlPlane.Spec.ControlPlaneConfig != nil {
		if controlPlane.Spec.ControlPlaneConfig.EndpointAccess != nil {
			variables["endpoint_public_access"] = fmt.Sprintf("%v", controlPlane.Spec.ControlPlaneConfig.EndpointAccess.Public)
			variables["endpoint_private_access"] = fmt.Sprintf("%v", controlPlane.Spec.ControlPlaneConfig.EndpointAccess.Private)
		}
	}

	if len(controlPlane.Spec.AdditionalTags) > 0 {
		for k, v := range controlPlane.Spec.AdditionalTags {
			variables[fmt.Sprintf("tags_%s", k)] = v
		}
	}

	return infrastructurev1beta1.WorkspaceTemplateApplySpec{
		TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
			Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
			Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
		},
		Variables: variables,
		WaitForWorkspaces: []infrastructurev1beta1.WorkspaceReference{
			{
				Name:      fmt.Sprintf("%s-vpc", controlPlane.Name),
				Namespace: controlPlane.Namespace,
			},
		},
	}
}

func (r *CAPTControlPlaneReconciler) updateStatus(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply, cluster *clusterv1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Initialize WorkspaceTemplateStatus if not exists
	if controlPlane.Status.WorkspaceTemplateStatus == nil {
		controlPlane.Status.WorkspaceTemplateStatus = &controlplanev1beta1.WorkspaceTemplateStatus{}
	}

	// Update status based on WorkspaceTemplateApply conditions
	var syncedCondition, readyCondition bool
	var errorMessage string

	for _, condition := range workspaceApply.Status.Conditions {
		logger.Info("Checking condition", "type", condition.Type, "status", condition.Status)
		if condition.Type == xpv1.TypeSynced {
			syncedCondition = condition.Status == corev1.ConditionTrue
			if !syncedCondition {
				errorMessage = condition.Message
			}
		}
		if condition.Type == xpv1.TypeReady {
			readyCondition = condition.Status == corev1.ConditionTrue
			if !readyCondition {
				errorMessage = condition.Message
			}
		}
	}

	logger.Info("Status check", "applied", workspaceApply.Status.Applied, "synced", syncedCondition, "ready", readyCondition)

	// Update status based on workspace conditions
	if !workspaceApply.Status.Applied || !syncedCondition || !readyCondition {
		if errorMessage != "" {
			meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
				Type:               controlplanev1beta1.ControlPlaneReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             controlplanev1beta1.ReasonWorkspaceError,
				Message:            errorMessage,
			})
		} else {
			meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
				Type:               controlplanev1beta1.ControlPlaneReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             controlplanev1beta1.ReasonCreating,
				Message:            "Control plane is being created",
			})
		}
		controlPlane.Status.Phase = "Creating"
		controlPlane.Status.Ready = false
		controlPlane.Status.Initialized = false
		controlPlane.Status.WorkspaceTemplateStatus.Ready = false
		if errorMessage != "" {
			controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage = errorMessage
			if workspaceApply.Status.LastAppliedTime != nil {
				controlPlane.Status.WorkspaceTemplateStatus.LastFailedRevision = workspaceApply.Status.LastAppliedTime.String()
			}
		}
	} else {
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:               controlplanev1beta1.ControlPlaneReadyCondition,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             controlplanev1beta1.ReasonReady,
			Message:            "Control plane is ready",
		})
		controlPlane.Status.Phase = "Ready"
		controlPlane.Status.Ready = true
		controlPlane.Status.Initialized = true
		controlPlane.Status.WorkspaceTemplateStatus.Ready = true
		controlPlane.Status.FailureReason = nil
		controlPlane.Status.FailureMessage = nil

		// TODO: Remove when endpoint is implemented
		// Get endpoint from Workspace
		// if workspaceApply.Status.WorkspaceName != "" {
		// 	if apiEndpoint, err := endpoint.GetEndpointFromWorkspace(ctx, r.Client, workspaceApply.Status.WorkspaceName); err != nil {
		// 		logger.Error(err, "Failed to get endpoint from Workspace")
		// 	} else if apiEndpoint != nil {
		// 		// Update CAPTControlPlane endpoint
		// 		patchBase := controlPlane.DeepCopy()
		// 		controlPlane.Spec.ControlPlaneEndpoint = *apiEndpoint

		// 		if err := r.Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		// 			logger.Error(err, "Failed to patch CAPTControlPlane endpoint")
		// 		} else {
		// 			logger.Info("Successfully patched CAPTControlPlane endpoint")

		// 			// Update owner cluster endpoint if it exists
		// 			if cluster != nil {
		// 				patchBase := cluster.DeepCopy()
		// 				cluster.Spec.ControlPlaneEndpoint = controlPlane.Spec.ControlPlaneEndpoint
		// 				if err := r.Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
		// 					logger.Error(err, "Failed to patch Cluster endpoint")
		// 				} else {
		// 					logger.Info("Successfully patched Cluster endpoint")
		// 				}
		// 			}
		// 		}
		// 	}
		// } else {
		// 	logger.Info("WorkspaceName not set in WorkspaceTemplateApply status")
		// }
	}

	// Update WorkspaceTemplateStatus fields
	if workspaceApply.Status.LastAppliedTime != nil {
		controlPlane.Status.WorkspaceTemplateStatus.LastAppliedRevision = workspaceApply.Status.LastAppliedTime.String()
	}

	// Update CAPTControlPlane status
	if err := r.Status().Update(ctx, controlPlane); err != nil {
		logger.Error(err, "Failed to update CAPTControlPlane status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully updated status")
	// Requeue to continue checking status
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *CAPTControlPlaneReconciler) setFailedStatus(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster, reason, message string) (ctrl.Result, error) {
	meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
		Type:               controlplanev1beta1.ControlPlaneReadyCondition,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
	controlPlane.Status.Phase = "Failed"
	controlPlane.Status.Ready = false
	controlPlane.Status.Initialized = false
	if controlPlane.Status.WorkspaceTemplateStatus != nil {
		controlPlane.Status.WorkspaceTemplateStatus.Ready = false
		controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage = message
	}

	// Update CAPTControlPlane status
	if err := r.Status().Update(ctx, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Update Cluster status if it exists
	if cluster != nil {
		patch := client.MergeFrom(cluster.DeepCopy())
		cluster.Status.ControlPlaneReady = false
		clusterReason := capierrors.ClusterStatusError(reason)
		cluster.Status.FailureReason = &clusterReason
		cluster.Status.FailureMessage = &message
		if err := r.Status().Patch(ctx, cluster, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Requeue to continue checking status
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *CAPTControlPlaneReconciler) reconcileDelete(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling deletion of CAPTControlPlane")

	// Find and delete associated WorkspaceTemplateApply
	var applyName string
	if controlPlane.Spec.WorkspaceTemplateApplyName != "" {
		applyName = controlPlane.Spec.WorkspaceTemplateApplyName
	} else {
		applyName = fmt.Sprintf("%s-eks-controlplane-apply", controlPlane.Name)
	}

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
		Owns(&infrastructurev1beta1.WorkspaceTemplateApply{}).
		Complete(r)
}
