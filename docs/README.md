# Cluster API Terraform Provider

## Overview

The Cluster API Terraform Provider is a tool for declaratively managing Kubernetes cluster infrastructure using Terraform. This provider streamlines the construction, management, and operation of infrastructure, providing cluster resources in a consistent manner.

## Core Components

### 1. Infrastructure Management (CAPTCluster)
- Manages core infrastructure components like VPC
- Supports VPC retention functionality for shared infrastructure scenarios
- Provides comprehensive status management and tracking

### 2. Control Plane Management (CAPTControlPlane)
- Manages EKS control plane configuration
- Handles control plane endpoint management
- Integrates with EKS addons and configurations

### 3. Compute Resource Management (Machine)
- Manages node groups and Fargate profiles
- Supports flexible scaling configurations
- Enables independent lifecycle management of compute resources

### 4. Template Management (WorkspaceTemplate)
- Provides infrastructure as code capabilities
- Enables reusable and modular infrastructure definitions
- Supports version control and configuration management

## Key Benefits

### 1. Declarative Infrastructure Management

Use WorkspaceTemplate to manage infrastructure as code:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: vpc-template
spec:
  template:
    metadata:
      description: "Template for creating AWS VPC"
      version: "1.0.0"
    spec:
      forProvider:
        source: Inline
        module: |
          module "vpc" {
            source = "terraform-aws-modules/vpc/aws"
            # VPC configuration
          }
```

- Clear configuration management through versioning and tagging
- Configuration drift detection through state tracking
- Utilization of standard Terraform modules

### 2. Robust Dependency Management

Explicitly define and safely manage dependencies between components:

```yaml
spec:
  waitForSecret:
    name: vpc-connection
    namespace: default
```

- Explicit dependency definition between components like VPC and EKS
- Secure configuration propagation through secrets
- Independent lifecycle management for each component

### 3. Secure Configuration Management

Provides security-focused configuration management features:

- Secure management of sensitive information using Kubernetes secrets
- Automatic configuration of OIDC authentication and IAM roles
- Centralized management of security groups and network policies
- Secure configuration migration between environments

### 4. High Operability and Reusability

Enables efficient operations and configuration reuse:

```yaml
spec:
  template:
    metadata:
      tags:
        environment: "dev"
        provider: "aws"
    spec:
      forProvider:
        vars:
          - key: cluster_name
            value: "demo-cluster"
```

- Reusable infrastructure templates
- Customization through environment-specific variables and tags
- Automatic management of Helm charts and EKS addons
- Compatibility with existing Terraform modules

### 5. Integration with Modern Kubernetes Features

Easily integrate with the latest Kubernetes features:

- Automatic Fargate profile configuration
- Efficient node scaling with Karpenter
- Integrated management of EKS addons
- Extensibility through Custom Resource Definitions (CRDs)

### 6. Advanced Infrastructure Management

Provides sophisticated infrastructure management capabilities:

- VPC retention functionality for shared infrastructure scenarios
- Comprehensive status management and monitoring
- Detailed lifecycle management for all components
- Robust error handling and recovery mechanisms

## Usage Examples

1. Creating a VPC with Retention:
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: demo-vpc-apply
spec:
  templateRef:
    name: vpc-template
  retainWorkspaceOnDelete: true  # Retain VPC after deletion
  variables:
    name: demo-cluster-vpc
```

2. Creating an EKS Cluster:
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

3. Managing Compute Resources:
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTMachineDeployment
metadata:
  name: demo-nodegroup
spec:
  template:
    spec:
      workspaceTemplateRef:
        name: nodegroup-template
```

## Best Practices

1. Resource Management
- Manage related resources in the same namespace
- Use consistent naming conventions
- Define clear dependencies
- Implement proper status monitoring

2. Security
- Manage sensitive information as secrets
- Configure IAM following the principle of least privilege
- Properly configure security groups
- Regularly rotate credentials and certificates

3. Operations Management
- Separate configurations by environment
- Utilize version control
- Regularly check for configuration drift
- Monitor cluster and component status
- Implement proper backup and disaster recovery procedures

4. Infrastructure Planning
- Consider VPC retention requirements for shared infrastructure
- Plan for scalability and high availability
- Design for multi-environment deployments
- Implement proper resource tagging strategy
