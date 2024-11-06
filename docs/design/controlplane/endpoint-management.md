# Control Plane Endpoint Management

## Overview

This document describes the design and implementation details of endpoint management in the CAPTControlPlane controller. The endpoint management system is responsible for retrieving and propagating EKS cluster endpoint information from Terraform-managed resources to Cluster API resources.

## Architecture

### Component Interaction

```
WorkspaceTemplateApply
        ↓
    Workspace
    ↙       ↘
Outputs    Secret
    ↘       ↙
CAPTControlPlane
        ↓
    Cluster
```

### Key Components

1. **WorkspaceTemplateApply**
   - Manages the lifecycle of Terraform workspaces
   - Contains the reference to the actual Workspace in its status

2. **Workspace**
   - Contains the actual Terraform state and outputs
   - Can expose information through:
     - status.atProvider.outputs
     - writeConnectionSecretToRef

3. **Endpoint Package**
   - Dedicated package for endpoint management
   - Uses unstructured.Unstructured for flexible resource access
   - Implements fallback mechanism for endpoint retrieval

## Implementation Details

### Endpoint Retrieval Strategy

The system implements a prioritized strategy for endpoint retrieval:

1. **Primary Source: Workspace Outputs**
   ```go
   outputs, found, err := unstructured.NestedMap(workspace.Object, "status", "atProvider", "outputs")
   if found && outputs != nil {
       if endpoint, ok := outputs["cluster_endpoint"].(string); ok {
           return &clusterv1.APIEndpoint{
               Host: endpoint,
               Port: 443,
           }, nil
       }
   }
   ```

2. **Fallback: Connection Secret**
   ```go
   if endpointData, ok := secret.Data["cluster_endpoint"]; ok {
       endpoint, err := base64.StdEncoding.DecodeString(string(endpointData))
       if err == nil {
           return &clusterv1.APIEndpoint{
               Host: string(endpoint),
               Port: 443,
           }, nil
       }
   }
   ```

### Resource Access Pattern

The implementation uses unstructured.Unstructured to access Workspace resources:

```go
workspace := &unstructured.Unstructured{}
workspace.SetGroupVersionKind(schema.GroupVersionKind{
    Group:   "tf.upbound.io",
    Version: "v1beta1",
    Kind:    "Workspace",
})
```

This approach:
- Avoids direct dependencies on provider-specific types
- Provides flexibility in accessing resource fields
- Reduces the impact of API changes

### Error Handling

The implementation includes comprehensive error handling:

1. **Resource Retrieval**
   - Graceful handling of missing resources
   - Detailed error messages for debugging

2. **Data Extraction**
   - Validation of data presence
   - Type assertion safety checks

3. **Endpoint Processing**
   - Base64 decoding error handling
   - Null safety checks

## RBAC Configuration

Required RBAC rules for endpoint management:

```yaml
//+kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
```

## Status Updates

The controller updates multiple resources:

1. **CAPTControlPlane**
   ```go
   patchBase := controlPlane.DeepCopy()
   controlPlane.Spec.ControlPlaneEndpoint = *apiEndpoint
   if err := r.Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
       // Error handling
   }
   ```

2. **Owner Cluster**
   ```go
   if cluster != nil {
       patchBase := cluster.DeepCopy()
       cluster.Spec.ControlPlaneEndpoint = controlPlane.Spec.ControlPlaneEndpoint
       if err := r.Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
           // Error handling
       }
   }
   ```

## Best Practices

1. **Separation of Concerns**
   - Endpoint management logic is isolated in a dedicated package
   - Clear responsibility boundaries

2. **Resource Access**
   - Use of unstructured.Unstructured for flexible resource access
   - Proper error handling and logging

3. **Status Updates**
   - Use of patch operations for atomic updates
   - Verification of updates through re-fetching resources

4. **Error Handling**
   - Comprehensive error checking
   - Detailed logging for debugging

## Logging

The implementation includes detailed logging:

```go
logger.Info("Found Workspace", "name", workspace.GetName())
logger.Info("Found cluster_endpoint in Workspace outputs", "endpoint", endpoint)
logger.Info("Successfully patched CAPTControlPlane endpoint")
```

This helps in:
- Debugging issues
- Understanding the flow of operations
- Monitoring system behavior

## Future Improvements

1. **Caching**
   - Implement caching for frequently accessed resources
   - Add cache invalidation strategies

2. **Retry Mechanism**
   - Add retry logic for transient failures
   - Implement exponential backoff

3. **Metrics**
   - Add metrics for endpoint retrieval success/failure
   - Track timing information

4. **Validation**
   - Add endpoint format validation
   - Implement health checks
