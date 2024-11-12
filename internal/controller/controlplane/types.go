package controlplane

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// requeueInterval is the interval to requeue reconciliation
	requeueInterval = 10 * time.Second
)

// Reconciler reconciles a CAPTControlPlane object
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}
