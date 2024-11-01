# Controlplane Workspace Template Migration Guide

## Overview

This document describes the migration of CAPT Controlplane from its current CRD-based implementation to a WorkspaceTemplate-based approach. This migration aligns with the existing infrastructure migration strategy and provides a more consistent way to manage EKS cluster resources.

## Current Implementation

The current implementation uses `CAPTControlPlane` CRD with the following key components:

- `CAPTControlPlaneSpec`: Defines the desired state of the control plane
  - `Version`: Kubernetes version
  - `MachineTemplate`: Template for creating control plane instances

## Target Implementation

The new implementation will use WorkspaceTemplate to manage the control plane configuration through Terraform. This approach offers several benefits:

1. Consistent management with infrastructure components
2. Direct integration with AWS APIs through Terraform
3. Better state management and drift detection
4. Simplified dependency management

## Implementation Steps

### 1. Create Control Plane WorkspaceTemplate

Create a new WorkspaceTemplate that defines the EKS control plane configuration. The template includes:

- EKS cluster configuration with Fargate profiles
- EKS Blueprints Addons (CoreDNS, VPC-CNI, Kube-Proxy)
- Karpenter configuration
- IAM roles and access entries

Example WorkspaceTemplate is available at: `config/samples/infrastructure_v1beta1_workspacetemplate_controlplane.yaml`

Key features of the template:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: eks-controlplane-template
spec:
  template:
    spec:
      forProvider:
        source: Inline
        module: |
          module "eks" {
            source  = "terraform-aws-modules/eks/aws"
            version = "~> 20.11"
            # ... EKS configuration
          }

          module "eks_blueprints_addons" {
            source  = "aws-ia/eks-blueprints-addons/aws"
            # ... Addons configuration
          }
```

### 2. Apply the WorkspaceTemplate

Use WorkspaceTemplateApply to create an instance of the control plane, with dependencies on VPC resources:

Example available at: `config/samples/infrastructure_v1beta1_workspacetemplateapply_controlplane.yaml`

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: eks-controlplane-apply
spec:
  templateRef:
    name: eks-controlplane-template
  variables:
    cluster_name: eks-karpenter-demo
    kubernetes_version: "1.31"
    vpc_id: ${dependencies.vpc-apply.outputs.vpc_id}
    private_subnet_ids: ${dependencies.vpc-apply.outputs.private_subnets}
  dependencies:
    - name: vpc-apply
      namespace: default
```

### 3. Update Controllers

1. Remove the existing `CAPTControlPlane` controller
2. Implement a new controller that:
   - Watches WorkspaceTemplate resources
   - Manages the lifecycle of EKS control plane
   - Handles updates and deletions
   - Manages dependencies with other components

### 4. Implementation Process

For the initial implementation:

1. Create the WorkspaceTemplate CRD for control plane management
2. Implement the controller logic for handling WorkspaceTemplate resources
3. Integrate with existing infrastructure components
4. Implement proper status reporting and error handling

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
2. All control plane operations (create, update, delete) work as expected
3. Fargate profiles are properly configured
4. EKS Blueprints addons are successfully deployed
5. Karpenter is operational
6. Monitoring and logging show expected behavior
7. All tests pass with the new implementation

## Implementation Timeline

1. Development Phase:
   - Create WorkspaceTemplate implementation
   - Develop new controller
   - Write tests

2. Testing Phase:
   - Run unit and integration tests
   - Validate functionality
   - Document any issues or limitations

3. Documentation Phase:
   - Update user documentation
   - Create example templates
   - Document best practices

## Migration Process

1. Preparation:
   - Deploy new WorkspaceTemplate CRDs
   - Create WorkspaceTemplates for control plane
   - Verify template validity

2. Migration:
   - For each existing cluster:
     1. Create WorkspaceTemplateApply for control plane
     2. Verify new control plane is operational
     3. Verify Fargate profiles and addons
     4. Remove old CAPTControlPlane resources

3. Validation:
   - Verify cluster functionality
   - Check monitoring and logging
   - Validate access and security
   - Test Fargate workload scheduling
   - Verify Karpenter operation

## References

- [WorkspaceTemplate Controlplane Sample](../config/samples/infrastructure_v1beta1_workspacetemplate_controlplane.yaml)
- [WorkspaceTemplateApply Controlplane Sample](../config/samples/infrastructure_v1beta1_workspacetemplateapply_controlplane.yaml)
- [WorkspaceTemplate API Specification](../api/v1beta1/workspacetemplate_types.go)
- [Current CAPTControlPlane Implementation](../api/controlplane/v1beta1/captcontrolplane_types.go)
