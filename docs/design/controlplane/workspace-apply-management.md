# WorkspaceTemplateApply Management Design

## Overview

This document provides detailed design information about how the CAPTControlPlane controller manages WorkspaceTemplateApply resources, focusing on naming, lifecycle, and state management.

## Background

The controller needs to maintain a stable relationship with WorkspaceTemplateApply resources while supporting both existing and new deployments. This design addresses the challenges of managing these resources consistently.

## Detailed Design

### WorkspaceTemplateApply Name Management

1. Name Storage
   ```yaml
   spec:
     workspaceTemplateApplyName: string  # Stores the reference to WorkspaceTemplateApply
   ```

2. Name Resolution Strategy
   ```go
   // Pseudo-code for name resolution
   if spec.WorkspaceTemplateApplyName != "" {
       // Use existing reference
       return spec.WorkspaceTemplateApplyName
   } else {
       // Try to find legacy name
       if exists("demo-eks-controlplane-apply") {
           // Use and store legacy name
           return "demo-eks-controlplane-apply"
       }
       // Create new name
       return fmt.Sprintf("%s-eks-controlplane-apply", controlPlane.Name)
   }
   ```

### State Management

1. Condition Tracking
   ```yaml
   status:
     conditions:
       - type: Ready
         status: "True"
         reason: "WorkspaceReady"
         message: "Control plane is ready"
   ```

2. Phase Transitions
   ```
   Creating -> Ready:
     - WorkspaceTemplateApply.Applied = true
     - WorkspaceTemplateApply.Synced = true
     - WorkspaceTemplateApply.Ready = true
   
   Any -> Failed:
     - Error condition detected
     - Clear status fields
     - Set failure reason
   ```

### Resource Lifecycle

1. Creation
   - Check for existing WorkspaceTemplateApply
   - Set owner references
   - Store name in spec

2. Updates
   - Maintain existing references
   - Update spec as needed
   - Preserve owner references

3. Deletion
   - Clean up WorkspaceTemplateApply
   - Remove finalizers
   - Clear status

## Implementation Details

### API Changes

1. CAPTControlPlaneSpec
   ```go
   type CAPTControlPlaneSpec struct {
       // ... existing fields ...
       
       // WorkspaceTemplateApplyName is the name of the WorkspaceTemplateApply
       // +optional
       WorkspaceTemplateApplyName string `json:"workspaceTemplateApplyName,omitempty"`
   }
   ```

### Controller Logic

1. Name Resolution
   ```go
   func (r *Reconciler) resolveWorkspaceApplyName(controlPlane *CAPTControlPlane) string {
       if controlPlane.Spec.WorkspaceTemplateApplyName != "" {
           return controlPlane.Spec.WorkspaceTemplateApplyName
       }
       // ... name resolution logic ...
   }
   ```

2. Status Updates
   ```go
   func (r *Reconciler) updateStatus(ctx context.Context, 
       controlPlane *CAPTControlPlane,
       workspaceApply *WorkspaceTemplateApply) error {
       // ... status update logic ...
   }
   ```

## Migration Strategy

### Existing Deployments

1. Detection
   ```go
   // Look for existing WorkspaceTemplateApply
   existingApply := &WorkspaceTemplateApply{}
   err := r.Get(ctx, types.NamespacedName{
       Name: "demo-eks-controlplane-apply",
       Namespace: controlPlane.Namespace,
   }, existingApply)
   ```

2. Migration
   ```go
   // Update spec with existing name
   controlPlane.Spec.WorkspaceTemplateApplyName = existingApply.Name
   if err := r.Update(ctx, controlPlane); err != nil {
       return err
   }
   ```

### New Deployments

1. Name Generation
   ```go
   // Generate new name
   applyName := fmt.Sprintf("%s-eks-controlplane-apply", controlPlane.Name)
   ```

2. Resource Creation
   ```go
   // Create new WorkspaceTemplateApply
   workspaceApply := &WorkspaceTemplateApply{
       ObjectMeta: metav1.ObjectMeta{
           Name: applyName,
           Namespace: controlPlane.Namespace,
       },
       // ... spec configuration ...
   }
   ```

## Testing Strategy

1. Unit Tests
   - Name resolution logic
   - Status update logic
   - Migration paths

2. Integration Tests
   - Resource creation flows
   - Status propagation
   - Error handling

3. Migration Tests
   - Existing resource detection
   - Name updates
   - Resource preservation
