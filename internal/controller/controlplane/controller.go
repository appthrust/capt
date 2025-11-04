package controlplane

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes/finalizers,verbs=update
// Additional permissions required by the control plane reconciler
// +kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces;workspaces/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates;workspacetemplateapplies;workspacetemplateapplies/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch;update;patch

const (
	// CAPTControlPlaneFinalizer is the finalizer added to CAPTControlPlane instances
	CAPTControlPlaneFinalizer = "controlplane.cluster.x-k8s.io/captcontrolplane"
	clusterKind               = "Cluster"
)

// createKubeconfigWorkspaceTemplateApply creates a WorkspaceTemplateApply for kubeconfig generation
func (r *Reconciler) createKubeconfigWorkspaceTemplateApply(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	logger := log.FromContext(ctx)


	// Get region from ControlPlaneConfig or cluster annotations
	var region string
	if controlPlane.Spec.ControlPlaneConfig != nil {
		region = controlPlane.Spec.ControlPlaneConfig.Region
	}
	if region == "" && cluster != nil {
		// Fallback: Cluster topology.variables[name=region]
		if cluster.Spec.Topology != nil {
			for _, v := range cluster.Spec.Topology.Variables {
				if v.Name == "region" {
					var s string
					if err := json.Unmarshal(v.Value.Raw, &s); err == nil && s != "" {
						region = s
						break
					}
				}
			}
		}
	}
	if region == "" && cluster != nil {
		// Fallback: Cluster annotation
		if ann := cluster.GetAnnotations(); ann != nil {
			if v := ann["cluster.x-k8s.io/region"]; v != "" {
				region = v
			} else if v := ann["installation.appthrust.com/aws-primary-region"]; v != "" {
				region = v
			}
		}
	}
	if region == "" {
		logger.Info("Region not found in ControlPlaneConfig, topology variables, or annotations")
		return nil
	}

	// Create kubeconfig WorkspaceTemplateApply
	kubeconfigApplyName := fmt.Sprintf("%s-kubeconfig-apply", controlPlane.Name)
	kubeconfigApply := &infrastructurev1beta1.WorkspaceTemplateApply{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconfigApplyName,
			Namespace: controlPlane.Namespace,
		},
		Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
			TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
				Name:      "eks-kubeconfig-template",
				Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
			},
			Variables: map[string]string{
				"cluster_name": cluster.Name,
                "region": region,
			},
			WriteConnectionSecretToRef: &xpv1.SecretReference{
				Name:      fmt.Sprintf("%s-outputs-kubeconfig", cluster.Name),
				Namespace: controlPlane.Namespace,
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(controlPlane, kubeconfigApply, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %v", err)
	}

	// Create or update the WorkspaceTemplateApply
	existingApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      kubeconfigApplyName,
		Namespace: controlPlane.Namespace,
	}, existingApply)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new WorkspaceTemplateApply
			if err := r.Create(ctx, kubeconfigApply); err != nil {
				return fmt.Errorf("failed to create kubeconfig WorkspaceTemplateApply: %v", err)
			}
			logger.Info("Created kubeconfig WorkspaceTemplateApply")
		} else {
			return fmt.Errorf("failed to get kubeconfig WorkspaceTemplateApply: %v", err)
		}
	} else {
		// Update existing WorkspaceTemplateApply only if spec changed
		if !reflect.DeepEqual(existingApply.Spec, kubeconfigApply.Spec) {
			existingApply.Spec = kubeconfigApply.Spec
			if err := r.Update(ctx, existingApply); err != nil {
				return fmt.Errorf("failed to update kubeconfig WorkspaceTemplateApply: %v", err)
			}
			logger.Info("Updated kubeconfig WorkspaceTemplateApply")
		} else {
			logger.Info("Kubeconfig WorkspaceTemplateApply unchanged, skipping update")
		}
	}

	return nil
}

// cleanupResources cleans up all resources associated with the CAPTControlPlane
func (r *Reconciler) cleanupResources(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) error {
	logger := log.FromContext(ctx)

	// 親クラスタを取得
	cluster := &clusterv1.Cluster{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      controlPlane.Name,
		Namespace: controlPlane.Namespace,
	}, cluster); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get parent cluster: %v", err)
		}
		// 親クラスタが既に削除されている場合は処理を続行
		logger.Info("Parent cluster already deleted")
		cluster = nil
	} else {
		// エンドポイントを削除
		cluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{}
		if err := r.Update(ctx, cluster); err != nil {
			return fmt.Errorf("failed to update cluster endpoint: %v", err)
		}
		logger.Info("Successfully cleared control plane endpoint")
	}

	// Find and check associated WorkspaceTemplateApply
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
		// Get workspace name before deleting WorkspaceTemplateApply
		workspaceName := workspaceApply.Status.WorkspaceName

		// Delete WorkspaceTemplateApply
		if err := r.Delete(ctx, workspaceApply); err != nil {
			logger.Error(err, "Failed to delete WorkspaceTemplateApply")
			return fmt.Errorf("failed to delete WorkspaceTemplateApply: %v", err)
		}
		logger.Info("Successfully deleted WorkspaceTemplateApply")

		// If we have a workspace name, check if it's fully deleted
		if workspaceName != "" {
			workspace := &unstructured.Unstructured{}
			workspace.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "tf.upbound.io",
				Version: "v1beta1",
				Kind:    "Workspace",
			})

			err := r.Get(ctx, types.NamespacedName{
				Name:      workspaceName,
				Namespace: controlPlane.Namespace,
			}, workspace)

			if err == nil {
				// Workspace still exists, requeue
				logger.Info("Waiting for workspace to be deleted",
					"name", workspaceName,
					"namespace", controlPlane.Namespace)
				return fmt.Errorf("waiting for workspace deletion")
			} else if !apierrors.IsNotFound(err) {
				// Unexpected error
				return fmt.Errorf("failed to check workspace status: %v", err)
			}
			// Workspace not found means it's deleted
			logger.Info("Workspace successfully deleted",
				"name", workspaceName,
				"namespace", controlPlane.Namespace)
		}
	} else if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get WorkspaceTemplateApply: %v", err)
	}

	// Delete kubeconfig WorkspaceTemplateApply if it exists
	kubeconfigApplyName := fmt.Sprintf("%s-kubeconfig-apply", controlPlane.Name)
	kubeconfigApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      kubeconfigApplyName,
		Namespace: controlPlane.Namespace,
	}, kubeconfigApply)

	if err == nil {
		if err := r.Delete(ctx, kubeconfigApply); err != nil {
			logger.Error(err, "Failed to delete kubeconfig WorkspaceTemplateApply")
			return fmt.Errorf("failed to delete kubeconfig WorkspaceTemplateApply: %v", err)
		}
		logger.Info("Successfully deleted kubeconfig WorkspaceTemplateApply")
	} else if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get kubeconfig WorkspaceTemplateApply: %v", err)
	}

	// Delete generated kubeconfig outputs secret if it exists
	if cluster != nil {
		outputsSecret := &corev1.Secret{}
		outputsName := fmt.Sprintf("%s-outputs-kubeconfig", cluster.Name)
		if err := r.Get(ctx, client.ObjectKey{Name: outputsName, Namespace: controlPlane.Namespace}, outputsSecret); err == nil {
			if delErr := r.Delete(ctx, outputsSecret); delErr != nil && !apierrors.IsNotFound(delErr) {
				logger.Error(delErr, "Failed to delete outputs kubeconfig secret", "name", outputsName)
				return fmt.Errorf("failed to delete outputs secret: %v", delErr)
			}
			logger.Info("Successfully deleted outputs kubeconfig secret", "name", outputsName)
		}
	}

	// Delete CA secret if it exists
	caSecret := &corev1.Secret{}
	caName := fmt.Sprintf("%s-ca", controlPlane.Name)
	if err := r.Get(ctx, client.ObjectKey{Name: caName, Namespace: controlPlane.Namespace}, caSecret); err == nil {
		if delErr := r.Delete(ctx, caSecret); delErr != nil && !apierrors.IsNotFound(delErr) {
			logger.Error(delErr, "Failed to delete CA secret", "name", caName)
			return fmt.Errorf("failed to delete CA secret: %v", delErr)
		}
		logger.Info("Successfully deleted CA secret", "name", caName)
	}

	return nil
}

// Reconcile handles CAPTControlPlane events
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CAPTControlPlane")

	// Get CAPTControlPlane
	controlPlane := &controlplanev1beta1.CAPTControlPlane{}
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get owner Cluster
	var clusterName string
	if name, ok := controlPlane.Labels[clusterv1.ClusterNameLabel]; ok && name != "" {
		clusterName = name
	} else {
		// Fallback to OwnerReference if present
		for _, ref := range controlPlane.OwnerReferences {
			if ref.Kind == clusterKind && ref.APIVersion == clusterv1.GroupVersion.String() && ref.Name != "" {
				clusterName = ref.Name
				break
			}
		}
		// As a final fallback in v0.4.x, default to same name as controlPlane
		if clusterName == "" {
			clusterName = controlPlane.Name
		}
	}

	var cluster *clusterv1.Cluster
	if clusterName != "" {
		c := &clusterv1.Cluster{}
		if err := r.Get(ctx, types.NamespacedName{Name: clusterName, Namespace: controlPlane.Namespace}, c); err != nil {
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
		} else {
			cluster = c
		}
	}

	// Ensure region/environment annotations on parent Cluster from ControlPlane spec/AdditionalTags
	if cluster != nil {
		var region string
		if controlPlane.Spec.ControlPlaneConfig != nil {
			region = controlPlane.Spec.ControlPlaneConfig.Region
		}
		var environment string
		if len(controlPlane.Spec.AdditionalTags) > 0 {
			if v, ok := controlPlane.Spec.AdditionalTags["Environment"]; ok {
				environment = v
			}
		}

		if region != "" || environment != "" {
			patchBase := cluster.DeepCopy()
			ann := cluster.GetAnnotations()
			if ann == nil {
				ann = map[string]string{}
			}
			if region != "" && ann["cluster.x-k8s.io/region"] != region {
				ann["cluster.x-k8s.io/region"] = region
			}
			if environment != "" && ann["capt.dev/environment"] != environment {
				ann["capt.dev/environment"] = environment
			}
			cluster.SetAnnotations(ann)
			if err := r.Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
				logger.Error(err, "Failed to patch cluster annotations")
			}
		}
	}

	// Handle deletion
	if !controlPlane.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(controlPlane, CAPTControlPlaneFinalizer) {
			// Clean up associated resources
			if err := r.cleanupResources(ctx, controlPlane); err != nil {
				logger.Error(err, "Failed to cleanup resources")
				return ctrl.Result{RequeueAfter: defaultRequeueInterval}, err
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(controlPlane, CAPTControlPlaneFinalizer)

			// Update the object to remove the finalizer
			if err := r.Update(ctx, controlPlane); err != nil {
				if !apierrors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}

			logger.Info("Successfully removed finalizer")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(controlPlane, CAPTControlPlaneFinalizer) {
		controllerutil.AddFinalizer(controlPlane, CAPTControlPlaneFinalizer)
		if err := r.Update(ctx, controlPlane); err != nil {
			return ctrl.Result{}, err
		}

		// Fetch the updated object
		if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle missing cluster case
	if cluster == nil {
		logger.Info("Owner cluster not found")
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:               controlplanev1beta1.ControlPlaneReadyCondition,
			Status:             metav1.ConditionFalse,
			Reason:             controlplanev1beta1.ReasonCreating,
			Message:            "Waiting for owner cluster",
			LastTransitionTime: metav1.Now(),
		})
		if err := r.Status().Update(ctx, controlPlane); err != nil {
			return ctrl.Result{}, err
		}

		// Fetch the updated object
		if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: initializationRequeueInterval}, nil
	}

	// Set owner reference if cluster exists
	if err := r.setOwnerReference(ctx, controlPlane, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Fetch the updated object after setting owner reference
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Get WorkspaceTemplate
	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
		Namespace: controlPlane.Namespace,
	}, workspaceTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get WorkspaceTemplate")
			result, setErr := r.setFailedStatus(ctx, controlPlane, cluster, "WorkspaceTemplateNotFound", fmt.Sprintf("Failed to get WorkspaceTemplate: %v", err))
			if setErr != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
			}
			return result, err
		}
		return ctrl.Result{}, err
	}

	// Ensure EC2 Spot Service-Linked Role exists
	if err := r.reconcileSpotServiceLinkedRole(ctx, controlPlane); err != nil {
		logger.Error(err, "Failed to reconcile Spot Service-Linked Role")
		result, setErr := r.setFailedStatus(ctx, controlPlane, cluster, "SpotServiceLinkedRoleFailed", fmt.Sprintf("Failed to reconcile Spot Service-Linked Role: %v", err))
		if setErr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return result, err
	}

	// Get or create WorkspaceTemplateApply
	workspaceApply, err := r.getOrCreateWorkspaceTemplateApply(ctx, controlPlane)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Set the WorkspaceTemplateApplyName only if it changed to avoid no-op updates triggering reconcile loops
	if controlPlane.Spec.WorkspaceTemplateApplyName != workspaceApply.Name {
		cp := controlPlane.DeepCopy()
		cp.Spec.WorkspaceTemplateApplyName = workspaceApply.Name
		if err := r.Update(ctx, cp); err != nil {
			return ctrl.Result{}, err
		}
		// Fetch the updated object
		if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update status based on WorkspaceTemplateApply conditions
	result, err := r.updateStatus(ctx, controlPlane, workspaceApply, cluster)
	if err != nil {
		return result, err
	}

	// Create kubeconfig WorkspaceTemplateApply if the main WorkspaceTemplateApply is ready
	if err := r.createKubeconfigWorkspaceTemplateApply(ctx, controlPlane, cluster, workspaceApply); err != nil {
		logger.Error(err, "Failed to create kubeconfig WorkspaceTemplateApply")
		if _, setErr := r.setFailedStatus(ctx, controlPlane, cluster, "KubeconfigGenerationFailed", fmt.Sprintf("Failed to create kubeconfig WorkspaceTemplateApply: %v", err)); setErr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return ctrl.Result{RequeueAfter: errorRequeueInterval}, err
	}

	// If kubeconfig WorkspaceTemplateApply is Ready, prioritize generating kubeconfig secret
	// and mark secrets ready, skipping main workspace-based secret reconciliation to avoid deadlocks.
	kubeApplyName := fmt.Sprintf("%s-kubeconfig-apply", controlPlane.Name)
	kubeApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	kubeReady := false
	if err := r.Get(ctx, types.NamespacedName{Name: kubeApplyName, Namespace: controlPlane.Namespace}, kubeApply); err == nil {
		for _, cond := range kubeApply.Status.Conditions {
			if cond.Type == xpv1.TypeReady && cond.Status == corev1.ConditionTrue {
				kubeReady = true
				break
			}
		}
	}
	if kubeReady {
		// Ensure kubeconfig Secret exists; reconcileKubeconfigSecret will also mark SecretsReady.
		if err := r.reconcileKubeconfigSecret(ctx, controlPlane, cluster); err != nil {
			logger.Error(err, "Failed to reconcile kubeconfig secret")
			if _, setErr := r.setFailedStatus(ctx, controlPlane, cluster, "KubeconfigSecretFailed", fmt.Sprintf("Failed to reconcile kubeconfig secret: %v", err)); setErr != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
			}
			return ctrl.Result{RequeueAfter: errorRequeueInterval}, err
		}
	} else {
		// Fallback: reconcile secrets based on the main WorkspaceTemplateApply
		if err := r.reconcileSecrets(ctx, controlPlane, cluster, workspaceApply); err != nil {
			logger.Error(err, "Failed to reconcile secrets")
			if _, setErr := r.setFailedStatus(ctx, controlPlane, cluster, "SecretReconciliationFailed", fmt.Sprintf("Failed to reconcile secrets: %v", err)); setErr != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
			}
			return ctrl.Result{RequeueAfter: errorRequeueInterval}, err
		}
	}

	// Fetch the final updated object
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Return the result from updateStatus to maintain the requeue interval
	return result, nil
}

// setOwnerReference sets the owner reference to the parent Cluster
func (r *Reconciler) setOwnerReference(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) error {
	if cluster == nil {
		return nil
	}

	// Check if owner reference already exists
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
