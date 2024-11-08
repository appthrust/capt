# VPC WorkspaceTemplate Design

This document describes the design of VPC configuration using WorkspaceTemplate in CAPT.

## Overview

CAPT uses WorkspaceTemplate to manage VPC configurations, providing a flexible and reusable approach to VPC management. This design allows for both creating new VPCs and using existing VPCs.

## Architecture

### Components

1. WorkspaceTemplate
   - Defines the VPC configuration template
   - Contains Terraform module configuration
   - Specifies connection secret reference

2. WorkspaceTemplateApply
   - Automatically created and managed by the CAPTCluster controller
   - Applies the VPC configuration
   - Manages the lifecycle of VPC resources
   - Handles dependencies and secrets
   - Name format: {cluster-name}-vpc

3. CAPTCluster
   - References VPC WorkspaceTemplate
   - Controller creates and manages WorkspaceTemplateApply
   - Tracks VPC status
   - Manages VPC lifecycle

### Resource Relationships

```
CAPTCluster
  ├── VPCTemplateRef ──> WorkspaceTemplate
  │                        └── WriteConnectionSecretToRef ──> Secret
  └── (Controller) ──> WorkspaceTemplateApply
                        └── TemplateRef ──> WorkspaceTemplate
```

## VPC Configuration

### Creating New VPC

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: sample-cluster
spec:
  region: us-west-2
  vpcTemplateRef:
    name: vpc-template
    namespace: default
```

Note: The controller will automatically create a WorkspaceTemplateApply named `{cluster-name}-vpc`

### Using Existing VPC

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: sample-cluster
spec:
  region: us-west-2
  existingVpcId: vpc-0123456789abcdef0
```

## State Management

### CAPTCluster Status

The CAPTCluster maintains the following status information:
- VPC ID
- VPC workspace name
- Ready state
- Conditions

### Conditions

VPC readiness is tracked using a single condition:
- Type: VPCReady
- Status: True/False
- Reason: Various reasons indicating the current state
- Message: Detailed information about the current state

### State Transitions

1. Initial State
   - Condition: VPCReady = False
   - Reason: VPCCreating
   - Message: "Initializing VPC creation"

2. WorkspaceTemplateApply Progress
   - Condition: VPCReady = False
   - Reason: VPCCreating
   - Message: From WorkspaceTemplateApply conditions

3. Final State
   - Condition: VPCReady = True
   - Reason: VPCCreated/ExistingVPCUsed
   - Message: Success message

## Implementation Details

### WorkspaceTemplate Readiness

The controller checks two conditions from WorkspaceTemplateApply:
- TypeSynced: Indicates successful Terraform sync
- TypeReady: Indicates resource availability

Both conditions must be true for the VPC to be considered ready.

### Secret Management

VPC information is stored in a connection secret:
- Created by WorkspaceTemplate
- Contains VPC ID and other VPC details
- Referenced by CAPTCluster for VPC information

### Retry Behavior

The controller implements a simple retry mechanism:
1. Requeues every 10 seconds
2. Checks for resource availability
3. No explicit timeout (TODO)

## Future Improvements

1. Timeout Implementation
   - Add overall VPC creation timeout
   - Add secret availability timeout
   - Add workspace readiness timeout

2. Error Handling
   - Improve error categorization
   - Add specific error conditions
   - Implement recovery mechanisms

3. Resource Cleanup
   - Implement proper finalizers
   - Handle resource dependencies
   - Clean up orphaned resources

## Usage Examples

### Standard VPC Template

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: vpc-template
spec:
  template:
    metadata:
      description: "Standard VPC configuration"
    spec:
      module:
        source: "terraform-aws-modules/vpc/aws"
        version: "5.0.0"
      variables:
        name:
          value: "${var.name}"
        cidr:
          value: "10.0.0.0/16"
  writeConnectionSecretToRef:
    name: vpc-connection
    namespace: default
```

### Using the Template in CAPTCluster

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: sample-cluster
spec:
  region: us-west-2
  vpcTemplateRef:
    name: vpc-template
    namespace: default
```

Note: Do not create WorkspaceTemplateApply manually, as it is managed by the controller.

## Migration Guide

For migrating from the old VPCConfig approach:

1. Create WorkspaceTemplate for your VPC configuration
2. Update CAPTCluster to use VPCTemplateRef
3. Remove old VPCConfig
4. Let the controller manage WorkspaceTemplateApply
5. Apply changes and monitor status

## Best Practices

1. Template Management
   - Use standardized templates
   - Version templates appropriately
   - Document template configurations

2. Status Monitoring
   - Monitor VPCReady condition
   - Check WorkspaceTemplateApply status
   - Verify connection secret availability

3. Error Handling
   - Handle transient failures
   - Implement proper retries
   - Log relevant error information

4. WorkspaceTemplateApply
   - Never create WorkspaceTemplateApply manually
   - Let the controller manage the lifecycle
   - Monitor through CAPTCluster status
