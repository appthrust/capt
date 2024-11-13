package controlplane

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

// CaptControlPlaneTemplateReconciler reconciles a CaptControlPlaneTemplate object
type CaptControlPlaneTemplateReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanetemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanetemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanetemplates/finalizers,verbs=update
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch

// Reconcile handles CaptControlPlaneTemplate reconciliation
func (r *CaptControlPlaneTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CaptControlPlaneTemplate instance
	template := &controlplanev1beta1.CaptControlPlaneTemplate{}
	if err := r.Get(ctx, req.NamespacedName, template); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Note: Variable resolution and patch application are handled by the Cluster API controller.
	// This controller only needs to validate the template and ensure referenced resources exist.

	// Validate WorkspaceTemplate reference
	if err := r.validateWorkspaceTemplateRef(ctx, template); err != nil {
		log.Error(err, "failed to validate WorkspaceTemplate reference")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// validateWorkspaceTemplateRef validates that the referenced WorkspaceTemplate exists
func (r *CaptControlPlaneTemplateReconciler) validateWorkspaceTemplateRef(ctx context.Context, template *controlplanev1beta1.CaptControlPlaneTemplate) error {
	if template.Spec.Template.Spec.WorkspaceTemplateRef.Name == "" {
		return fmt.Errorf("workspaceTemplateRef.name cannot be empty")
	}

	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	key := client.ObjectKey{
		Name:      template.Spec.Template.Spec.WorkspaceTemplateRef.Name,
		Namespace: template.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
	}
	if key.Namespace == "" {
		key.Namespace = template.Namespace
	}

	if err := r.Get(ctx, key, workspaceTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "WorkspaceTemplate %s/%s not found", key.Namespace, key.Name)
		}
		return errors.Wrapf(err, "failed to get WorkspaceTemplate %s/%s", key.Namespace, key.Name)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CaptControlPlaneTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1beta1.CaptControlPlaneTemplate{}).
		Complete(r)
}
