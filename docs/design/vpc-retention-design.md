# VPC Retention Design

## Overview

This document describes the design and implementation of VPC retention functionality in CAPTCluster. This feature enables VPC resources to be retained when needed, even if the parent cluster is deleted.

## Background

### Challenges

1. VPC resources can be shared across multiple projects
2. Shared VPCs are deleted when the parent cluster is deleted
3. VPC deletion can impact other projects

### Requirements

1. Need functionality to control VPC retention/deletion
2. This control must be explicitly configurable
3. Existing VPC usage should not be affected

## Design Decision

### API Design

#### CaptCluster

```go
type CaptClusterSpec struct {
    // RetainVPCOnDelete specifies whether to retain the VPC when the parent cluster is deleted
    // This is useful when the VPC is shared among multiple projects
    // This field is only effective when VPCTemplateRef is set
    // +optional
    RetainVPCOnDelete bool `json:"retainVpcOnDelete,omitempty"`
}
```

#### WorkspaceTemplateApply

```go
type WorkspaceTemplateApplySpec struct {
    // RetainWorkspaceOnDelete specifies whether to retain the Workspace when this WorkspaceTemplateApply is deleted
    // This is useful when the Workspace manages shared resources that should outlive this WorkspaceTemplateApply
    // +optional
    RetainWorkspaceOnDelete bool `json:"retainWorkspaceOnDelete,omitempty"`
}
```

### Design Considerations

1. **Explicit Configuration**
   - CaptCluster: Defaults to false (VPC is deleted)
   - WorkspaceTemplateApply: Defaults to false (Workspace is deleted)
   - Resources are only retained when explicitly set to true

2. **Scope Limitations**
   - CaptCluster: Only effective when using VPCTemplateRef
   - WorkspaceTemplateApply: Effective in all cases

3. **Validation**
   - RetainVPCOnDelete is only valid when VPCTemplateRef is set
   - Invalid combinations are detected early

### WorkspaceTemplateApply Retention Feature

1. **Design Background**
   - Current behavior where related Workspace is automatically deleted when WorkspaceTemplateApply is deleted
   - This behavior is problematic for cases requiring shared resource (VPC, etc.) retention

2. **New Design Approach**
   - Add direct Workspace retention setting to WorkspaceTemplateApply
   - Can be controlled individually for each WorkspaceTemplateApply, independent of parent resource settings

3. **Benefits**
   - Flexibility: Can be controlled individually for each WorkspaceTemplateApply
   - Clarity: Resource retention intent is explicitly documented
   - Reusability: Different retention strategies can be applied for different use cases using the same WorkspaceTemplate

### Key Implementation Points

1. **Deletion Control**
```go
func (r *workspaceTemplateApplyReconciler) reconcileDelete(ctx context.Context, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) (ctrl.Result, error) {
    if workspaceApply.Spec.RetainWorkspaceOnDelete {
        // Skip deletion process if Workspace should be retained
        return ctrl.Result{}, nil
    }
    // Normal deletion process
    ...
}
```

2. **Validation**
```go
func (s *WorkspaceTemplateApplySpec) ValidateConfiguration() error {
    // Add validations as needed
    return nil
}
```

3. **Logging**
   - Clearly log resource retention/deletion decisions
   - Facilitate troubleshooting

## Usage Example

### VPC Retention Configuration

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: CAPTCluster
metadata:
  name: example-retained-vpc
spec:
  region: ap-northeast-1
  vpcTemplateRef:
    name: vpc-template
  retainVpcOnDelete: true  # Configuration to retain VPC
```

### Workspace Retention Configuration

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplateApply
metadata:
  name: shared-vpc
spec:
  templateRef:
    name: vpc-template
  retainWorkspaceOnDelete: true  # Configuration to retain Workspace
```

## Lessons Learned

1. **Feature Separation**
   - Resource retention/deletion control needs to be managed independently at each resource level
   - This makes configuration intent clear and prevents misuse

2. **Importance of Explicit Configuration**
   - Default to the safe side (delete resources)
   - Require explicit configuration when retention is needed

3. **Importance of Validation**
   - Early validation prevents configuration mistakes
   - Error messages should be specific and easy to understand

4. **Documentation and Samples**
   - Clear samples demonstrating feature usage are important
   - Make configuration intent and impact easy to understand

## Future Considerations

1. **Extensibility**
   - Similar retention functionality may be needed for other resource types
   - Consider abstraction as a common pattern

2. **Monitoring**
   - Make resource retention/deletion decisions monitorable
   - Consider metrics collection

3. **Lifecycle Management**
   - How to manage retained resources
   - Consider cleanup policies

## References

- [Terraform Outputs Management](./terraform-outputs-management.md)
- [Cluster Status Management](./cluster-status-management.md)
