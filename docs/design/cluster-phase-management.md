# Cluster Phase Management in CAPT

## Overview

This document describes the design and implementation details for managing Cluster phases in the CAPT provider, based on lessons learned from our implementation experience.

## ClusterAPI Phase Management Design

### Phase Responsibility

In ClusterAPI, phase management follows a specific responsibility model:

1. **Control Plane Controller**
   - Primary responsibility for managing the overall cluster phase
   - Controls the transition to "Provisioned" phase
   - Has direct knowledge of the control plane's readiness

2. **Infrastructure Provider**
   - Manages only its own resource status (Ready/Not Ready)
   - Updates InfrastructureReady status
   - Should NOT directly modify cluster phase

### Phase Transitions

The cluster goes through several phases:

1. **Provisioning**
   - Initial state when infrastructure and control plane are being prepared
   - Remains in this state until both infrastructure and control plane are ready

2. **Provisioned**
   - Indicates that all necessary infrastructure and control plane components are ready
   - Controlled by the Control Plane controller
   - Requires both InfrastructureReady and ControlPlaneReady to be true

### Implementation Lessons

#### 1. Phase Update Responsibility

**Problem:**
- Both CAPTCluster and CAPTControlPlane controllers were attempting to update the cluster phase
- This led to race conditions and inconsistent phase states

**Solution:**
- Remove phase update logic from CAPTCluster controller
- Let CAPTControlPlane controller manage phase transitions
- Infrastructure provider only updates InfrastructureReady status

#### 2. Status Synchronization

**Problem:**
- Status updates were not properly synchronized between controllers
- Phase transitions were happening before all components were ready

**Solution:**
- Clear separation of responsibilities:
  ```go
  // CAPTCluster controller only updates infrastructure status
  cluster.Status.InfrastructureReady = captCluster.Status.Ready

  // CAPTControlPlane controller manages phase transitions
  if cluster.Status.ControlPlaneReady && cluster.Status.InfrastructureReady {
      cluster.Status.Phase = string(clusterv1.ClusterPhaseProvisioned)
  }
  ```

#### 3. Status Update Order

Proper order for status updates:
1. Update provider-specific status (CAPTCluster/CAPTControlPlane)
2. Update cluster infrastructure/control plane ready status
3. Let Control Plane controller handle phase transitions

## Best Practices

### 1. Infrastructure Provider Implementation

```go
// Infrastructure provider should focus on its own status
func (r *CAPTClusterReconciler) updateStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
    // Update infrastructure status
    cluster.Status.InfrastructureReady = captCluster.Status.Ready

    // Update other infrastructure-specific status
    if len(captCluster.Status.FailureDomains) > 0 {
        cluster.Status.FailureDomains = captCluster.Status.FailureDomains
    }

    // Do not update phase
    logger.Info("Current cluster phase", "phase", cluster.Status.Phase)
}
```

### 2. Control Plane Implementation

```go
// Control plane controller manages phase transitions
func (r *CAPTControlPlaneReconciler) updateStatus(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) error {
    // Update control plane status
    cluster.Status.ControlPlaneReady = controlPlane.Status.Ready

    // Manage phase transitions
    if !cluster.Status.ControlPlaneReady {
        cluster.Status.Phase = string(clusterv1.ClusterPhaseProvisioning)
    } else if cluster.Status.ControlPlaneReady && cluster.Status.InfrastructureReady {
        cluster.Status.Phase = string(clusterv1.ClusterPhaseProvisioned)
    }
}
```

## Important Considerations

1. **Status Update Timing**
   - Always update provider status before cluster status
   - Use patch operations for cluster status updates
   - Consider using server-side apply for critical updates

2. **Phase Management**
   - Infrastructure provider should not modify cluster phase
   - Control plane controller is responsible for phase transitions
   - Phase updates should be based on both infrastructure and control plane readiness

3. **Error Handling**
   - Infrastructure provider should focus on reporting its own errors
   - Control plane controller should aggregate status from all components
   - Use appropriate error types and conditions for status reporting

## Future Improvements

1. **Status Conditions**
   - Add more detailed status conditions for better observability
   - Implement condition aggregation in control plane controller
   - Consider adding phase transition conditions

2. **Phase Management**
   - Consider implementing more granular phase transitions
   - Add support for degraded states
   - Improve phase transition logging for better debugging

3. **Reconciliation**
   - Optimize status update frequency
   - Implement better conflict resolution
   - Add metrics for phase transitions
