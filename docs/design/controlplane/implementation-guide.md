# ControlPlane Implementation Guide

## Overview

This document provides detailed implementation guidance for the CAPTControlPlane controller, focusing on the WorkspaceTemplateApply management and status updates.

## API Changes

### CAPTControlPlaneSpec

```go
type CAPTControlPlaneSpec struct {
    // ... existing fields ...

    // WorkspaceTemplateApplyName is the name of the WorkspaceTemplateApply
    // used for this control plane. This field is managed by the controller
    // and should not be modified manually.
    // +optional
    WorkspaceTemplateApplyName string `json:"workspaceTemplateApplyName,omitempty"`
}
```

## Controller Implementation

### WorkspaceTemplateApply Management

1. Name Resolution
```go
// Try to find existing WorkspaceTemplateApply
if controlPlane.Spec.WorkspaceTemplateApplyName != "" {
    // Use the name from spec if it exists
    applyName = controlPlane.Spec.WorkspaceTemplateApplyName
} else {
    // Try to find the existing apply with the legacy name
    legacyName := "demo-eks-controlplane-apply"
    err := r.Get(ctx, types.NamespacedName{
        Name: legacyName,
        Namespace: controlPlane.Namespace,
    }, workspaceApply)
    if err == nil {
        applyName = legacyName
        // Update the spec with the found name
        patch := client.MergeFrom(controlPlane.DeepCopy())
        controlPlane.Spec.WorkspaceTemplateApplyName = legacyName
        if err := r.Patch(ctx, controlPlane, patch); err != nil {
            return err
        }
    }
}
```

2. Resource Creation/Update
```go
// Create or update WorkspaceTemplateApply
workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{
    ObjectMeta: metav1.ObjectMeta{
        Name:      applyName,
        Namespace: controlPlane.Namespace,
    },
}

// Set owner reference
if err := ctrl.SetControllerReference(controlPlane, workspaceApply, r.Scheme); err != nil {
    return nil, fmt.Errorf("failed to set controller reference: %w", err)
}
```

### Status Management

1. Condition Updates
```go
// Update status based on WorkspaceTemplateApply conditions
var syncedCondition, readyCondition bool
var errorMessage string

for _, condition := range workspaceApply.Status.Conditions {
    if condition.Type == xpv1.TypeSynced {
        syncedCondition = condition.Status == corev1.ConditionTrue
        if !syncedCondition {
            errorMessage = condition.Message
        }
    }
    if condition.Type == xpv1.TypeReady {
        readyCondition = condition.Status == corev1.ConditionTrue
        if !readyCondition {
            errorMessage = condition.Message
        }
    }
}
```

2. Phase Updates
```go
// Update phase based on conditions
if !workspaceApply.Status.Applied || !syncedCondition || !readyCondition {
    controlPlane.Status.Phase = "Creating"
    controlPlane.Status.Ready = false
} else {
    controlPlane.Status.Phase = "Ready"
    controlPlane.Status.Ready = true
}
```

## Testing

### Unit Tests

1. Name Resolution
```go
func TestWorkspaceApplyNameResolution(t *testing.T) {
    tests := []struct {
        name           string
        controlPlane   *CAPTControlPlane
        existingApply  *WorkspaceTemplateApply
        expectedName   string
    }{
        {
            name: "Use existing name from spec",
            controlPlane: &CAPTControlPlane{
                Spec: CAPTControlPlaneSpec{
                    WorkspaceTemplateApplyName: "existing-name",
                },
            },
            expectedName: "existing-name",
        },
        // ... more test cases ...
    }
    // ... test implementation ...
}
```

2. Status Updates
```go
func TestStatusUpdates(t *testing.T) {
    tests := []struct {
        name           string
        workspaceApply *WorkspaceTemplateApply
        expectedPhase  string
        expectedReady  bool
    }{
        {
            name: "All conditions true",
            workspaceApply: &WorkspaceTemplateApply{
                Status: WorkspaceTemplateApplyStatus{
                    Applied: true,
                    Conditions: []metav1.Condition{
                        {Type: "Ready", Status: "True"},
                        {Type: "Synced", Status: "True"},
                    },
                },
            },
            expectedPhase: "Ready",
            expectedReady: true,
        },
        // ... more test cases ...
    }
    // ... test implementation ...
}
```

### Integration Tests

1. Resource Creation Flow
```go
func TestResourceCreationFlow(t *testing.T) {
    // Setup test environment
    // Create CAPTControlPlane
    // Verify WorkspaceTemplateApply creation
    // Verify status updates
}
```

2. Migration Flow
```go
func TestMigrationFlow(t *testing.T) {
    // Setup test environment with existing resources
    // Create CAPTControlPlane
    // Verify existing resource detection
    // Verify spec updates
    // Verify status propagation
}
```

## Debugging

### Common Issues

1. Name Resolution
- Check if WorkspaceTemplateApplyName is set in spec
- Verify existing resource detection
- Check owner references

2. Status Updates
- Verify condition propagation
- Check phase transitions
- Monitor error handling

### Logging

Add detailed logging at key points:
```go
logger.Info("Reconciling normal state")
logger.Info("Looking for WorkspaceTemplateApply", "name", applyName)
logger.Info("Status check",
    "applied", workspaceApply.Status.Applied,
    "synced", syncedCondition,
    "ready", readyCondition)
```

## Migration Guide

1. Existing Deployments
- No manual action required
- Controller will detect and update automatically
- Status will be preserved

2. New Deployments
- Use standard naming scheme
- Controller manages all references
- Full status management from creation
