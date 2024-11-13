# CAPTControlPlane Refactoring Design

## Overview

This document outlines the design decisions and implementation plan for refactoring the CAPTControlPlane controller.
The main goal is to improve code organization, maintainability, and testability by following patterns established in the CAPTCluster implementation.

## Current Issues

1. Code Organization
   - All controller logic is contained in a single file
   - Lack of clear separation of concerns
   - Difficult to maintain and test

2. Error Handling
   - Inconsistent error handling patterns
   - Mixed responsibility between error handling and status updates

3. Testing
   - Lack of unit tests
   - Difficult to test individual components

## Design Goals

1. Improve Code Organization
   - Separate concerns into distinct files
   - Clear responsibility boundaries
   - Consistent patterns across the codebase

2. Enhanced Error Handling
   - Unified error handling approach
   - Clear separation between error handling and status updates
   - Improved error reporting in status conditions

3. Better Testing
   - Unit tests for each component
   - Improved testability through separation of concerns
   - Higher test coverage

## Implementation Plan

### 1. Directory Structure

```
internal/controller/controlplane/
├── controller.go        # Main controller logic
├── status.go           # Status management
├── finalizer.go        # Finalizer handling
├── workspace.go        # WorkspaceTemplate operations
├── types.go           # Common types and constants
└── tests/             # Test files
```

### 2. Component Responsibilities

#### controller.go
- Main reconciliation loop
- High-level orchestration
- Resource ownership management

#### status.go
- Status condition management
- Status update operations
- Status-related helper functions

#### finalizer.go
- Finalizer addition/removal
- Cleanup operations
- Resource deletion handling

#### workspace.go
- WorkspaceTemplate operations
- WorkspaceTemplateApply management
- Workspace-related helper functions

#### types.go
- Common types
- Constants
- Shared interfaces

### 3. Error Handling Strategy

1. Error Types
   - Define specific error types for different failure scenarios
   - Implement error wrapping for better context

2. Status Updates
   - Separate error handling from status updates
   - Consistent status condition updates
   - Clear error messages in status

### 4. Testing Strategy

1. Unit Tests
   - Test each component independently
   - Mock dependencies
   - Focus on edge cases

2. Integration Tests
   - Test component interactions
   - Verify end-to-end workflows
   - Test failure scenarios

## Migration Plan

1. Phase 1: File Structure
   - Create new file structure
   - Move existing code to appropriate files
   - Maintain current functionality

2. Phase 2: Refactoring
   - Implement new patterns
   - Update error handling
   - Add new helper functions

3. Phase 3: Testing
   - Add unit tests
   - Implement integration tests
   - Verify coverage

## Benefits

1. Maintainability
   - Easier to understand and modify
   - Clear separation of concerns
   - Consistent patterns

2. Testability
   - Better test coverage
   - Easier to write tests
   - More reliable testing

3. Reliability
   - Improved error handling
   - Better status management
   - Clearer failure reporting

## Risks and Mitigation

1. Risks
   - Breaking changes during refactoring
   - Regression issues
   - Performance impact

2. Mitigation
   - Comprehensive testing
   - Gradual implementation
   - Performance monitoring

## Future Considerations

1. Extensibility
   - Design for future features
   - Plugin architecture
   - API versioning

2. Performance
   - Optimization opportunities
   - Caching strategies
   - Resource usage

## References

- CAPTCluster implementation patterns
- Kubernetes controller best practices
- Go project layout standards
