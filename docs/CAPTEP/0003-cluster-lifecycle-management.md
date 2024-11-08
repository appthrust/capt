# CAPTEP-0003: Cluster Lifecycle Management Improvement

## Summary

This proposal addresses an issue where the CAPTCluster controller creates WorkspaceTemplateApply resources even when the parent Cluster resource does not exist. This behavior violates Cluster API's design principles and has been corrected through implementation.

## Motivation

### Current Behavior

Previously, the CAPTCluster controller would proceed with the creation of WorkspaceTemplateApply resources even when it could not find the parent Cluster resource. This was evident in the following code:

```go
if err := r.Get(ctx, types.NamespacedName{Namespace: captCluster.Namespace, Name: captCluster.Name}, cluster); err != nil {
    if !apierrors.IsNotFound(err) {
        logger.Error(err, "Failed to get owner Cluster")
        return ctrl.Result{}, err
    }
    // Cluster not found, but proceeded with reconciliation
    cluster = nil
    logger.Info("Owner Cluster not found, proceeding with reconciliation")
}
```

### Problems

1. **Violation of Cluster API Principles**: Infrastructure providers in Cluster API should follow the lifecycle of their parent Cluster resource. Creating infrastructure resources without a parent Cluster violates this principle.

2. **Inconsistent Resource Management**: This behavior led to orphaned resources and inconsistent cluster state management.

3. **Potential Resource Leaks**: Without proper lifecycle management tied to the parent Cluster, resources might not be cleaned up correctly.

## Implementation

### Code Organization

The implementation has been organized into focused files:
- `controller.go`: Main reconciliation logic
- `vpc.go`: VPC-specific reconciliation
- `status.go`: Status management
- `finalizer.go`: Resource cleanup and finalizer handling

### Key Changes

1. Parent Cluster Validation:
```go
if cluster == nil {
    logger.Info("Parent cluster is nil, cannot proceed with VPC reconciliation")
    return Result{}, fmt.Errorf("parent cluster is required for VPC reconciliation")
}
```

2. Waiting State Management:
```go
meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
    Type:               WaitingForClusterCondition,
    Status:             metav1.ConditionTrue,
    LastTransitionTime: metav1.Now(),
    Reason:            "ClusterNotFound",
    Message:           "Waiting for owner Cluster to be created",
})
```

3. Resource Cleanup:
```go
func (r *Reconciler) cleanupWorkspaceTemplateApply(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) error {
    // Implementation details for cleaning up resources
}
```

### Migration Strategy

1. The changes are backward compatible
2. Existing WorkspaceTemplateApply resources will be cleaned up if their parent Cluster is missing
3. No manual intervention required

## Testing Results

The implementation has been verified through manual testing:

1. Parent Cluster Missing:
   - Confirmed WorkspaceTemplateApply is not created
   - Verified proper status conditions are set

2. Parent Cluster Present:
   - Verified WorkspaceTemplateApply creation
   - Confirmed proper owner references

3. Deletion:
   - Tested resource cleanup
   - Verified finalizer handling

## Implementation History

- [x] 2024-01-XX: Initial proposal
- [x] Implementation complete
- [x] Manual testing complete
- [x] Documentation updated

## References

1. [Cluster Parent Dependency Management Design](../design/cluster-parent-dependency.md)
2. Cluster API Provider Implementation Guide
3. Kubernetes Controller Best Practices
