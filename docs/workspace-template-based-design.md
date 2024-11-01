# WorkspaceTemplate Based Design

## Overview

This document describes the new design approach using WorkspaceTemplate for managing EKS clusters. The design focuses on using Terraform workspaces through Crossplane's Terraform Provider to manage infrastructure components.

## Key Components

### 1. WorkspaceTemplate
- Defines infrastructure as code using Terraform
- Supports metadata, versioning, and tagging
- Manages connection secrets for secure access
- Enables reusable infrastructure templates

### 2. VPC Management
```yaml
# VPC WorkspaceTemplate outputs
output "vpc_id" {
  value     = module.vpc.vpc_id
  sensitive = true
}

output "private_subnets" {
  value     = module.vpc.private_subnets
  sensitive = true
}
```

### 3. EKS Control Plane
```yaml
# EKS Control Plane vars reference
vars:
  - key: vpc_id
    secretKeyRef:
      name: vpc-connection
      key: vpc_id
  - key: private_subnets
    secretKeyRef:
      name: vpc-connection
      key: private_subnets
```

## Design Principles

### 1. Separation of Concerns
- VPC and EKS Control Plane are managed as separate workspaces
- Each component has its own lifecycle and state management
- Clear boundaries between infrastructure components

### 2. Configuration Management
- Use of Terraform outputs as Kubernetes secrets
- Direct secret references in dependent workspaces
- Type-safe variable passing between components

### 3. Dependency Management
- Explicit dependency declaration through WaitForWorkspaces
- Automatic secret synchronization
- Proper ordering of resource creation

## Implementation Details

### 1. VPC Configuration
```yaml
spec:
  template:
    spec:
      forProvider:
        source: Inline
        module: |
          module "vpc" {
            source = "terraform-aws-modules/vpc/aws"
            # VPC configuration
          }
      writeConnectionSecretToRef:
        name: vpc-connection
        namespace: default
```

### 2. EKS Control Plane Configuration
```yaml
spec:
  template:
    spec:
      forProvider:
        source: Inline
        vars:
          - key: vpc_id
            secretKeyRef:
              name: vpc-connection
              key: vpc_id
        module: |
          module "eks" {
            source = "terraform-aws-modules/eks/aws"
            # EKS configuration
          }
```

### 3. Dependency Flow
1. VPC WorkspaceTemplate creates infrastructure
2. Outputs are stored in Kubernetes secrets
3. EKS WorkspaceTemplate references VPC secrets
4. WorkspaceTemplateApply manages the lifecycle

## Benefits

1. Infrastructure Management
- Terraform-based resource provisioning
- State tracking and drift detection
- Secure secret management

2. Operational Benefits
- Simplified dependency management
- Clear resource ownership
- Automated secret handling

3. Development Benefits
- Reusable infrastructure templates
- Type-safe configuration
- Clear component boundaries

## Migration Impact

### Removed Components
- CAPTMachineTemplate (replaced by Karpenter)
- CAPTVPCTemplate (replaced by WorkspaceTemplate)
- Legacy CRDs and sample files

### New Components
- WorkspaceTemplate based VPC management
- Secret-based configuration sharing
- Integrated EKS Control Plane management

## Usage Example

1. Create VPC:
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: demo-vpc-apply
spec:
  templateRef:
    name: vpc-template
  variables:
    name: demo-cluster-vpc
```

2. Create EKS Control Plane:
```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: CAPTControlPlane
metadata:
  name: demo-cluster
spec:
  version: "1.31"
  workspaceTemplateRef:
    name: eks-controlplane-template
```

## Best Practices

1. Resource Organization
- Keep related resources in the same namespace
- Use consistent naming across resources
- Maintain clear dependencies

2. Secret Management
- Use sensitive outputs for secure values
- Reference secrets directly in dependent resources
- Maintain proper secret lifecycle

3. Dependency Management
- Explicitly declare dependencies
- Use WaitForWorkspaces feature
- Monitor resource readiness

## References

- [WorkspaceTemplate API](../api/v1beta1/workspacetemplate_types.go)
- [CAPTControlPlane Implementation](../api/controlplane/v1beta1/captcontrolplane_types.go)
- [VPC Configuration Sample](../config/samples/cluster/vpc.yaml)
- [EKS Control Plane Sample](../config/samples/cluster/controlplane.yaml)
