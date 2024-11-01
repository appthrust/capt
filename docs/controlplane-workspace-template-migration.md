# Controlplane Workspace Template Migration Guide

## Overview

This document describes the migration of CAPT Controlplane from its current CRD-based implementation to a WorkspaceTemplate-based approach. This migration aligns with the existing infrastructure migration strategy and provides a more consistent way to manage EKS cluster resources.

## Current Implementation

The current implementation uses `CAPTControlPlane` CRD with the following key components:

- `CAPTControlPlaneSpec`: Defines the desired state of the control plane
  - `Version`: Kubernetes version
  - `MachineTemplate`: Template for creating control plane instances

## Target Implementation

The new implementation uses WorkspaceTemplate to manage the control plane configuration through Terraform. This approach offers several benefits:

1. Consistent management with infrastructure components
2. Direct integration with AWS APIs through Terraform
3. Better state management and drift detection
4. Simplified dependency management

## Implementation Details

### 1. CAPTControlPlane Changes

The CAPTControlPlane CRD has been updated to use WorkspaceTemplate:

```yaml
spec:
  version: "1.31"
  workspaceTemplateRef:
    name: eks-controlplane-template
    namespace: default
  controlPlaneConfig:
    endpointAccess:
      public: true
      private: true
    fargateProfiles:
      - name: kube-system
        selectors:
          - namespace: kube-system
```

Key changes:
- Removed MachineTemplate in favor of WorkspaceTemplateRef
- Added ControlPlaneConfig for EKS-specific settings
- Added support for Fargate profiles and addons

### 2. WorkspaceTemplate Implementation

The EKS control plane is managed through a WorkspaceTemplate that includes:

1. EKS Cluster Configuration:
   - Fargate profiles
   - Security settings
   - Network configuration

2. EKS Blueprints Addons:
   - CoreDNS with Fargate optimization
   - VPC-CNI
   - Kube-proxy
   - Karpenter

3. Access Management:
   - IAM roles and policies
   - Karpenter node access

For detailed implementation, see: [EKS Control Plane Template](../config/samples/cluster/controlplane.yaml)

### 3. Controller Implementation

The controller has been updated to:
- Watch WorkspaceTemplate resources
- Manage WorkspaceTemplateApply resources
- Handle lifecycle operations
- Update status based on WorkspaceTemplate state

## Sample Implementation

The complete implementation is organized into three main components:

1. [VPC Infrastructure](../config/samples/cluster/vpc.yaml):
   - VPC WorkspaceTemplate
   - Network configuration
   - Subnet tagging for EKS

2. [Control Plane](../config/samples/cluster/controlplane.yaml):
   - EKS WorkspaceTemplate
   - Fargate profiles
   - EKS Blueprints addons

3. [Cluster Configuration](../config/samples/cluster/cluster.yaml):
   - CAPTCluster
   - CAPI Cluster
   - Resource references

For detailed design documentation, see: [EKS Cluster Design](eks-cluster-design.md)

## Migration Process

1. Preparation:
   - Deploy new CRDs with WorkspaceTemplate support
   - Create WorkspaceTemplates for control plane
   - Verify template validity

2. Migration Steps:
   - For each existing cluster:
     1. Create VPC WorkspaceTemplate and apply
     2. Create EKS WorkspaceTemplate and apply
     3. Update CAPTControlPlane to use WorkspaceTemplate
     4. Verify cluster functionality

3. Validation:
   - Verify control plane operations
   - Check Fargate profile creation
   - Validate addon deployment
   - Test cluster access

## Testing Strategy

1. Unit Tests:
   - Test WorkspaceTemplate validation
   - Test controller logic
   - Test Terraform configurations

2. Integration Tests:
   - Test complete cluster creation
   - Test control plane updates
   - Test failure scenarios
   - Verify Fargate profile creation
   - Validate addon deployment

## Success Criteria

The implementation will be considered successful when:

1. Control plane can be created using WorkspaceTemplate
2. All control plane operations work as expected:
   - Creation
   - Updates
   - Deletion
3. Fargate profiles are properly configured
4. EKS Blueprints addons are successfully deployed
5. Karpenter is operational
6. Monitoring and logging show expected behavior
7. All tests pass

## Implementation Timeline

1. Development Phase:
   - Update CAPTControlPlane CRD
   - Implement WorkspaceTemplate support
   - Update controller logic
   - Create sample configurations

2. Testing Phase:
   - Run unit tests
   - Perform integration testing
   - Validate migration process
   - Document any issues

3. Documentation Phase:
   - Update API documentation
   - Create migration guides
   - Document best practices
   - Create example templates

## References

- [EKS Cluster Design](eks-cluster-design.md)
- [VPC Template Sample](../config/samples/cluster/vpc.yaml)
- [Control Plane Template Sample](../config/samples/cluster/controlplane.yaml)
- [Cluster Configuration Sample](../config/samples/cluster/cluster.yaml)
- [WorkspaceTemplate API Specification](../api/v1beta1/workspacetemplate_types.go)
- [CAPTControlPlane Implementation](../api/controlplane/v1beta1/captcontrolplane_types.go)
