package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/controller/controlplane/endpoint"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util/conditions"
)

const (
	// requeueInterval is the interval to requeue when waiting for resources
	requeueInterval = 10 * time.Second

	// ControlPlaneInitializedCondition represents the condition type for control plane initialization
	ControlPlaneInitializedCondition clusterv1.ConditionType = "ControlPlaneInitialized"

	// InfrastructureReadyCondition represents the condition type for infrastructure readiness
	InfrastructureReadyCondition clusterv1.ConditionType = "InfrastructureReady"
)

// CAPTClusterReconciler reconciles a CAPTCluster object
type CAPTClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch;update;patch

func (r *CAPTClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CAPTCluster")

	// Fetch the CAPTCluster instance
	captCluster := &infrastructurev1beta1.CAPTCluster{}
	if err := r.Get(ctx, req.NamespacedName, captCluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("CAPTCluster resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get CAPTCluster")
		return ctrl.Result{}, err
	}

	// Get owner Cluster
	cluster := &clusterv1.Cluster{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: captCluster.Namespace, Name: captCluster.Name}, cluster); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get owner Cluster")
			return ctrl.Result{}, err
		}
		// Cluster not found, could be a standalone CAPTCluster
		cluster = nil
		return ctrl.Result{}, nil
	}

	// Handle deletion
	if !captCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, captCluster)
	}

	// Set owner reference if not already set
	if err := r.setOwnerReference(ctx, captCluster, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Validate VPC configuration
	if err := captCluster.Spec.ValidateVPCConfiguration(); err != nil {
		logger.Error(err, "Invalid VPC configuration")
		return r.setFailedStatus(ctx, captCluster, cluster, "InvalidVPCConfig", err.Error())
	}

	// Handle finalizer
	if err := handleFinalizer(ctx, r.Client, captCluster); err != nil {
		return ctrl.Result{}, err
	}

	// Handle VPC configuration
	result, err := r.reconcileVPC(ctx, captCluster, cluster)
	if err != nil {
		logger.Error(err, "Failed to reconcile VPC")
		return r.setFailedStatus(ctx, captCluster, cluster, "VPCReconciliationFailed", err.Error())
	}

	logger.Info("Successfully reconciled CAPTCluster", "name", captCluster.Name)
	return result, nil
}

func (r *CAPTClusterReconciler) setOwnerReference(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
	if cluster == nil {
		return nil
	}

	// Check if owner reference is already set
	for _, ref := range captCluster.OwnerReferences {
		if ref.Kind == "Cluster" && ref.APIVersion == clusterv1.GroupVersion.String() {
			return nil
		}
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(cluster, captCluster, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %v", err)
	}

	return r.Update(ctx, captCluster)
}

func (r *CAPTClusterReconciler) updateStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
	logger := log.FromContext(ctx)
	logger.Info("Updating status", "captCluster.Status.Ready", captCluster.Status.Ready)

	// Update CAPTCluster status
	if err := r.Status().Update(ctx, captCluster); err != nil {
		logger.Error(err, "Failed to update CAPTCluster status")
		return fmt.Errorf("failed to update CAPTCluster status: %v", err)
	}

	// Update Cluster status if it exists
	if cluster != nil {
		logger.Info("Updating cluster status",
			"InfrastructureReady", cluster.Status.InfrastructureReady,
			"ControlPlaneReady", cluster.Status.ControlPlaneReady,
			"CurrentPhase", cluster.Status.Phase)

		patch := client.MergeFrom(cluster.DeepCopy())

		// Update infrastructure ready status
		cluster.Status.InfrastructureReady = captCluster.Status.Ready
		logger.Info("Set InfrastructureReady", "value", cluster.Status.InfrastructureReady)

		// Update failure reason and message if present
		if captCluster.Status.FailureReason != nil {
			reason := capierrors.ClusterStatusError(*captCluster.Status.FailureReason)
			cluster.Status.FailureReason = &reason
			logger.Info("Updated failure reason", "reason", *captCluster.Status.FailureReason)
		}
		if captCluster.Status.FailureMessage != nil {
			cluster.Status.FailureMessage = captCluster.Status.FailureMessage
			logger.Info("Updated failure message", "message", *captCluster.Status.FailureMessage)
		}

		// Update failure domains if present
		if len(captCluster.Status.FailureDomains) > 0 {
			cluster.Status.FailureDomains = captCluster.Status.FailureDomains
			logger.Info("Updated failure domains", "count", len(captCluster.Status.FailureDomains))
		}

		// Update conditions when infrastructure is ready
		if captCluster.Status.Ready {
			// Set ControlPlaneInitialized condition
			conditions.Set(cluster, &clusterv1.Condition{
				Type:               ControlPlaneInitializedCondition,
				Status:             corev1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             "ControlPlaneInitialized",
				Message:            "Control plane has been initialized",
			})
			logger.Info("Set ControlPlaneInitialized condition to True")

			// Set InfrastructureReady condition
			conditions.Set(cluster, &clusterv1.Condition{
				Type:               InfrastructureReadyCondition,
				Status:             corev1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             "InfrastructureReady",
				Message:            "Infrastructure is ready",
			})
			logger.Info("Set InfrastructureReady condition to True")
		}

		// Note: We don't update the Phase here as it's managed by the Control Plane controller
		logger.Info("Current cluster phase", "phase", cluster.Status.Phase)

		if err := r.Status().Patch(ctx, cluster, patch); err != nil {
			logger.Error(err, "Failed to patch cluster status")
			return fmt.Errorf("failed to update Cluster status: %v", err)
		}
		logger.Info("Successfully patched cluster status")
	} else {
		logger.Info("Cluster is nil, skipping cluster status update")
	}

	return nil
}

func (r *CAPTClusterReconciler) setFailedStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster, reason, message string) (ctrl.Result, error) {
	meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
		Type:               infrastructurev1beta1.VPCFailedCondition,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
	captCluster.Status.Ready = false
	captCluster.Status.FailureReason = &reason
	captCluster.Status.FailureMessage = &message

	if captCluster.Status.WorkspaceTemplateStatus != nil {
		captCluster.Status.WorkspaceTemplateStatus.Ready = false
		captCluster.Status.WorkspaceTemplateStatus.LastFailureMessage = message
	}

	if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, fmt.Errorf(message)
}

func (r *CAPTClusterReconciler) reconcileDelete(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling deletion of CAPTCluster")

	// Check if VPC should be retained
	if captCluster.Spec.RetainVPCOnDelete && captCluster.Spec.VPCTemplateRef != nil {
		logger.Info("RetainVPCOnDelete is true, skipping VPC deletion",
			"vpcId", captCluster.Status.VPCID,
			"workspaceTemplateApplyName", captCluster.Spec.WorkspaceTemplateApplyName)
		return ctrl.Result{}, nil
	}

	// Find and delete associated WorkspaceTemplateApply
	if captCluster.Spec.WorkspaceTemplateApplyName != "" {
		workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      captCluster.Spec.WorkspaceTemplateApplyName,
			Namespace: captCluster.Namespace,
		}, workspaceApply)

		if err == nil {
			// WorkspaceTemplateApply exists, delete it
			if err := r.Delete(ctx, workspaceApply); err != nil {
				logger.Error(err, "Failed to delete WorkspaceTemplateApply")
				return ctrl.Result{}, err
			}
			logger.Info("Successfully deleted WorkspaceTemplateApply")
		}
	}

	return ctrl.Result{}, nil
}

func (r *CAPTClusterReconciler) reconcileVPC(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting VPC reconciliation", "clusterName", captCluster.Name)

	if captCluster.Spec.ExistingVPCID != "" {
		logger.Info("Using existing VPC", "vpcId", captCluster.Spec.ExistingVPCID)
		// Use existing VPC
		captCluster.Status.VPCID = captCluster.Spec.ExistingVPCID
		captCluster.Status.Ready = true
		meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
			Type:               infrastructurev1beta1.VPCReadyCondition,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             infrastructurev1beta1.ReasonExistingVPCUsed,
			Message:            "Using existing VPC",
		})

		if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Create new VPC using template
	if captCluster.Spec.VPCTemplateRef != nil {
		logger.Info("Creating new VPC using template", "templateRef", captCluster.Spec.VPCTemplateRef)

		// Get the referenced WorkspaceTemplate
		vpcTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
		templateName := types.NamespacedName{
			Name:      captCluster.Spec.VPCTemplateRef.Name,
			Namespace: captCluster.Namespace,
		}
		if err := r.Get(ctx, templateName, vpcTemplate); err != nil {
			logger.Error(err, "Failed to get VPC WorkspaceTemplate")
			return ctrl.Result{}, fmt.Errorf("failed to get VPC WorkspaceTemplate: %v", err)
		}

		// Try to find existing WorkspaceTemplateApply
		workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
		var applyName string

		if captCluster.Spec.WorkspaceTemplateApplyName != "" {
			// Use the name from spec if it exists
			applyName = captCluster.Spec.WorkspaceTemplateApplyName
			logger.Info("Using existing WorkspaceTemplateApply name from spec", "name", applyName)
		} else {
			// Create a new name
			applyName = fmt.Sprintf("%s-vpc", captCluster.Name)
			logger.Info("Creating new WorkspaceTemplateApply", "name", applyName)
		}

		// Get or create WorkspaceTemplateApply
		err := r.Get(ctx, types.NamespacedName{Name: applyName, Namespace: captCluster.Namespace}, workspaceApply)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
			// Create new WorkspaceTemplateApply
			workspaceApply = &infrastructurev1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Name:      applyName,
					Namespace: captCluster.Namespace,
				},
				Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
					TemplateRef: *captCluster.Spec.VPCTemplateRef,
					Variables: map[string]string{
						"name":        fmt.Sprintf("%s-vpc", captCluster.Name),
						"environment": "production", // TODO: Make this configurable
					},
				},
			}

			// Set owner reference
			if err := controllerutil.SetControllerReference(captCluster, workspaceApply, r.Scheme); err != nil {
				logger.Error(err, "Failed to set owner reference")
				return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %v", err)
			}

			if err := r.Create(ctx, workspaceApply); err != nil {
				logger.Error(err, "Failed to create WorkspaceTemplateApply")
				return ctrl.Result{}, fmt.Errorf("failed to create WorkspaceTemplateApply: %v", err)
			}

			// Update WorkspaceTemplateApplyName in Spec
			patch := client.MergeFrom(captCluster.DeepCopy())
			captCluster.Spec.WorkspaceTemplateApplyName = applyName
			if err := r.Patch(ctx, captCluster, patch); err != nil {
				logger.Error(err, "Failed to update WorkspaceTemplateApplyName in spec")
				return ctrl.Result{}, err
			}
		} else {
			// Update existing WorkspaceTemplateApply if needed
			workspaceApply.Spec = infrastructurev1beta1.WorkspaceTemplateApplySpec{
				TemplateRef: *captCluster.Spec.VPCTemplateRef,
				Variables: map[string]string{
					"name":        fmt.Sprintf("%s-vpc", captCluster.Name),
					"environment": "production", // TODO: Make this configurable
				},
			}
			if err := r.Update(ctx, workspaceApply); err != nil {
				logger.Error(err, "Failed to update WorkspaceTemplateApply")
				return ctrl.Result{}, fmt.Errorf("failed to update WorkspaceTemplateApply: %v", err)
			}
		}

		// Initialize WorkspaceTemplateStatus if not exists
		if captCluster.Status.WorkspaceTemplateStatus == nil {
			captCluster.Status.WorkspaceTemplateStatus = &infrastructurev1beta1.CAPTClusterWorkspaceStatus{}
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
				meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
					Type:               infrastructurev1beta1.VPCReadyCondition,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             infrastructurev1beta1.ReasonVPCCreationFailed,
					Message:            errorMessage,
				})
			} else {
				meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
					Type:               infrastructurev1beta1.VPCReadyCondition,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             infrastructurev1beta1.ReasonVPCCreating,
					Message:            "VPC is being created",
				})
			}
			captCluster.Status.Ready = false
			captCluster.Status.WorkspaceTemplateStatus.Ready = false
			if errorMessage != "" {
				captCluster.Status.WorkspaceTemplateStatus.LastFailureMessage = errorMessage
				if workspaceApply.Status.LastAppliedTime != nil {
					captCluster.Status.WorkspaceTemplateStatus.LastFailedRevision = workspaceApply.Status.LastAppliedTime.String()
				}
			}

			if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: requeueInterval}, nil
		}

		// Get VPC ID from workspace outputs
		if workspaceApply.Status.WorkspaceName != "" {
			logger.Info("Attempting to get VPC ID from workspace",
				"workspaceApplyName", workspaceApply.Name,
				"workspaceName", workspaceApply.Status.WorkspaceName)

			vpcID, err := endpoint.GetVPCIDFromWorkspace(ctx, r.Client, workspaceApply.Status.WorkspaceName)
			if err != nil {
				logger.Error(err, "Failed to get VPC ID from workspace")
				return ctrl.Result{RequeueAfter: requeueInterval}, nil
			}

			logger.Info("Retrieved VPC ID from workspace", "vpcId", vpcID)

			if vpcID != "" {
				captCluster.Status.VPCID = vpcID
				logger.Info("Set VPC ID in status", "vpcId", vpcID)
			} else {
				logger.Info("VPC ID not found in workspace outputs, requeueing")
				return ctrl.Result{RequeueAfter: requeueInterval}, nil
			}
		} else {
			logger.Info("WorkspaceName not set in WorkspaceTemplateApply status", "workspaceApplyName", workspaceApply.Name)
			return ctrl.Result{RequeueAfter: requeueInterval}, nil
		}

		// VPC is ready
		meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
			Type:               infrastructurev1beta1.VPCReadyCondition,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             infrastructurev1beta1.ReasonVPCCreated,
			Message:            "VPC has been successfully created",
		})

		captCluster.Status.Ready = true
		captCluster.Status.WorkspaceTemplateStatus.Ready = true
		captCluster.Status.WorkspaceTemplateStatus.LastAppliedRevision = workspaceApply.Status.LastAppliedTime.String()

		logger.Info("Updating final status",
			"ready", captCluster.Status.Ready,
			"vpcId", captCluster.Status.VPCID,
			"workspaceTemplateStatus", captCluster.Status.WorkspaceTemplateStatus)

		if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CAPTClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CAPTCluster{}).
		Owns(&infrastructurev1beta1.WorkspaceTemplateApply{}).
		Complete(r)
}
