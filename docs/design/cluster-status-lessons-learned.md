# Cluster Status Management - Lessons Learned

## Overview

This document captures the lessons learned while implementing Cluster status management in the CAPT provider.

## Key Findings

### 1. Status Update Loops

**Problem:**
- Implementing Watches for both CAPTCluster and Cluster resources caused infinite reconciliation loops
- Status updates triggered new reconciliations, leading to high CPU usage

**Solution:**
- Remove explicit Watches for Cluster resources
- Rely on OwnerReferences for automatic reconciliation
- Only update status when actual changes occur

### 2. Phase Management

**Problem:**
- Cluster phase remained in "Provisioning" even when both infrastructure and control plane were ready
- Phase updates were not properly propagated through the status chain

**Root Cause:**
- Status updates were not properly synchronized between CAPTCluster and CAPTControlPlane
- Phase transition logic was not properly implemented in both controllers

**Solution:**
- Implement proper phase transition logic in both controllers
- Ensure status updates are properly synchronized
- Use server-side apply for status updates to avoid conflicts

### 3. Controller Design

**Important Considerations:**
1. Status Updates
   - Always update provider status (CAPTCluster/CAPTControlPlane) before Cluster status
   - Use patch operations instead of updates to minimize conflicts
   - Avoid unnecessary status updates that could trigger reconciliation loops

2. Resource Ownership
   - Proper OwnerReference setup is crucial for garbage collection
   - Let Kubernetes handle resource relationships through OwnerReferences

3. Error Handling
   - Propagate detailed error information through status conditions
   - Use appropriate error types from ClusterAPI

## Best Practices

1. **Controller Implementation**
   ```go
   func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
       return ctrl.NewControllerManagedBy(mgr).
           For(&infrastructurev1beta1.CAPTCluster{}).
           Owns(&infrastructurev1beta1.WorkspaceTemplateApply{}).
           Complete(r)
   }
   ```
   - Keep controller setup simple
   - Avoid unnecessary Watches
   - Let OwnerReferences handle resource relationships

2. **Status Updates**
   ```go
   // Before updating status, check if there are actual changes
   if !reflect.DeepEqual(oldStatus, newStatus) {
       if err := r.Status().Update(ctx, resource); err != nil {
           return err
       }
   }
   ```
   - Only update status when there are actual changes
   - Use DeepEqual to compare status before updating

3. **Phase Transitions**
   ```go
   // Update phase if both infrastructure and control plane are ready
   if cluster.Status.InfrastructureReady && cluster.Status.ControlPlaneReady {
       cluster.Status.Phase = string(clusterv1.ClusterPhaseProvisioned)
   }
   ```
   - Make phase transitions explicit and deterministic
   - Consider all required conditions before changing phase

## Common Pitfalls

1. **Status Update Loops**
   - Avoid watching resources that trigger status updates
   - Be careful with status updates that could trigger new reconciliations

2. **Resource Dependencies**
   - Don't assume resources are created in a specific order
   - Handle missing resources gracefully
   - Use proper error handling and requeue mechanisms

3. **Status Synchronization**
   - Be careful with concurrent status updates
   - Use server-side apply when possible
   - Consider using generation/observedGeneration for change detection

## Future Considerations

1. **Status Update Optimization**
   - Implement better status update batching
   - Add status update rate limiting
   - Consider using conditions for more granular status tracking

2. **Testing**
   - Add more comprehensive unit tests for status management
   - Implement integration tests for phase transitions
   - Add stress tests for concurrent status updates

3. **Monitoring**
   - Add metrics for status update frequency
   - Monitor reconciliation loop frequency
   - Track phase transition timing
