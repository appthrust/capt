# Status Management Best Practices

## Overview

This document outlines the best practices and lessons learned from implementing status management in the CAPT controllers, particularly focusing on the CAPTControlPlane controller's endpoint management implementation.

## Key Principles

### 1. Separation of Concerns

#### Problem
Status management logic can become complex and intertwined with other controller logic.

#### Solution
- Isolate status-related logic into dedicated packages or functions
- Clear separation between status updates and other operations
- Example:
  ```go
  // Dedicated package for endpoint management
  package endpoint

  // Clear, single-responsibility function
  func GetEndpointFromWorkspace(ctx context.Context, c client.Client, workspaceName string) (*clusterv1.APIEndpoint, error)
  ```

### 2. Resource Access Patterns

#### Problem
Direct dependencies on provider-specific types can make code brittle and hard to maintain.

#### Solution
- Use unstructured.Unstructured for flexible resource access
- Implement proper error handling and type assertions
- Example:
  ```go
  workspace := &unstructured.Unstructured{}
  workspace.SetGroupVersionKind(schema.GroupVersionKind{
      Group:   "tf.upbound.io",
      Version: "v1beta1",
      Kind:    "Workspace",
  })
  ```

### 3. Status Update Patterns

#### Problem
Status updates can fail due to resource conflicts or race conditions.

#### Solution
- Use patch operations instead of updates where possible
- Always work with fresh copies of resources
- Example:
  ```go
  patchBase := controlPlane.DeepCopy()
  // Make changes to controlPlane
  if err := r.Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
      // Handle error
  }
  ```

## Status Propagation

### 1. Multi-Resource Status Chain

#### Pattern
```
Source Resource → Intermediate Resource → Target Resource
(Workspace) → (CAPTControlPlane) → (Cluster)
```

#### Best Practices
- Clear definition of status flow
- Proper error handling at each step
- Verification of successful propagation

### 2. Status Source Priority

#### Pattern
```
Primary Source → Fallback Source
(Workspace Outputs) → (Connection Secret)
```

#### Best Practices
- Clear priority order for status sources
- Proper logging of source selection
- Graceful fallback handling

## Error Handling

### 1. Comprehensive Error Checking

#### Pattern
```go
if err != nil {
    logger.Error(err, "Failed to get resource",
        "name", name,
        "namespace", namespace)
    return nil, fmt.Errorf("failed to get resource: %w", err)
}
```

#### Best Practices
- Detailed error messages
- Proper error wrapping
- Context-rich logging

### 2. Graceful Degradation

#### Pattern
```go
if primarySource == nil {
    // Try fallback source
    if fallbackSource == nil {
        // Handle complete failure
    }
}
```

#### Best Practices
- Multiple fallback options
- Clear logging of fallback attempts
- Proper error propagation

## Logging

### 1. Structured Logging

#### Pattern
```go
logger.Info("Processing resource",
    "name", resource.Name,
    "type", resource.Kind,
    "status", resource.Status)
```

#### Best Practices
- Consistent log levels
- Structured key-value pairs
- Meaningful context information

### 2. Operation Tracking

#### Pattern
```go
logger.Info("Starting operation")
// ... operation ...
logger.Info("Completed operation", "result", result)
```

#### Best Practices
- Clear operation boundaries
- Success/failure logging
- Performance metrics

## Resource Management

### 1. Resource Lifecycle

#### Pattern
```go
// Create/Update/Delete operations
if err := r.Create(ctx, resource); err != nil {
    // Handle error
}
```

#### Best Practices
- Proper cleanup
- Resource ownership
- Dependency management

### 2. Resource References

#### Pattern
```go
// Reference management
if err := ctrl.SetControllerReference(owner, resource, r.Scheme); err != nil {
    // Handle error
}
```

#### Best Practices
- Clear ownership chains
- Proper garbage collection
- Reference validation

## Testing

### 1. Status Update Testing

#### Pattern
```go
// Test status updates
func TestUpdateStatus(t *testing.T) {
    // Setup
    // Execute
    // Verify
}
```

#### Best Practices
- Comprehensive test cases
- Edge case coverage
- Proper mocking

### 2. Error Path Testing

#### Pattern
```go
// Test error conditions
func TestErrorHandling(t *testing.T) {
    // Setup error conditions
    // Execute
    // Verify error handling
}
```

#### Best Practices
- Error injection
- Recovery testing
- Timeout handling

## Lessons Learned

### 1. Status Design

- Keep status updates atomic
- Use proper type assertions
- Implement clear status flow

### 2. Error Handling

- Implement comprehensive error checking
- Provide clear error messages
- Use proper error wrapping

### 3. Resource Access

- Use unstructured types when appropriate
- Implement proper RBAC
- Handle resource not found cases

### 4. Testing

- Test all status paths
- Verify error handling
- Test resource cleanup

## Future Improvements

### 1. Status Management

- Implement status caching
- Add status validation
- Improve error recovery

### 2. Resource Access

- Implement resource watching
- Add resource caching
- Improve performance

### 3. Testing

- Add integration tests
- Improve test coverage
- Add performance tests
