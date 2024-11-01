# VPC Configuration Migration Guide

This guide explains how to migrate from the current VPCConfig-based approach to the new WorkspaceTemplate-based approach for VPC configuration in CAPT.

## Overview

The CAPT project is moving from embedding VPC configuration directly in CAPTClusterSpec to using WorkspaceTemplate for better separation of concerns and reusability. This document provides step-by-step instructions for migrating existing clusters and creating new ones using the new approach.

## Changes

### API Changes

1. CAPTClusterSpec:
   - Removed: `VPCConfig` struct
   - Added: `VPCTemplateRef` for referencing VPC WorkspaceTemplate
   - Added: `VPCSelector` for selecting existing VPCs

2. New Resources:
   - VPC WorkspaceTemplate
   - EKS WorkspaceTemplate

## Migration Steps

### For Existing Clusters

1. Create a VPC WorkspaceTemplate:
   ```yaml
   apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
   kind: WorkspaceTemplate
   metadata:
     name: vpc-template
   spec:
     template:
       metadata:
         description: "Standard VPC configuration for EKS clusters"
         version: "1.0.0"
       spec:
         module:
           source: "terraform-aws-modules/vpc/aws"
           version: "5.0.0"
         variables:
           name:
             value: "${var.cluster_name}-vpc"
           cidr:
             value: "${var.vpc_cidr}"
           # ... other VPC configurations
   ```

2. Update CAPTCluster configuration:
   ```yaml
   apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
   kind: CAPTCluster
   metadata:
     name: my-cluster
   spec:
     region: us-west-2
     vpcTemplateRef:
       apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
       kind: WorkspaceTemplate
       name: vpc-template
     # ... other configurations
   ```

### For New Clusters

1. Choose or create a VPC WorkspaceTemplate that matches your requirements.

2. Create your CAPTCluster with a reference to the VPC WorkspaceTemplate:
   ```yaml
   apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
   kind: CAPTCluster
   metadata:
     name: new-cluster
   spec:
     region: us-west-2
     vpcTemplateRef:
       apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
       kind: WorkspaceTemplate
       name: vpc-template
     eks:
       version: "1.27"
       publicAccess: true
       privateAccess: true
     # ... other configurations
   ```

## Using Existing VPCs

To use an existing VPC instead of creating a new one:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: my-cluster
spec:
  region: us-west-2
  vpcSelector:
    matchLabels:
      environment: production
      type: eks
  # ... other configurations
```

## Best Practices

1. Use standardized VPC templates for consistent configurations across clusters
2. Maintain version control for WorkspaceTemplates
3. Document any customizations made to standard templates
4. Use descriptive names and tags for better resource management

## Troubleshooting

1. If migration fails:
   - Check WorkspaceTemplate syntax
   - Verify all required variables are provided
   - Check controller logs for detailed error messages

2. Common issues:
   - Missing or incorrect template references
   - Invalid variable values
   - Permission issues

## Rollback Procedure

If you need to rollback to the previous VPCConfig-based approach:

1. Keep the old VPCConfig in your CAPTCluster spec
2. Remove VPCTemplateRef if present
3. Update to the previous version of the CAPT controller

## Support

For issues or questions about the migration:
1. Check the troubleshooting guide
2. Review controller logs
3. Open an issue in the CAPT repository

## Timeline

1. Phase 1: Dual support for both approaches
2. Phase 2: Deprecation warning for VPCConfig
3. Phase 3: Remove VPCConfig support

## Next Steps

1. Review existing cluster configurations
2. Create standardized VPC templates
3. Plan migration timeline
4. Test in non-production environment first
5. Monitor and validate migrated clusters
