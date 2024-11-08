package captcluster

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Result is an alias for ctrl.Result to avoid importing ctrl in multiple files
type Result = ctrl.Result

// Client is an alias for client.Client to avoid importing client in multiple files
type Client = client.Client
