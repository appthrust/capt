package controlplane

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// defaultRequeueInterval is the default interval to requeue reconciliation
	defaultRequeueInterval = 30 * time.Second
	// errorRequeueInterval is the interval to requeue reconciliation when in error state
	errorRequeueInterval = 10 * time.Second
	// initializationRequeueInterval is the interval to requeue reconciliation during initialization
	initializationRequeueInterval = 5 * time.Second
)

// Reconciler reconciles a CAPTControlPlane object
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}
