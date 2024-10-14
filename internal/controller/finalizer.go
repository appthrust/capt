package controller

import (
	"context"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	captClusterFinalizer = "infrastructure.cluster.x-k8s.io/finalizer"
)

func handleFinalizer(ctx context.Context, c client.Client, captCluster *infrastructurev1beta1.CAPTCluster) error {
	// Check if the CAPTCluster instance is marked to be deleted
	if captCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(captCluster, captClusterFinalizer) {
			controllerutil.AddFinalizer(captCluster, captClusterFinalizer)
			if err := c.Update(ctx, captCluster); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(captCluster, captClusterFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := deleteExternalResources(ctx, c, captCluster); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(captCluster, captClusterFinalizer)
			if err := c.Update(ctx, captCluster); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteExternalResources(ctx context.Context, c client.Client, captCluster *infrastructurev1beta1.CAPTCluster) error {
	// Delete the associated Terraform Workspace
	return deleteWorkspace(ctx, c, captCluster)
}
