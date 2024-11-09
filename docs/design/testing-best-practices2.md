# Testing Best Practices Part 2: Controller Deletion Testing

## Background

During the improvement of controlplane controller tests, we discovered several important patterns and considerations for testing deletion scenarios in Kubernetes controllers. This document captures these learnings to help improve future controller implementations.

## Key Findings

### 1. Deletion Flow Testing

When testing controller deletion flows, we found the following critical points:

- **Resource Cleanup Order**: Resources must be cleaned up before removing finalizers
- **State Verification**: Object state must be verified at appropriate points in the deletion flow
- **Error Handling**: NotFound errors must be handled appropriately during deletion

### 2. Common Issues

Several issues were identified in the original implementation:

1. **Finalizer Removal Timing**
   - Issue: Finalizer was removed before resource cleanup
   - Solution: Ensure cleanup completes before finalizer removal

2. **Object State Verification**
   - Issue: Attempting to verify object state after deletion
   - Solution: Expect and handle NotFound errors appropriately

3. **Test Client Management**
   - Issue: Creating new clients for verification
   - Solution: Use the same client throughout the test

## Improved Testing Pattern

### Example: Deletion Test Case

```go
{
    name: "Resource being deleted",
    existingObjs: []runtime.Object{
        &v1beta1.YourObject{
            ObjectMeta: metav1.ObjectMeta{
                DeletionTimestamp: &metav1.Time{Time: time.Now()},
                Finalizers:        []string{YourFinalizer},
            },
        },
    },
    validate: func(t *testing.T, client client.Client, result Result, err error) {
        // Verify operation result
        assert.NoError(t, err)
        assert.Equal(t, expectedResult, result)

        // Verify object deletion
        obj := &v1beta1.YourObject{}
        err = client.Get(context.Background(), key, obj)
        assert.True(t, apierrors.IsNotFound(err), 
            "Expected NotFound error, got %v", err)
    },
}
```

### Key Improvements

1. **Validation Function Signature**
   - Include operation results in validation
   - Use the same client instance
   - Provide clear error messages

2. **Error Handling**
   - Properly handle NotFound errors
   - Distinguish between expected and unexpected errors
   - Improve error messages for debugging

3. **State Verification**
   - Verify object state at appropriate times
   - Handle deletion state correctly
   - Check cleanup completion

## Best Practices

1. **Resource Cleanup**
   - Clean up dependent resources first
   - Verify cleanup completion
   - Handle cleanup errors appropriately

2. **Finalizer Management**
   - Remove finalizers after successful cleanup
   - Verify finalizer removal
   - Handle update errors appropriately

3. **Error Handling**
   - Handle NotFound errors as expected in deletion flows
   - Provide clear error messages
   - Consider timing of operations

## Impact

These improvements have led to:
- More reliable deletion testing
- Clearer test failure messages
- Better handling of edge cases
- More maintainable test code

## Future Considerations

1. **Test Helper Functions**
   - Consider creating helper functions for common validation patterns
   - Standardize error handling in tests
   - Improve test readability

2. **Documentation**
   - Document common testing patterns
   - Provide examples of best practices
   - Keep testing documentation updated
