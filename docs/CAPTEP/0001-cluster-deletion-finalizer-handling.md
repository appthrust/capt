# CAPTEP-0001: Improving Cluster and CAPTCluster Deletion Handling

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [User Stories](#user-stories)
  - [Implementation Details](#implementation-details)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Alternatives Considered](#alternatives-considered)
- [Upgrade Strategy](#upgrade-strategy)

## Summary

This proposal aims to improve the deletion handling between Cluster and CAPTCluster resources by addressing the circular dependency issues and finalizer handling that currently cause deletion operations to become stuck.

## Motivation

Currently, when a user attempts to delete a Cluster resource, the deletion process can become stuck due to circular dependencies between the Cluster and CAPTCluster resources, combined with the current finalizer implementation. This results in a situation where manual intervention is required to remove finalizers and complete the deletion process.

### Goals

- Implement proper handling of resource deletion between Cluster and CAPTCluster
- Ensure clean-up of all dependent resources (WorkspaceTemplateApply, Workspace)
- Eliminate the need for manual finalizer removal
- Maintain data consistency and prevent resource leaks

### Non-Goals

- Modifying the core Cluster API controller behavior
- Changing the overall architecture of how Cluster references infrastructure providers

## Proposal

### User Stories

#### Story 1: Clean Cluster Deletion

As a cluster operator, I want to delete a Cluster and have all associated resources (CAPTCluster, WorkspaceTemplateApply, Workspace) cleaned up automatically without manual intervention.

#### Story 2: Graceful Deletion with Retained Resources

As a cluster operator, I want to delete a Cluster while retaining specified resources (e.g., VPC) as configured in the CAPTCluster spec.

### Implementation Details

#### Current Implementation

The current implementation has the following issues:

1. Circular Dependencies:
```yaml
# CAPTCluster
metadata:
  ownerReferences:
    - kind: Cluster
      name: demo-cluster

# Cluster
spec:
  infrastructureRef:
    kind: CAPTCluster
    name: demo-cluster
```

2. Finalizer Chain:
```
Cluster (cluster.cluster.x-k8s.io)
└── CAPTCluster (infrastructure.cluster.x-k8s.io/finalizer)
    └── WorkspaceTemplateApply
        └── Workspace
```

3. Deletion Process:
- Cluster deletion is blocked waiting for CAPTCluster deletion
- CAPTCluster deletion is blocked by ownerReference to Cluster
- Results in a deadlock requiring manual intervention

#### Proposed Changes

1. Enhanced CAPTCluster Finalizer:
```go
func deleteExternalResources(ctx context.Context, c client.Client, captCluster *infrastructurev1beta1.CAPTCluster) error {
    // Check if owner Cluster is being deleted
    if ownerCluster, err := getOwnerCluster(ctx, c, captCluster); err == nil && ownerCluster.DeletionTimestamp != nil {
        // Early finalizer removal if owner is being deleted
        return nil
    }

    // Process WorkspaceTemplateApply deletion
    if err := handleWorkspaceTemplateApplyDeletion(ctx, c, captCluster); err != nil {
        return err
    }

    return nil
}
```

2. Deletion Order Management:
```go
func (r *CAPTClusterReconciler) reconcileDelete(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (ctrl.Result, error) {
    // Check if owner Cluster exists and is being deleted
    ownerCluster, err := r.getOwnerCluster(ctx, captCluster)
    if err == nil && ownerCluster.DeletionTimestamp != nil {
        // Skip normal deletion process and remove finalizer
        controllerutil.RemoveFinalizer(captCluster, captClusterFinalizer)
        return ctrl.Result{}, r.Update(ctx, captCluster)
    }

    // Normal deletion process
    return r.handleNormalDeletion(ctx, captCluster)
}
```

3. Resource Clean-up Verification:
```go
func (r *CAPTClusterReconciler) handleNormalDeletion(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (ctrl.Result, error) {
    // Verify WorkspaceTemplateApply deletion
    if err := r.verifyWorkspaceTemplateApplyDeletion(ctx, captCluster); err != nil {
        return ctrl.Result{RequeueAfter: requeueInterval}, err
    }

    // Verify Workspace deletion
    if err := r.verifyWorkspaceDeletion(ctx, captCluster); err != nil {
        return ctrl.Result{RequeueAfter: requeueInterval}, err
    }

    return ctrl.Result{}, nil
}
```

### Risks and Mitigations

1. Risk: Resource Leaks
   - Mitigation: Implement thorough verification of resource deletion before removing finalizers

2. Risk: Race Conditions
   - Mitigation: Use proper Kubernetes watch mechanisms and status conditions

3. Risk: Backward Compatibility
   - Mitigation: Maintain support for existing resource configurations while implementing new behavior

## Alternatives Considered

### Alternative 1: Manual Finalizer Management

Continue with the current approach of requiring manual finalizer removal:
- Pros: Simpler implementation
- Cons: Requires manual intervention, poor user experience

### Alternative 2: Remove Owner References

Remove the owner reference from CAPTCluster to Cluster:
- Pros: Eliminates circular dependency
- Cons: Loses benefits of Kubernetes garbage collection

### Alternative 3: Custom Garbage Collection

Implement custom garbage collection logic:
- Pros: Full control over deletion process
- Cons: Complex implementation, potential for bugs

## Upgrade Strategy

1. Implementation Phases:
   - Phase 1: Implement enhanced finalizer logic
   - Phase 2: Add status conditions for deletion tracking
   - Phase 3: Add metrics for deletion monitoring

2. Backward Compatibility:
   - Maintain support for existing resource configurations
   - Document manual intervention process for older versions

3. Testing:
   - Add comprehensive e2e tests for deletion scenarios
   - Include upgrade tests from previous versions
