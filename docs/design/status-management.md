# Status Management Design

## Overview

This document describes the status management implementation for CaptCluster and CAPTControlPlane resources in the context of Cluster API integration. The design ensures proper status propagation from infrastructure and control plane providers to the Cluster API's Cluster resource.

## Requirements

### Infrastructure Provider Requirements

As per Cluster API specifications, an infrastructure provider must:

1. Provide status fields:
   - Required:
     - `controlPlaneEndpoint` - API server endpoint
     - `ready` - Infrastructure readiness state
   - Optional:
     - `failureReason` - Programmatic error token
     - `failureMessage` - Human-readable error message
     - `failureDomains` - Available failure domains for machine placement

2. Handle owner references:
   - Wait for owner reference from Cluster controller
   - Take no action until owner reference is set

3. Handle control plane endpoint:
   - Either provide own endpoint or use Cluster's endpoint
   - Exit reconciliation if endpoint not available

### Control Plane Provider Requirements

The control plane provider must:

1. Manage control plane endpoint:
   - Provide or consume endpoint information
   - Ensure endpoint availability for cluster operations

2. Maintain status:
   - Ready state
   - Initialization state
   - Error conditions

## Implementation Details

### CaptCluster Status Management

```go
type CAPTClusterStatus struct {
    // VPC-specific status
    VPCWorkspaceName string
    VPCID           string
    
    // Cluster API required fields
    Ready           bool
    FailureReason   *string
    FailureMessage  *string
    FailureDomains  clusterv1.FailureDomains
    
    // Conditions for detailed state tracking
    Conditions      []metav1.Condition
}
```

Key implementation points:

1. VPC Status Tracking:
   - Tracks VPC creation through WorkspaceTemplateApply
   - Monitors workspace conditions (Synced and Ready)
   - Extracts VPC ID from connection secret

2. Error Handling:
   - Detailed error conditions with reasons and messages
   - Proper error propagation to Cluster resource
   - Terminal error states handling

3. Status Propagation:
   ```go
   func (r *CAPTClusterReconciler) updateClusterStatus(ctx context.Context, cluster *clusterv1.Cluster, captCluster *infrastructurev1beta1.CAPTCluster) error {
       cluster.Status.InfrastructureReady = captCluster.Status.Ready
       if captCluster.Status.FailureReason != nil {
           reason := capierrors.ClusterStatusError(*captCluster.Status.FailureReason)
           cluster.Status.FailureReason = &reason
       }
       // ...
   }
   ```

### CAPTControlPlane Status Management

```go
type CAPTControlPlaneStatus struct {
    Ready                  bool
    Initialized           bool
    Phase                 string
    FailureReason         *string
    FailureMessage        *string
    WorkspaceTemplateStatus *WorkspaceTemplateStatus
    Conditions            []metav1.Condition
}
```

Key implementation points:

1. Phase Management:
   - Clear phase transitions: Creating → Ready/Failed
   - Phase-specific conditions and messages
   - Timeout handling for each phase

2. Workspace Integration:
   ```go
   type WorkspaceTemplateStatus struct {
       Ready              bool
       State             string
       LastAppliedRevision string
       LastFailedRevision string
       LastFailureMessage string
       Outputs           map[string]string
   }
   ```

3. Condition Types:
   ```go
   const (
       ControlPlaneReadyCondition = "Ready"
       ControlPlaneInitializedCondition = "Initialized"
       ControlPlaneFailedCondition = "Failed"
       ControlPlaneCreatingCondition = "Creating"
   )
   ```

## Timeout Handling

Both controllers implement timeout handling for critical operations:

1. CaptCluster:
   - VPC creation timeout
   - Secret availability timeout
   - Workspace ready timeout

2. CAPTControlPlane:
   - Control plane creation timeout
   - VPC ready timeout
   - Workspace ready timeout

Example:
```go
const (
    controlPlaneTimeout = 30 * time.Minute
    vpcReadyTimeout    = 15 * time.Minute
    secretTimeout      = 5 * time.Minute
)
```

## Error Recovery

The implementation includes robust error recovery mechanisms:

1. Transient Errors:
   - Automatic retries with backoff
   - Clear error conditions
   - Status preservation during recovery

2. Terminal Errors:
   - Clear failure indication
   - Detailed error messages
   - No automatic recovery (requires manual intervention)

## Status Visibility

Enhanced status visibility through kubectl:

1. CaptCluster:
   ```bash
   $ kubectl get captcluster
   NAME            VPC-ID          READY   ENDPOINT        AGE
   test-cluster    vpc-123456789   true    10.0.0.1:6443   10m
   ```

2. CAPTControlPlane:
   ```bash
   $ kubectl get captcontrolplane
   NAME            READY   PHASE      VERSION   ENDPOINT        AGE
   test-cluster    true    Ready      1.31      10.0.0.1:6443   10m
   ```

## Best Practices

1. Status Updates:
   - Atomic updates to avoid race conditions
   - Proper error handling during updates
   - Clear status transitions

2. Error Handling:
   - Detailed error messages
   - Proper error categorization
   - Clear recovery paths

3. Condition Management:
   - Use standard condition types
   - Include transition timestamps
   - Provide clear messages

## Testing Considerations

1. Status Transitions:
   - Test all possible state transitions
   - Verify timeout handling
   - Check error recovery paths

2. Integration Testing:
   - Verify status propagation to Cluster
   - Test owner reference handling
   - Validate endpoint management

3. Error Scenarios:
   - Test various error conditions
   - Verify error message propagation
   - Check recovery mechanisms

## Future Improvements

1. Enhanced Status:
   - More detailed progress information
   - Resource usage metrics
   - Health indicators

2. Error Handling:
   - Automated recovery for certain errors
   - More detailed error categorization
   - Better error correlation

3. Monitoring:
   - Status-based alerts
   - Performance metrics
   - Resource health monitoring
