package controller

import (
	"github.com/appthrust/capt/internal/controller/captcluster"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupCAPTClusterController adds the CAPTCluster controller to the manager
func SetupCAPTClusterController(mgr ctrl.Manager) error {
	reconciler := &captcluster.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	return reconciler.SetupWithManager(mgr)
}
