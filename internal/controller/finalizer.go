package controller

import (
	"context"
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	logger := log.FromContext(ctx)

	// Check if VPC should be retained
	if captCluster.Spec.RetainVPCOnDelete && captCluster.Spec.VPCTemplateRef != nil {
		logger.Info("RetainVPCOnDelete is true, skipping VPC deletion",
			"vpcId", captCluster.Status.VPCID,
			"workspaceTemplateApplyName", captCluster.Spec.WorkspaceTemplateApplyName)
		return nil
	}

	// Find and delete associated WorkspaceTemplateApply
	if captCluster.Spec.WorkspaceTemplateApplyName != "" {
		workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
		err := c.Get(ctx, types.NamespacedName{
			Name:      captCluster.Spec.WorkspaceTemplateApplyName,
			Namespace: captCluster.Namespace,
		}, workspaceApply)

		if err == nil {
			// Check if WorkspaceTemplateApply is already being deleted
			if workspaceApply.DeletionTimestamp != nil {
				logger.Info("WorkspaceTemplateApply is being deleted, waiting",
					"name", workspaceApply.Name)
				return fmt.Errorf("waiting for WorkspaceTemplateApply deletion")
			}

			// WorkspaceTemplateApply exists and not being deleted, delete it
			if err := c.Delete(ctx, workspaceApply); err != nil {
				logger.Error(err, "Failed to delete WorkspaceTemplateApply")
				return fmt.Errorf("failed to delete WorkspaceTemplateApply: %v", err)
			}
			logger.Info("Initiated WorkspaceTemplateApply deletion",
				"name", workspaceApply.Name)
			return fmt.Errorf("waiting for WorkspaceTemplateApply deletion")
		}
	}

	return nil
}
