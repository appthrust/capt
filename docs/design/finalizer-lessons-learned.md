# Finalizer Implementation Lessons Learned

## Overview

This document captures the insights, discoveries, and design decisions made during the implementation and troubleshooting of finalizer handling in CAPT, particularly focusing on the interaction between Cluster API resources and infrastructure providers.

## Problem Discovery

### Initial Symptoms

1. Resource Deletion Hanging
   - CaptCluster deletion would not complete
   - WorkspaceTemplateApply would be deleted
   - Workspace would remain
   - Manual finalizer removal was required

### Investigation Process

1. First Layer: WorkspaceTemplateApply Deletion
   ```yaml
   # WorkspaceTemplateApply successfully deleted, but Workspace remained
   # This led to investigation of finalizer implementation
   ```

2. Second Layer: CaptCluster Finalizer
   ```yaml
   metadata:
     finalizers:
     - infrastructure.cluster.x-k8s.io/finalizer
   ```

3. Third Layer: Cluster Reference
   ```yaml
   # Discovered circular reference between Cluster and CaptCluster
   metadata:
     ownerReferences:
     - kind: Cluster
       name: demo-cluster
   spec:
     infrastructureRef:
       kind: CAPTCluster
       name: demo-cluster
   ```

## Key Discoveries

### 1. Cluster API Resource Dependencies

1. Reference Chain:
   ```
   Cluster
   ├── infrastructureRef -> CAPTCluster
   └── controlPlaneRef -> CAPTControlPlane
   ```

2. Deletion Order:
   ```
   Cluster (waiting for children)
   └── CaptCluster (waiting for owner)
      └── WorkspaceTemplateApply
         └── Workspace
   ```

### 2. Finalizer Behavior

1. Core Controller Behavior:
   - Cluster API's core controller maintains its finalizer until all referenced resources are deleted
   - This includes both infrastructureRef and controlPlaneRef resources

2. Infrastructure Provider Behavior:
   - CaptCluster maintains its finalizer until WorkspaceTemplateApply is deleted
   - WorkspaceTemplateApply maintains its finalizer until Workspace is deleted

### 3. Circular Dependencies

1. Owner Reference Impact:
   ```yaml
   # CaptCluster cannot be deleted due to owner reference
   ownerReferences:
   - kind: Cluster
     name: demo-cluster
     blockOwnerDeletion: true
   ```

2. Infrastructure Reference Impact:
   ```yaml
   # Cluster cannot be deleted due to infrastructure reference
   spec:
     infrastructureRef:
       kind: CAPTCluster
       name: demo-cluster
   ```

## Design Decisions

### 1. Finalizer Implementation

Original Implementation:
```go
func deleteExternalResources(ctx context.Context, c client.Client, captCluster *infrastructurev1beta1.CAPTCluster) error {
    // Delete WorkspaceTemplateApply without waiting
    if err := c.Delete(ctx, workspaceApply); err != nil {
        return err
    }
    return nil
}
```

Issues:
- No verification of resource deletion
- No handling of dependent resources
- No consideration of parent-child relationships

Improved Implementation:
```go
func deleteExternalResources(ctx context.Context, c client.Client, captCluster *infrastructurev1beta1.CAPTCluster) error {
    // Check if WorkspaceTemplateApply is already being deleted
    if workspaceApply.DeletionTimestamp != nil {
        return fmt.Errorf("waiting for WorkspaceTemplateApply deletion")
    }

    // Delete and wait for completion
    if err := c.Delete(ctx, workspaceApply); err != nil {
        return err
    }
    return fmt.Errorf("waiting for WorkspaceTemplateApply deletion")
}
```

Benefits:
- Ensures complete resource cleanup
- Prevents premature finalizer removal
- Maintains data consistency

### 2. Resource Deletion Strategy

1. Explicit Deletion Order:
   ```
   1. Initiate Cluster deletion
   2. CaptCluster recognizes owner deletion
   3. WorkspaceTemplateApply deletion
   4. Workspace deletion verification
   5. Remove CaptCluster finalizer
   6. Remove Cluster finalizer
   ```

2. Status Tracking:
   - Monitor deletion progress through status conditions
   - Provide clear feedback about deletion state

## Implementation Considerations

### 1. Error Handling

1. Temporary Errors:
   ```go
   // Return error to requeue
   return ctrl.Result{RequeueAfter: requeueInterval}, err
   ```

2. Permanent Errors:
   ```go
   // Update status and continue
   meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
       Type:    infrastructurev1beta1.DeletionFailedCondition,
       Status:  metav1.ConditionTrue,
       Reason:  reason,
       Message: err.Error(),
   })
   ```

### 2. Race Conditions

1. Resource State Verification:
   ```go
   // Always verify current state
   if err := r.Get(ctx, types.NamespacedName{...}, resource); err != nil {
       if !apierrors.IsNotFound(err) {
           return err
       }
       // Resource already deleted
       return nil
   }
   ```

2. Deletion Timestamp Checking:
   ```go
   if resource.DeletionTimestamp != nil {
       // Resource is being deleted
       return fmt.Errorf("waiting for deletion")
   }
   ```

## Lessons Learned

1. Finalizer Design:
   - Always verify resource deletion completion
   - Consider parent-child relationships
   - Implement proper error handling and retries

2. Resource Dependencies:
   - Understand Cluster API's resource model
   - Consider circular reference implications
   - Design for proper cleanup order

3. Testing:
   - Test deletion scenarios thoroughly
   - Include error cases and race conditions
   - Verify resource cleanup completion

4. Monitoring:
   - Add detailed logging
   - Implement status conditions
   - Consider adding metrics for deletion tracking

## Future Improvements

1. Enhanced Status Reporting:
   ```go
   type DeletionStatus struct {
       Phase           string
       StartTime      *metav1.Time
       CompletionTime *metav1.Time
       FailureReason  *string
   }
   ```

2. Metrics Implementation:
   ```go
   // Prometheus metrics for deletion tracking
   deletionDuration := prometheus.NewHistogramVec(
       prometheus.HistogramOpts{
           Name: "capt_resource_deletion_duration_seconds",
           Help: "Duration of resource deletion in seconds",
       },
       []string{"resource_type"},
   )
   ```

3. Automated Recovery:
   - Implement automatic finalizer cleanup for stuck resources
   - Add safety checks for manual intervention
