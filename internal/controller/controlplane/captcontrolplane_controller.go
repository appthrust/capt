package controlplane

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Timeouts
	controlPlaneTimeout = 30 * time.Minute
	vpcReadyTimeout     = 15 * time.Minute
	requeueInterval     = 10 * time.Second
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
		return r.setFailedStatus(ctx, controlPlane, controlplanev1beta1.ReasonFailed, fmt.Sprintf("Failed to get WorkspaceTemplate: %v", err))
	}

	// Get the CAPTCluster instance
	captCluster := &infrastructurev1beta1.CAPTCluster{}
	if err := r.Get(ctx, types.NamespacedName{Name: controlPlane.Name, Namespace: controlPlane.Namespace}, captCluster); err != nil {
		logger.Error(err, "Failed to get CAPTCluster")
		return r.setFailedStatus(ctx, controlPlane, controlplanev1beta1.ReasonFailed, fmt.Sprintf("Failed to get CAPTCluster: %v", err))
	}

	// Check VPC readiness
	vpcReady := false
	if captCluster.Status.Ready {
		vpcReady = true
	} else {
		// Check VPC conditions
		vpcCondition := meta.FindStatusCondition(captCluster.Status.Conditions, infrastructurev1beta1.VPCReadyCondition)
		if vpcCondition != nil && vpcCondition.Status == metav1.ConditionTrue {
			vpcReady = true
		}
	}

	if !vpcReady {
		// Get the last transition time for WaitingForVPC condition
		var lastTransitionTime time.Time
		if condition := meta.FindStatusCondition(controlPlane.Status.Conditions, controlplanev1beta1.ControlPlaneInitializedCondition); condition != nil {
			lastTransitionTime = condition.LastTransitionTime.Time
		} else {
			lastTransitionTime = controlPlane.CreationTimestamp.Time
		}

		// Check if we've exceeded the VPC ready timeout
		if time.Since(lastTransitionTime) > vpcReadyTimeout {
			// Update status to indicate VPC timeout
			meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
				Type:               controlplanev1beta1.ControlPlaneInitializedCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             controlplanev1beta1.ReasonVPCReadyTimeout,
				Message:            "Timed out waiting for VPC to be ready",
			})
			controlPlane.Status.Phase = "Creating"
			if err := r.Status().Update(ctx, controlPlane); err != nil {
				logger.Error(err, "Failed to update status")
				return ctrl.Result{}, err
			}
			// Requeue to continue checking VPC status
			return ctrl.Result{RequeueAfter: requeueInterval}, nil
		}

		// Update status to indicate waiting for VPC
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:               controlplanev1beta1.ControlPlaneInitializedCondition,
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             controlplanev1beta1.ReasonWaitingForVPC,
			Message:            "Waiting for VPC to be ready",
		})
		controlPlane.Status.Phase = "Creating"
		if err := r.Status().Update(ctx, controlPlane); err != nil {
			logger.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	// Create or update WorkspaceTemplateApply
	workspaceApply, err := r.reconcileWorkspaceTemplateApply(ctx, controlPlane, workspaceTemplate)
	if err != nil {
		logger.Error(err, "Failed to reconcile WorkspaceTemplateApply")
		return r.setFailedStatus(ctx, controlPlane, controlplanev1beta1.ReasonFailed, fmt.Sprintf("Failed to reconcile WorkspaceTemplateApply: %v", err))
	}

	// Check workspace conditions
	var syncedCondition, readyCondition bool
	var errorMessage string

	for _, condition := range workspaceApply.Status.Conditions {
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

	// Check if we've exceeded the overall timeout
	if !readyCondition {
		if !controlPlane.CreationTimestamp.IsZero() && time.Since(controlPlane.CreationTimestamp.Time) > controlPlaneTimeout {
			// Update status to indicate control plane timeout
			meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
				Type:               controlplanev1beta1.ControlPlaneReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             controlplanev1beta1.ReasonControlPlaneTimeout,
				Message:            "Control plane creation timed out",
			})
			controlPlane.Status.Phase = "Creating"
			if err := r.Status().Update(ctx, controlPlane); err != nil {
				logger.Error(err, "Failed to update status")
				return ctrl.Result{}, err
			}
			// Requeue to continue checking status
			return ctrl.Result{RequeueAfter: requeueInterval}, nil
		}
	}

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
			controlPlane.Status.Phase = "Creating"
		} else {
			meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
				Type:               controlplanev1beta1.ControlPlaneReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             controlplanev1beta1.ReasonCreating,
				Message:            "Control plane is being created",
			})
			controlPlane.Status.Phase = "Creating"
		}
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
	}

	if err := r.Status().Update(ctx, controlPlane); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	// Requeue to continue checking status
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
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

func (r *CAPTControlPlaneReconciler) setFailedStatus(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, reason, message string) (ctrl.Result, error) {
	meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
		Type:               controlplanev1beta1.ControlPlaneReadyCondition,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
	controlPlane.Status.Phase = "Creating"
	controlPlane.Status.Ready = false
	controlPlane.Status.Initialized = false
	if controlPlane.Status.WorkspaceTemplateStatus != nil {
		controlPlane.Status.WorkspaceTemplateStatus.Ready = false
		controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage = message
	}

	if err := r.Status().Update(ctx, controlPlane); err != nil {
		return ctrl.Result{}, err
	}
	// Requeue to continue checking status
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
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
