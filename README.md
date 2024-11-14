# CAPT (Cluster API Provider Terraform)

CAPT is a Cluster API provider that leverages Terraform to create and manage EKS clusters on AWS. It uses Crossplane's Terraform Provider to manage infrastructure components through Kubernetes-native resources.

## Overview

CAPT implements a modular approach to EKS cluster management where each infrastructure component (VPC, Control Plane, Machine Resources) is managed through its own WorkspaceTemplate. This design enables:

- Clear separation of concerns between infrastructure components
- Reusable infrastructure templates
- Secure configuration management through Kubernetes secrets
- Terraform-based state management and drift detection
- ClusterClass support for standardized cluster deployments
- Independent compute resource management through Machine concept

## Architecture

The cluster creation is divided into four main components:

1. VPC Infrastructure
2. EKS Control Plane
3. Compute Resources (Machine)
4. Cluster Configuration

Each component is managed independently through WorkspaceTemplates and can be templated using ClusterClass. The controllers automatically manage WorkspaceTemplateApply resources for infrastructure provisioning:

```mermaid
graph TD
    A[Cluster] --> B[CAPTCluster]
    A --> C[CAPTControlPlane]
    A --> D[CAPTMachineDeployment]
    B --> E[VPC WorkspaceTemplate]
    C --> F[EKS WorkspaceTemplate]
    D --> G[NodeGroup WorkspaceTemplate]
    B --> |Controller| H[VPC WorkspaceTemplateApply]
    C --> |Controller| I[EKS WorkspaceTemplateApply]
    D --> |Controller| J[NodeGroup WorkspaceTemplateApply]
    H --> E
    I --> F
    J --> G
    H --> K[VPC Infrastructure]
    I --> L[EKS Control Plane]
    I --> M[EKS Blueprints Addons]
    J --> N[Compute Resources]
    O[ClusterClass] --> A
    O --> P[CaptControlPlaneTemplate]
    P --> F
```

## Key Benefits

### 1. Declarative Infrastructure Management
- Version control and tagging for clear configuration management
- State tracking for configuration drift detection
- Utilization of standard Terraform modules
- ClusterClass templates for standardized deployments
- Automatic WorkspaceTemplateApply management by controllers
- VPC retention capability for shared infrastructure scenarios

### 2. Robust Dependency Management
- Explicit dependency definition between components (e.g., VPC and EKS)
- Secure configuration propagation through secrets
- Independent lifecycle management for each component
- Template-based configuration with variable substitution

### 3. Secure Configuration Management
- Secure handling of sensitive information through Kubernetes secrets
- Automatic OIDC authentication and IAM role configuration
- Centralized security group and network policy management
- Secure configuration migration between environments

### 4. High Operability and Reusability
- Reusable infrastructure templates
- Customization through environment-specific variables and tags
- Automatic management of Helm charts and EKS addons
- Compatibility with existing Terraform modules
- ClusterClass for consistent cluster deployments

### 5. Modern Kubernetes Feature Integration
- Automatic Fargate profile configuration
- Efficient node scaling with Karpenter
- Integrated EKS addon management
- Extensibility through Custom Resource Definitions (CRDs)
- ClusterTopology support for advanced cluster management

## Quick Start Guide

This guide will help you get started with CAPT, deploy it on your Kubernetes cluster, and set up a basic integration with Cluster API.

### Prerequisites

Before you begin, ensure you have the following:

1. A Kubernetes cluster (v1.19+)
2. kubectl installed and configured to access your cluster
3. Cluster API (v1.0+) installed on your cluster
4. AWS credentials with appropriate permissions
5. Crossplane with Terraform Provider installed

### Step 1: Install CAPT

1. Download the latest CAPT release:
   ```
   curl -LO https://github.com/appthrust/capt/releases/latest/download/capt.yaml
   ```

2. Install CAPT:
   ```
   kubectl apply -f capt.yaml
   ```

   Note: The `capt.yaml` file includes all necessary Custom Resource Definitions (CRDs), RBAC settings, and the CAPT controller deployment.

3. Verify the controller is running:
   ```
   kubectl get pods -n capt-system
   ```

Important: The default installation uses the `controller:latest` image tag. For production use, it's recommended to use a specific version tag. You can modify the image tag in the `capt.yaml` file before applying it.

### Step 2: Configure AWS Credentials

1. Create a Kubernetes secret with your AWS credentials:
   ```
   kubectl create secret generic aws-credentials \
     --from-literal=AWS_ACCESS_KEY_ID=<your-access-key> \
     --from-literal=AWS_SECRET_ACCESS_KEY=<your-secret-key> \
     -n capt-system
   ```

2. Apply the Crossplane Terraform Provider configuration:
   ```
   kubectl apply -f crossplane-terraform-config/provider-config.yaml
   ```

### Step 3: Create a Simple EKS Cluster

1. Create a VPC WorkspaceTemplate:
   ```yaml
   apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
   kind: WorkspaceTemplate
   metadata:
     name: simple-vpc
   spec:
     template:
       metadata:
         description: "Simple VPC configuration"
       spec:
         module:
           source: "terraform-aws-modules/vpc/aws"
           version: "5.0.0"
         variables:
           name:
             value: "simple-vpc"
           cidr:
             value: "10.0.0.0/16"
   ```
   Save this as `simple-vpc.yaml` and apply it:
   ```
   kubectl apply -f simple-vpc.yaml
   ```

2. Create a CAPTCluster resource:
   ```yaml
   apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
   kind: CAPTCluster
   metadata:
     name: simple-cluster
   spec:
     region: us-west-2
     vpcTemplateRef:
       name: simple-vpc
   ```
   Save this as `simple-cluster.yaml` and apply it:
   ```
   kubectl apply -f simple-cluster.yaml
   ```

3. Create a Cluster resource:
   ```yaml
   apiVersion: cluster.x-k8s.io/v1beta1
   kind: Cluster
   metadata:
     name: simple-cluster
   spec:
     infrastructureRef:
       apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
       kind: CAPTCluster
       name: simple-cluster
   ```
   Save this as `cluster.yaml` and apply it:
   ```
   kubectl apply -f cluster.yaml
   ```

### Step 4: Monitor Cluster Creation

1. Check the status of your cluster:
   ```
   kubectl get clusters
   ```

2. View the CAPTCluster resource:
   ```
   kubectl get captclusters
   ```

3. Check the WorkspaceTemplateApply resources:
   ```
   kubectl get workspacetemplateapplies
   ```

### Step 5: Access Your EKS Cluster

Once the cluster is ready:

1. Get the kubeconfig for your new EKS cluster:
   ```
   aws eks get-token --cluster-name simple-cluster > kubeconfig
   ```

2. Use the new kubeconfig to interact with your EKS cluster:
   ```
   kubectl --kubeconfig=./kubeconfig get nodes
   ```

## Usage

### 1. Using ClusterClass (Recommended)

ClusterClass provides a templated approach to cluster creation, enabling standardized deployments across your organization:

1. Define ClusterClass:
```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: ClusterClass
metadata:
  name: eks-class
spec:
  controlPlane:
    ref:
      apiVersion: controlplane.cluster.x-k8s.io/v1beta1
      kind: CaptControlPlaneTemplate
      name: eks-control-plane-template
  variables:
    - name: controlPlane.version
      required: true
      schema:
        openAPIV3Schema:
          type: string
          enum: ["1.27", "1.28", "1.29", "1.30", "1.31"]
```

2. Create Cluster using ClusterClass:
```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: demo-cluster
spec:
  topology:
    class: eks-class
    version: "1.31"
    variables:
      - name: controlPlane.version
        value: "1.31"
      - name: environment
        value: dev
```

### 2. Traditional Approach

#### Create VPC Infrastructure Template with Retention

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
```

#### Create CAPTCluster with VPC Retention

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: demo-cluster
spec:
  region: us-west-2
  vpcTemplateRef:
    name: vpc-template
    namespace: default
  retainVpcOnDelete: true  # VPC will be retained when cluster is deleted
```

#### Create Compute Resources (Machine)

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTMachineDeployment
metadata:
  name: demo-nodegroup
spec:
  replicas: 3
  template:
    spec:
      workspaceTemplateRef:
        name: nodegroup-template
      instanceType: t3.medium
      diskSize: 50
```

#### Create NodeGroup Template

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: nodegroup-template
spec:
  template:
    metadata:
      description: "EKS Node Group configuration"
    spec:
      module:
        source: "./internal/tf_module/eks_node_group"
      variables:
        instance_types:
          value: ["${var.instance_type}"]
        disk_size:
          value: "${var.disk_size}"
```

#### Apply Cluster Configuration

```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: demo-cluster
spec:
  clusterNetwork:
    services:
      cidrBlocks: ["10.96.0.0/12"]
    pods:
      cidrBlocks: ["192.168.0.0/16"]
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
    kind: CAPTCluster
    name: demo-cluster
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1beta1
    kind: CAPTControlPlane
    name: demo-cluster
```

Note: WorkspaceTemplateApply resources are automatically created and managed by the controllers. You do not need to create them manually.

## Best Practices

### 1. Resource Management
- Manage related resources in the same namespace
- Use consistent naming conventions
- Define clear dependencies between components
- Regular configuration drift checks
- Utilize ClusterClass for standardized deployments
- Let controllers manage WorkspaceTemplateApply resources

### 2. Security
- Manage sensitive information as secrets
- Follow the principle of least privilege for IAM configuration
- Proper security group configuration
- Implement secure network policies

### 3. Operations
- Separate configurations per environment
- Utilize version control effectively
- Monitor and manage component lifecycles
- Regular security and compliance audits
- Use ClusterClass for consistent deployments

### 4. Template Management
- Document template purposes and requirements
- Version templates appropriately
- Implement proper tagging strategies
- Maintain backward compatibility
- Leverage ClusterClass variables for flexibility
- Use WorkspaceTemplate for infrastructure definitions
- Let controllers handle WorkspaceTemplateApply lifecycle

## Features

### ClusterClass Support
- Standardized cluster templates
- Variable-based configuration
- Reusable control plane templates
- Consistent cluster deployments
- Environment-specific customization

### WorkspaceTemplate Management
- Infrastructure as code using Terraform
- Version control and metadata tracking
- Secure secret management
- Reusable infrastructure templates
- Automatic WorkspaceTemplateApply management by controllers

### Machine Management
- Independent compute resource lifecycle
- Flexible node group configuration
- Support for multiple instance types
- Automated scaling configuration
- Integration with cluster autoscaling
- Template-based node group management

### VPC Management
- Multi-AZ deployment
- Public and private subnets
- NAT Gateway configuration
- EKS and Karpenter integration
- VPC retention for shared infrastructure
- Independent VPC lifecycle management

### EKS Control Plane
- Fargate profiles for system workloads
- EKS Blueprints addons integration
- CoreDNS, VPC-CNI, and Kube-proxy configuration
- Karpenter setup for node management
- Template-based configuration with ClusterClass

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Releasing

When creating a new release:

1. Update the version number in relevant files (e.g., `VERSION`, `Chart.yaml`, etc.)
2. Update the CHANGELOG.md file with the new version and its changes
3. Create a new tag with the version number (e.g., `v1.0.0`)
4. Push the tag to the repository
5. The CI/CD pipeline will automatically:
   - Build the project
   - Generate the `capt.yaml` file
   - Create a new GitHub release
   - Attach the `capt.yaml` file to the release

Users can then download and apply the `capt.yaml` file to install or upgrade CAPT.

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
