package controlplane

import (
    controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/builder"
    "sigs.k8s.io/controller-runtime/pkg/predicate"
    "sigs.k8s.io/controller-runtime/pkg/log"
)

// ControllerVersion indicates the running version of the CAPTControlPlane controller
const ControllerVersion = "v0.4.3"

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Log controller version at startup
    log.Log.WithName("captcontrolplane").Info("Starting CAPTControlPlane controller", "version", ControllerVersion)
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1beta1.CAPTControlPlane{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
