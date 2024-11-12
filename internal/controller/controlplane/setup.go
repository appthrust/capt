package controlplane

import (
	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1beta1.CAPTControlPlane{}).
		Complete(r)
}
