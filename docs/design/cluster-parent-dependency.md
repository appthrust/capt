# Cluster Parent Dependency Management

## Overview

This document describes the design and implementation of parent cluster dependency management in CAPT. The implementation ensures that infrastructure resources (specifically WorkspaceTemplateApply) are not created until the parent Cluster resource exists, adhering to Cluster API's design principles.

## Background

In Cluster API, infrastructure providers should follow the lifecycle of their parent Cluster resource. Previously, our implementation would create WorkspaceTemplateApply resources even when the parent Cluster did not exist, which could lead to resource management issues and violated Cluster API's design principles.

## Design Goals

1. Ensure infrastructure resources are only created when parent Cluster exists
2. Provide clear status information about waiting state
3. Clean up any existing resources if parent Cluster is not present
4. Maintain proper resource lifecycle management

## Implementation Details

### Parent Cluster Validation

```go
// Get owner Cluster
cluster := &clusterv1.Cluster{}
if err := r.Get(ctx, types.NamespacedName{...}, cluster); err != nil {
    if !apierrors.IsNotFound(err) {
        logger.Error(err, "Failed to get owner Cluster")
        return Result{}, err
    }
    return r.handleMissingCluster(ctx, captCluster)
}
```

The controller checks for the parent Cluster's existence before proceeding with any resource creation.

### Waiting State Management

When the parent Cluster is not found, the controller:
1. Sets a WaitingForCluster condition
2. Marks the CAPTCluster as not ready
3. Cleans up any existing WorkspaceTemplateApply resources
4. Requeues the reconciliation

```go
meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
    Type:               WaitingForClusterCondition,
    Status:             metav1.ConditionTrue,
    LastTransitionTime: metav1.Now(),
    Reason:            "ClusterNotFound",
    Message:           "Waiting for owner Cluster to be created",
})
```

### Resource Cleanup

When waiting for the parent Cluster, the controller ensures no orphaned resources exist:

```go
func (r *Reconciler) cleanupWorkspaceTemplateApply(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) error {
    // Remove any existing WorkspaceTemplateApply
    // Clear references
    // Update status
}
```

### Code Organization

The implementation is organized into focused files:
- `controller.go`: Main reconciliation logic
- `vpc.go`: VPC-specific reconciliation
- `status.go`: Status management
- `finalizer.go`: Resource cleanup and finalizer handling

## Status Management

The controller uses conditions to track the state:

1. WaitingForCluster:
   - Set when parent Cluster is missing
   - Cleared when parent Cluster is found

2. VPCReady:
   - Set based on WorkspaceTemplateApply status
   - Only updated when parent Cluster exists

## Deletion Handling

Resource deletion follows this sequence:

1. Check RetainVPCOnDelete flag
2. Delete WorkspaceTemplateApply if retention not requested
3. Verify deletion completion
4. Remove finalizer

## Testing Strategy

The implementation has been verified through manual testing:

1. Parent Cluster Missing:
   - Confirmed WorkspaceTemplateApply is not created
   - Verified proper status conditions are set
   - Checked requeuing behavior

2. Parent Cluster Present:
   - Verified WorkspaceTemplateApply creation
   - Confirmed proper owner references
   - Validated status updates

3. Deletion:
   - Tested resource cleanup
   - Verified finalizer handling
   - Confirmed retention flag behavior

## Future Considerations

1. Automated Testing:
   - Add unit tests for parent dependency logic
   - Implement e2e tests for lifecycle management

2. Status Enhancement:
   - Add more detailed status conditions
   - Improve error reporting

3. Resource Management:
   - Consider handling additional dependent resources
   - Implement resource adoption strategies

## References

1. Cluster API Design Principles
2. Kubernetes Controller Patterns
3. CAPT Architecture Documentation
