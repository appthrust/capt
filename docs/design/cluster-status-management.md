# Cluster Status Management Design

## Overview

This document describes the design and implementation details for managing Cluster status in the CAPT provider.

## Background

ClusterAPI requires infrastructure providers to manage the status of both their own resources (CAPTCluster, CAPTControlPlane) and the core Cluster resource. This includes:

1. Infrastructure readiness
2. Control plane readiness
3. Cluster phase transitions
4. Status conditions and error handling

## Requirements from ClusterAPI

### Infrastructure Provider Requirements

1. **Owner References**
   - Infrastructure provider must set owner references on infrastructure objects
   - The Cluster controller sets owner references on infrastructure objects referenced in `Cluster.spec.infrastructureRef`

2. **Status Management**
   - Must provide a status object with required fields:
     - `ready` - boolean indicating if infrastructure is ready
     - `controlPlaneEndpoint` - endpoint for connecting to the cluster's API server

3. **Optional Status Fields**
   - `failureReason` - string explaining why a fatal error occurred
   - `failureMessage` - detailed error message
   - `failureDomains` - map of failure domains for machine placement

### Cluster Phase Management

The Cluster goes through several phases:
1. `Provisioning` - Initial state when infrastructure and control plane are being prepared
2. `Provisioned` - Infrastructure and control plane are ready
3. `Running` - Cluster is fully operational

Phase transitions are determined by:
- `InfrastructureReady` status from CAPTCluster
- `ControlPlaneReady` status from CAPTControlPlane

## Implementation Details

### CAPTCluster Controller

1. **Status Updates**
   ```go
   func (r *CAPTClusterReconciler) updateStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
       // Update CAPTCluster status first
       if err := r.Status().Update(ctx, captCluster); err != nil {
           return err
       }

       // Update Cluster status if it exists
       if cluster != nil {
           patch := client.MergeFrom(cluster.DeepCopy())
           cluster.Status.InfrastructureReady = captCluster.Status.Ready
           
           // Update phase if both infrastructure and control plane are ready
           if cluster.Status.InfrastructureReady && cluster.Status.ControlPlaneReady {
               cluster.Status.Phase = string(clusterv1.ClusterPhaseProvisioned)
           }

           if err := r.Status().Patch(ctx, cluster, patch); err != nil {
               return err
           }
       }

       return nil
   }
   ```

2. **Owner References**
   ```go
   func (r *CAPTClusterReconciler) setOwnerReference(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
       if cluster == nil {
           return nil
       }

       // Check if owner reference is already set
       for _, ref := range captCluster.OwnerReferences {
           if ref.Kind == "Cluster" && ref.APIVersion == clusterv1.GroupVersion.String() {
               return nil
           }
       }

       return controllerutil.SetControllerReference(cluster, captCluster, r.Scheme)
   }
   ```

### CAPTControlPlane Controller

Similar to CAPTCluster, the CAPTControlPlane controller:
1. Updates its own status
2. Updates the Cluster's ControlPlaneReady status
3. Provides control plane endpoint information

## Important Considerations

1. **Status Update Order**
   - Always update the provider's status (CAPTCluster/CAPTControlPlane) before updating Cluster status
   - Use server-side apply (patch) for Cluster status updates to avoid conflicts

2. **Phase Transitions**
   - Phase transitions should be deterministic and based on clear conditions
   - Both infrastructure and control plane must be ready before moving to Provisioned phase

3. **Error Handling**
   - Propagate detailed error information through status conditions
   - Use appropriate error types from ClusterAPI (e.g., `ClusterStatusError`)

4. **Reconciliation**
   - Avoid unnecessary status updates to prevent reconciliation loops
   - Use patch operations instead of updates when possible
   - Consider using generation/observedGeneration for change detection

## Future Improvements

1. **Status Conditions**
   - Implement more detailed status conditions for better observability
   - Add conditions for different stages of infrastructure provisioning

2. **Failure Domain Support**
   - Add support for AWS availability zones as failure domains
   - Implement failure domain-aware machine placement

3. **Phase Management**
   - Add support for more granular phase transitions
   - Implement better handling of degraded states

4. **Status Updates**
   - Optimize status update frequency
   - Implement better conflict resolution for status updates
