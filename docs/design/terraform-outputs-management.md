# Terraform Outputs Management

## Overview

This document describes the design considerations and best practices for managing Terraform outputs in the CAPT project, particularly focusing on how outputs are exposed through Workspace status and secrets.

## Problem Statement

When working with Terraform outputs in Workspaces, we encountered an issue where certain outputs were not appearing in the Workspace's status.atProvider.outputs. This was particularly problematic for the VPC ID, which needed to be reflected in the CAPTCluster's status.

The root cause was that all outputs in the VPC Workspace were marked as `sensitive = true`, which prevented them from appearing in the status.atProvider.outputs field.

## Design Decision

### Output Visibility Strategy

We have established the following strategy for managing Terraform outputs:

1. **Non-Sensitive Outputs (Primary Source)**
   - Basic resource identifiers (e.g., vpc_id, subnet_ids)
   - Public endpoints or ARNs
   - Resource names or other non-sensitive metadata
   - These should be exposed in status.atProvider.outputs

2. **Sensitive Outputs (Stored in Secrets)**
   - Configuration data containing multiple fields
   - Credentials or secrets
   - Connection strings containing sensitive information
   - These should be stored in Secrets via writeConnectionSecretToRef

### Component Interaction

```
WorkspaceTemplateApply
        ↓
    Workspace
    ↙       ↘
Outputs    Secret
    ↘       ↙
  Resource Status
```

This flow aligns with the existing Control Plane endpoint management architecture, ensuring consistency across different components.

### Implementation Guidelines

1. **Resource Identifiers**
```hcl
output "vpc_id" {
  description = "The ID of the VPC"
  value       = module.vpc.vpc_id
  # No sensitive = true for basic identifiers
}
```

2. **Configuration Data**
```hcl
output "vpc_config" {
  description = "VPC configuration in HCL format"
  value       = <<-EOT
    vpc_id = "${module.vpc.vpc_id}"
    private_subnets = ${jsonencode(module.vpc.private_subnets)}
  EOT
  sensitive   = true  # Configuration data should be sensitive
}
```

### Retrieval Strategy

Following the pattern established in endpoint management, implement a prioritized retrieval strategy:

1. **Primary: Status Outputs**
```go
outputs, found, err := unstructured.NestedMap(workspace.Object, "status", "atProvider", "outputs")
if found && outputs != nil {
    if value, ok := outputs["resource_id"].(string); ok {
        return value, nil
    }
}
```

2. **Fallback: Secrets**
```go
if secretData, ok := secret.Data["resource_id"]; ok {
    value, err := base64.StdEncoding.DecodeString(string(secretData))
    if err == nil {
        return string(value), nil
    }
}
```

## Example: VPC and Control Plane Comparison

### VPC Workspace
```hcl
# Basic identifiers - Not sensitive
output "vpc_id" {
  value = module.vpc.vpc_id
}

# Configuration data - Sensitive
output "vpc_config" {
  value     = "..."
  sensitive = true
}
```

### Control Plane Workspace
```hcl
# Basic endpoint - Not sensitive
output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

# Credentials - Sensitive
output "kubeconfig" {
  value     = "..."
  sensitive = true
}
```

## Status Management

The status management follows the established pattern:

1. **Status Flow**
```
Creating -> Ready (Outputs Available)
    |
    ↓
Failed (Error State)
```

2. **Resource Relationship**
```
Resource CR -> WorkspaceTemplateApply -> Workspace
                                          ↓
                                    Outputs/Secrets
                                          ↓
                                    Status Update
```

## Impact on Controllers

Controllers should implement the following pattern:

1. **Output Retrieval**
   - First attempt to get from status.atProvider.outputs
   - Fall back to secrets if necessary
   - Handle both cases consistently

2. **Status Updates**
   - Update resource status with retrieved information
   - Maintain proper error handling
   - Ensure atomic updates using patch operations

## Migration Guide

When implementing new Terraform modules:

1. Identify outputs that are basic resource identifiers
2. Remove `sensitive = true` from these basic outputs
3. Keep `sensitive = true` for configuration data and credentials
4. Update controllers to expect resource IDs in status.atProvider.outputs
5. Implement proper fallback to secrets when needed

## Best Practices

1. **Separation of Concerns**
   - Keep basic identifiers in status outputs
   - Use secrets for sensitive configuration
   - Maintain clear boundaries between different types of data

2. **Resource Access**
   - Use unstructured.Unstructured for workspace access
   - Implement proper error handling
   - Add comprehensive logging

3. **Status Updates**
   - Use patch operations for atomic updates
   - Verify updates through re-fetching
   - Maintain proper error handling

## Future Considerations

1. **Documentation**
   - Update module documentation to clearly indicate which outputs are sensitive
   - Provide examples of proper output configuration

2. **Validation**
   - Add validation for output sensitivity settings
   - Implement checks in CI/CD pipeline

3. **Monitoring**
   - Add metrics for tracking output availability in status
   - Monitor secret creation and updates

4. **Enhancement**
   - Consider caching mechanisms for frequently accessed data
   - Implement retry mechanisms for transient failures
   - Add health checking capabilities
