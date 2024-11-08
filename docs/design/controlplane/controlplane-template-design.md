# CaptControlPlaneTemplate Design

## Overview

CaptControlPlaneTemplate is a new API resource that enables the use of ClusterClass feature in Cluster API Provider Terraform (CAPT). This design document outlines the structure and implementation details of the CaptControlPlaneTemplate resource.

## Variable Types and Resolution

### Built-in Variables

Built-in variables are system-provided values that are automatically injected by the Cluster API controller. These variables provide access to fundamental cluster information without requiring explicit definition by users.

1. Purpose:
   - Provide access to core cluster information
   - Ensure consistency in referencing system values
   - Reduce configuration complexity

2. Available Built-in Variables:
   - `builtin.cluster.*`: Core cluster information
     * name: Cluster name
     * namespace: Cluster namespace
     * uid: Cluster unique identifier
     * topology.version: Desired Kubernetes version
   - `builtin.controlPlane.*`: Control plane information
     * version: Kubernetes version
     * replicas: Number of replicas
     * name: Control plane name
   - `builtin.machineDeployment.*`: Machine deployment information
     * version: Kubernetes version for the deployment
     * replicas: Number of replicas
     * name: Deployment name

### User-defined Variables

User-defined variables are specified in the ClusterClass and their values are provided in the Cluster specification.

1. Purpose:
   - Enable customization of cluster configuration
   - Allow environment-specific settings
   - Support flexible infrastructure configuration

2. Definition in ClusterClass:
```yaml
variables:
  - name: controlPlane.endpointAccess.public
    required: true
    schema:
      openAPIV3Schema:
        type: boolean
        default: true
```

### Variable Resolution Process

1. Cluster API Controller responsibilities:
   - Injects built-in variables based on cluster state
   - Validates user-defined variables against schema
   - Performs variable substitution in templates
   - Applies patches with resolved variables

2. CAPT Controller responsibilities:
   - Receives templates with all variables resolved
   - Creates/updates resources based on resolved templates
   - Manages infrastructure provisioning

## Implementation Details

### Template Structure

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: CaptControlPlaneTemplate
metadata:
  name: example-control-plane-template
spec:
  template:
    spec:
      # Using built-in variable for version
      version: ${builtin.controlPlane.version}
      workspaceTemplateRef:
        name: eks-controlplane-template
      controlPlaneConfig:
        # Using user-defined variables
        endpointAccess:
          public: ${controlPlane.endpointAccess.public}
          private: ${controlPlane.endpointAccess.private}
```

### Variable Usage Examples

1. Built-in Variables:
```yaml
# Automatically use cluster name
cluster_name: ${builtin.cluster.name}

# Use control plane version
kubernetes_version: ${builtin.controlPlane.version}
```

2. User-defined Variables:
```yaml
# Environment configuration
environment: ${cluster.environment}

# Access control settings
public_access: ${controlPlane.endpointAccess.public}
```

## Best Practices

1. Variable Selection:
   - Use built-in variables for system values (names, versions, etc.)
   - Use user-defined variables for customizable settings
   - Avoid duplicating built-in variable information

2. Template Design:
   - Leverage built-in variables for consistent naming
   - Use user-defined variables for environment-specific settings
   - Combine both types for flexible yet consistent templates

3. Variable Resolution:
   - Trust built-in variables for system state
   - Validate user-defined variables with appropriate schemas
   - Consider default values for optional settings

## Implementation Plan

### Phase 1: Basic Template Support
1. Implement CaptControlPlaneTemplate CRD
2. Add template controller
3. Support variable resolution integration
4. Integrate with existing WorkspaceTemplate

### Phase 2: Variable Integration
1. Support built-in variable usage
2. Implement user-defined variable handling
3. Add variable validation
4. Test variable resolution

### Phase 3: Advanced Features
1. Add complex variable support
2. Implement conditional logic
3. Add template composition
4. Enhance security features
