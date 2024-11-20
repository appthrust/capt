# CAPTEP-0033: RBAC Permissions Update for Cluster Status

## Summary

This proposal outlines the necessary RBAC permission updates to allow CAPT to properly manage cluster status resources.

## Motivation

CAPT requires access to cluster status resources to properly manage and monitor the state of clusters. The current RBAC configuration lacks the necessary permissions to access `clusters/status` resources, which prevents proper cluster status management.

### Goals

- Add necessary RBAC permissions for `clusters/status` resources
- Ensure consistent cluster status management
- Document the permission requirements

### Non-Goals

- Modifying the overall RBAC structure
- Changing existing permissions for other resources

## Proposal

### User Stories

#### Story 1: Cluster Administrator

As a cluster administrator, I want to ensure that CAPT has the correct permissions to manage cluster status resources, so that cluster state can be properly monitored and updated.

### Implementation Details

1. Add `clusters/status` to the list of resources under the `cluster.x-k8s.io` API group:

```yaml
- apiGroups:
  - cluster.x-k8s.io
  resources:
  - clusters
  - clusters/status
  verbs:
  - get
  - list
  - patch
  - update
  - watch
```

### Risks and Mitigations

1. Risk: Insufficient permissions could prevent proper cluster management
   - Mitigation: Comprehensive testing of cluster lifecycle operations

2. Risk: Excessive permissions could pose security risks
   - Mitigation: Carefully scoped permissions to only required resources

## Design Details

### Test Plan

1. Verify cluster status updates are properly managed
2. Ensure cluster lifecycle operations work as expected
3. Validate permission changes through cluster operations

### Graduation Criteria

1. All tests passing with new permissions
2. No regression in existing functionality
3. Documentation updated to reflect permission requirements

## Implementation History

- 2024-01-24: Initial proposal
- 2024-01-24: Implementation of RBAC changes

## Alternatives Considered

1. Using separate service accounts for status management
   - Rejected due to increased complexity
   - Would require additional coordination between components

2. Implementing status management without status subresource
   - Rejected as it would not follow Kubernetes best practices
   - Would make it harder to manage status updates
