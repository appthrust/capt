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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// requeueInterval is the interval to requeue when waiting for resources
	requeueInterval = 10 * time.Second
)

// TODO: Add timeout handling for:
// - VPC creation (overall process)
// - Secret availability after workspace is ready
// - Workspace readiness check

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

func (r *CAPTClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CAPTCluster instance
	captCluster := &infrastructurev1beta1.CAPTCluster{}
	if err := r.Get(ctx, req.NamespacedName, captCluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("CAPTCluster resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CAPTCluster")
		return ctrl.Result{}, err
	}

	// Validate VPC configuration
	if err := captCluster.Spec.ValidateVPCConfiguration(); err != nil {
		log.Error(err, "Invalid VPC configuration")
		return ctrl.Result{}, err
	}

	// Handle finalizer
	if err := handleFinalizer(ctx, r.Client, captCluster); err != nil {
		return ctrl.Result{}, err
	}

	// Handle VPC configuration
	result, err := r.reconcileVPC(ctx, captCluster)
	if err != nil {
		log.Error(err, "Failed to reconcile VPC")
		return result, err
	}

	log.Info("Successfully reconciled CAPTCluster", "name", captCluster.Name)
	return result, nil
}

func (r *CAPTClusterReconciler) reconcileVPC(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (ctrl.Result, error) {
	if captCluster.Spec.ExistingVPCID != "" {
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
		return ctrl.Result{}, r.Status().Update(ctx, captCluster)
	}

	// Create new VPC using template
	if captCluster.Spec.VPCTemplateRef != nil {
		// Get the referenced WorkspaceTemplate
		vpcTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
		templateName := types.NamespacedName{
			Name:      captCluster.Spec.VPCTemplateRef.Name,
			Namespace: captCluster.Namespace,
		}
		if err := r.Get(ctx, templateName, vpcTemplate); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get VPC WorkspaceTemplate: %v", err)
		}

		// Create or update WorkspaceTemplateApply for VPC
		vpcApply := &infrastructurev1beta1.WorkspaceTemplateApply{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-vpc", captCluster.Name),
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
		if err := controllerutil.SetControllerReference(captCluster, vpcApply, r.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %v", err)
		}

		if err := r.Create(ctx, vpcApply); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return ctrl.Result{}, fmt.Errorf("failed to create VPC WorkspaceTemplateApply: %v", err)
			}
			// Update existing WorkspaceTemplateApply
			existing := &infrastructurev1beta1.WorkspaceTemplateApply{}
			if err := r.Get(ctx, types.NamespacedName{Name: vpcApply.Name, Namespace: vpcApply.Namespace}, existing); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to get existing VPC WorkspaceTemplateApply: %v", err)
			}
			existing.Spec = vpcApply.Spec
			if err := r.Update(ctx, existing); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to update VPC WorkspaceTemplateApply: %v", err)
			}
			vpcApply = existing
		}

		// Update status with VPC workspace name
		captCluster.Status.VPCWorkspaceName = vpcApply.Name

		// Check if WorkspaceTemplateApply is ready
		if !vpcApply.Status.Applied {
			// Initial apply not completed yet
			meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
				Type:               infrastructurev1beta1.VPCReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             infrastructurev1beta1.ReasonVPCCreating,
				Message:            "Initializing VPC creation",
			})
			captCluster.Status.Ready = false
			if err := r.Status().Update(ctx, captCluster); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to update CAPTCluster status: %v", err)
			}
			return ctrl.Result{RequeueAfter: requeueInterval}, nil
		}

		// Check workspace conditions
		syncedCondition := false
		readyCondition := false
		var message string

		for _, condition := range vpcApply.Status.Conditions {
			switch condition.Type {
			case xpv1.TypeSynced:
				syncedCondition = condition.Status == corev1.ConditionTrue
				if !syncedCondition {
					message = condition.Message
				}
			case xpv1.TypeReady:
				readyCondition = condition.Status == corev1.ConditionTrue
				if !readyCondition {
					message = condition.Message
				}
			}
		}

		if !syncedCondition || !readyCondition {
			// Workspace not ready yet
			meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
				Type:               infrastructurev1beta1.VPCReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             infrastructurev1beta1.ReasonVPCCreating,
				Message:            message,
			})
			captCluster.Status.Ready = false
			if err := r.Status().Update(ctx, captCluster); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to update CAPTCluster status: %v", err)
			}
			return ctrl.Result{RequeueAfter: requeueInterval}, nil
		}

		// Get VPC ID from secret
		if vpcTemplate.Spec.WriteConnectionSecretToRef != nil {
			secret := &corev1.Secret{}
			secretName := types.NamespacedName{
				Name:      vpcTemplate.Spec.WriteConnectionSecretToRef.Name,
				Namespace: vpcTemplate.Spec.WriteConnectionSecretToRef.Namespace,
			}
			if err := r.Get(ctx, secretName, secret); err != nil {
				if apierrors.IsNotFound(err) {
					// Secret not found yet, requeue
					return ctrl.Result{RequeueAfter: requeueInterval}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to get VPC connection secret: %v", err)
			}

			vpcID, ok := secret.Data["vpc_id"]
			if !ok {
				// vpc_id not found yet, requeue
				return ctrl.Result{RequeueAfter: requeueInterval}, nil
			}

			captCluster.Status.VPCID = string(vpcID)
		}

		// VPC is ready and VPC ID is available
		meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
			Type:               infrastructurev1beta1.VPCReadyCondition,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             infrastructurev1beta1.ReasonVPCCreated,
			Message:            "VPC has been successfully created",
		})
		captCluster.Status.Ready = true

		if err := r.Status().Update(ctx, captCluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update CAPTCluster status: %v", err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CAPTClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CAPTCluster{}).
		Complete(r)
}
