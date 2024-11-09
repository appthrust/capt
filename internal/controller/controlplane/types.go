package controlplane

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// requeueInterval is the interval to requeue reconciliation
	requeueInterval = 10 * time.Second
)

// Result represents the result of a reconciliation
type Result struct {
	// Requeue tells the Controller to requeue the reconcile key
	Requeue bool
	// RequeueAfter if greater than 0, tells the Controller to requeue the reconcile key after the Duration
	RequeueAfter time.Duration
}

// Reconciler reconciles a CAPTControlPlane object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}
