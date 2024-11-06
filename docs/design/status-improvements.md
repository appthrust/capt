# Status Management Improvements

## Overview

This document describes improvements made to the status management in both CAPTCluster and CAPTControlPlane controllers to better align with Cluster API requirements and improve observability.

## CAPTCluster Status Improvements

### Conditions Management

The CAPTCluster controller now properly manages two key conditions:

1. VPCReady (existing)
   - Indicates the readiness state of the VPC infrastructure
   - Maintained for backward compatibility

2. InfrastructureReady (new)
   - Aligns with Cluster API's standard condition types
   - Set to "True" when the infrastructure is fully ready
   - Example:
     ```yaml
     conditions:
       - type: InfrastructureReady
         status: "True"
         reason: "InfrastructureReady"
         message: "Infrastructure is ready"
     ```

This improvement ensures better integration with Cluster API's status management patterns and provides clearer infrastructure state visibility.

## CAPTControlPlane Status Improvements

### Control Plane Endpoint Management

The CAPTControlPlane controller has been improved to properly manage the control plane endpoint:

1. Endpoint Source
   - Now retrieves endpoint information from WorkspaceTemplateApply outputs
   - Specifically looks for the "endpoint" output from the EKS workspace

2. Status Flow
   ```
   WorkspaceTemplateApply Outputs
           ↓
   CAPTControlPlane Status
           ↓
   Cluster Spec
   ```

3. Implementation Details
   ```go
   // Get endpoint information from WorkspaceTemplateApply outputs
   if endpoint, ok := workspaceApply.Status.Outputs["endpoint"]; ok {
       controlPlane.Status.ControlPlaneEndpoint = clusterv1.APIEndpoint{
           Host: endpoint,
           Port: 443, // EKS API server always uses port 443
       }
   }
   ```

4. Result Format
   ```yaml
   spec:
     controlPlaneEndpoint:
       host: "your-cluster-endpoint.eks.amazonaws.com"
       port: 443
   ```

### Benefits

1. Reliability
   - Endpoint information is now sourced directly from the infrastructure provider
   - Eliminates potential mismatches between actual and reported endpoints

2. Consistency
   - Ensures the control plane endpoint is always set with valid values
   - Maintains proper port configuration (443 for EKS)

3. Observability
   - Better logging of endpoint updates
   - Clear status propagation chain

## Integration with Cluster API

These improvements ensure better integration with Cluster API's expectations:

1. Status Propagation
   - Infrastructure readiness properly reflected in conditions
   - Control plane endpoint correctly propagated to Cluster resource

2. Condition Management
   - Standard condition types used where appropriate
   - Clear status transitions and reason messages

3. Phase Management
   - Proper phase transitions based on infrastructure and control plane readiness
   - Clear relationship between conditions and phases
